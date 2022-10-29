package letme

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"text/tabwriter"
)

var listCmd = &cobra.Command{
	Use: "list",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(utils.GetHomeDirectory() + "/.letme/letme-config"); err == nil {
		} else {
			fmt.Println("letme: Could not locate any config file. Please run 'letme config-file' to create one.")
			os.Exit(1)
		}
	},
	Short: "List AWS accounts",
	Long: `Lists all of the AWS accounts and their main region
specified in the DynamoDB table or in your cache file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// grab fields from config file
		profile := utils.ConfigFileResultString("Aws_source_profile")
		region := utils.ConfigFileResultString("Aws_source_profile_region")
		table := utils.ConfigFileResultString("Dynamodb_table")
		sesAws, err := session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewSharedCredentials("", profile),
		})
		utils.CheckAndReturnError(err)
		_, err = sesAws.Config.Credentials.Get()
		utils.CheckAndReturnError(err)
		// if cache file exists, work on that file, either go all against dynamodb service
		if utils.CacheFileExists() {
			type (
				accountFields struct {
					Id     int      `toml:"id"`
					Name   string   `toml:"name"`
					Region []string `toml:"region"`
					Role   []string `toml:"role"`
				}
				general map[string]accountFields
			)
			
			var allitems general
			sorted := make([]string, 0, len(allitems))
			_, err := toml.DecodeFile(utils.GetHomeDirectory()+"/.letme/.letme-cache", &allitems)
			utils.CheckAndReturnError(err)  
			
			 
			for _, value := range allitems {
				sorted = append(sorted, value.Name + "\t" + value.Region[0] )

			}
			
			sort.Strings(sorted)
			fmt.Println(sorted)
			w := tabwriter.NewWriter(os.Stdout, 25, 200, 1, ' ', 0) 
			fmt.Fprintln(w, "NAME:\tMAIN REGION:") 
			fmt.Fprintln(w, "-----\t------------")
			for _, id := range sorted {
				fmt.Fprintln(w, id)
				w.Flush()

			}
		} else {
			sesAwsDB := dynamodb.New(sesAws)
			proj := expression.NamesList(expression.Name("name"), expression.Name("region"))
			expr, err := expression.NewBuilder().WithProjection(proj).Build()
			utils.CheckAndReturnError(err)
			inputs := &dynamodb.ScanInput{
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ProjectionExpression:      expr.Projection(),
				TableName:                 aws.String(table),
			}
			// once the query is prepared, scan the table name (specified on letme-config) to retrieve the fields and loop through the results
			scanTable, err := sesAwsDB.Scan(inputs)
			utils.CheckAndReturnError(err)
			type account struct {
				Id     int      `json:"id"`
				Name   string   `json:"name"`
				Role   []string `json:"role"`
				Region []string `json:"region"`
			}
			sorted := make([]string, 0, len(scanTable.Items))
			for _, value := range scanTable.Items {
				items := account{}
				err = dynamodbattribute.UnmarshalMap(value, &items)
				utils.CheckAndReturnError(err)
				sorted = append(sorted, items.Name + "\t" + items.Region[0] )

			}
			sort.Strings(sorted)
			fmt.Println(sorted)
		 	w := tabwriter.NewWriter(os.Stdout, 25, 200, 1, ' ', 0) 
			fmt.Fprintln(w, "NAME:\tMAIN REGION:") 
			fmt.Fprintln(w, "-----\t------------")
			for _, id := range sorted {
				fmt.Fprintln(w, id)
				w.Flush()

			} 
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
