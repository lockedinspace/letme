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
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if result {
		} else {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "List AWS accounts",
	Long: `Lists all of the AWS accounts and their main region
specified in the DynamoDB table or in your cache file.`,
	Run: func(cmd *cobra.Command, args []string) {
		localFlag, _ := cmd.Flags().GetBool("local")
		// grab and save fields from the config file into variables
		profile := utils.ConfigFileResultString("Aws_source_profile")
		region := utils.ConfigFileResultString("Aws_source_profile_region")
		table := utils.ConfigFileResultString("Dynamodb_table")
		// create a new aws session and try to get credentials
		sesAws, err := session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewSharedCredentials("", profile),
		})
		utils.CheckAndReturnError(err)
		_, err = sesAws.Config.Credentials.Get()
		utils.CheckAndReturnError(err)

		// check if the .letme-cache file exists, if not, queries must be satisfied through internet
		if utils.CacheFileExists() {

			// create a struct and a map to iterate over them
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

			// generate a slice which will contain sorted elements
			sorted := make([]string, 0, len(allitems))

			// decode the toml file and append the elements into the sorted array also check if the local flag is passed
			_, err := toml.DecodeFile(utils.GetHomeDirectory()+"/.letme/.letme-cache", &allitems)
			utils.CheckAndReturnError(err)
			if localFlag {
				for _, value := range allitems {
					sorted = append(sorted, value.Name+"\t"+value.Region[0]+"\t"+utils.CheckAccountLocally(value.Name))
				}
			} else {
				for _, value := range allitems {
					sorted = append(sorted, value.Name+"\t"+value.Region[0])
				}
			}
			
			// sort the slice and using a tabwriter print a nicely formed output
			sort.Strings(sorted)
			w := tabwriter.NewWriter(os.Stdout, 35, 200, 1, ' ', 0)
			if localFlag {
				fmt.Fprintln(w, "NAME:\tMAIN REGION:\tLOCAL STATUS (credentials,config):")
				fmt.Fprintln(w, "-----\t------------\t----------------------------------")
			} else {
				fmt.Fprintln(w, "NAME:\tMAIN REGION:")
				fmt.Fprintln(w, "-----\t------------")
			}
			for _, id := range sorted {
				fmt.Fprintln(w, id)
				w.Flush()
			}
		} else {
			// create a new dynamodb session and prepare a query
			sesAwsDB := dynamodb.New(sesAws)
			proj := expression.NamesList(expression.Name("name"), expression.Name("region"))
			expr, err := expression.NewBuilder().WithProjection(proj).Build()
			utils.CheckAndReturnError(err)

			// scan the results from the previous query
			inputs := &dynamodb.ScanInput{
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ProjectionExpression:      expr.Projection(),
				TableName:                 aws.String(table),
			}
			scanTable, err := sesAwsDB.Scan(inputs)
			utils.CheckAndReturnError(err)

			// create a struct to match the data and create a new slice to sort data later
			type account struct {
				Id     int      `json:"id"`
				Name   string   `json:"name"`
				Role   []string `json:"role"`
				Region []string `json:"region"`
			}
			sorted := make([]string, 0, len(scanTable.Items))

			// loop through the results and append the results into the slice
			if localFlag {
				for _, value := range scanTable.Items {
					items := account{}
					err = dynamodbattribute.UnmarshalMap(value, &items)
					utils.CheckAndReturnError(err)
					sorted = append(sorted, items.Name+"\t"+items.Region[0]+"\t"+utils.CheckAccountLocally(items.Name))
				}
			} else {
				for _, value := range scanTable.Items {
				items := account{}
				err = dynamodbattribute.UnmarshalMap(value, &items)
				utils.CheckAndReturnError(err)
				sorted = append(sorted, items.Name+"\t"+items.Region[0])
				}
			}
			

			// sort the slice and using a tabwriter print a nicely formed output
			sort.Strings(sorted)
			w := tabwriter.NewWriter(os.Stdout, 35, 200, 1, ' ', 0)
			if localFlag {
				fmt.Fprintln(w, "NAME:\tMAIN REGION:\tLOCAL STATUS (credentials,config):")
				fmt.Fprintln(w, "-----\t------------\t----------------------------------")
			} else {
				fmt.Fprintln(w, "NAME:\tMAIN REGION:")
				fmt.Fprintln(w, "-----\t------------")
			}
			for _, id := range sorted {
				fmt.Fprintln(w, id)
				w.Flush()
			}
		}
	},
}

func init() {
	var local bool
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&local, "local", "l", false, "lists local accounts and their respective status")
}
