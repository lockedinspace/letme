package letme

import (
	"bufio"
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
	"text/tabwriter"
)

var initCmd = &cobra.Command{
	Use: "init",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(utils.GetHomeDirectory() + "/.letme/letme-config"); err == nil {
			utils.CheckAndReturnError(err)
		} else {
			fmt.Println("letme: could not locate any config file. Please run 'letme config-file' to create one.")
			os.Exit(1)
		}
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if result {
		} else {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Creates a file to speed up response times",
	Long: `Creates a cache file in your '$HOME/.letme/' directory.
IDs, account names, roles to be assumed and regions will be present in the cache file. 
This will improve performance as common queries will be firstly answered by the cache file.
        `,
	Run: func(cmd *cobra.Command, args []string) {

		// if remove flag is passed, remove cache file
		removeFlag, _ := cmd.Flags().GetBool("remove")
		if removeFlag {
			if _, err := os.Stat(utils.GetHomeDirectory() + "/.letme/.letme-cache"); err == nil {
				err := os.Remove(utils.GetHomeDirectory() + "/.letme/.letme-cache")
				utils.CheckAndReturnError(err)
				fmt.Println("Cache file successfully removed.")
				os.Exit(0)
			} else {
				fmt.Println("letme: could not find nor remove cache file.")
				os.Exit(1)
			}
		}

		// import a struct to unmarshal the letme-config (toml) document.
		type structUnmarshal = utils.GeneralParams
		type general map[string]structUnmarshal
		var generalConfig general

		// once letme-config exists decode it and alert the user for any strange key field which is not present on the struct
		_, err := toml.DecodeFile(utils.GetHomeDirectory()+"/.letme/letme-config", &generalConfig)
		utils.CheckAndReturnError(err)

		// parse toml structure and make unmarshalled variables global
		var exportedProfile string
		var exportedProfileRegion string
		var exportedDynamoDBTable string
		w := tabwriter.NewWriter(os.Stdout, 25, 200, 10, ' ', 0)
		fmt.Println("Creating cache file with the following specs:\n")
		fmt.Fprintln(w, "PROFILE:\tPROFILE REGION:\tDYNAMODB TABLE:")
		fmt.Fprintln(w, "--------\t---------------\t---------------")
		for _, name := range []string{"general"} {
			a := generalConfig[name]
			fmt.Fprintln(w, a.Aws_source_profile+"\t"+a.Aws_source_profile_region+"\t"+a.Dynamodb_table+"\n")
			w.Flush()
			exportedProfile = a.Aws_source_profile
			exportedProfileRegion = a.Aws_source_profile_region
			exportedDynamoDBTable = a.Dynamodb_table
		}

		// create a struct to hold the data that will be passed into .letme-cache file
		type account struct {
			Id     int      `json:"id"`
			Name   string   `json:"name"`
			Role   []string `json:"role"`
			Region []string `json:"region"`
		}

		// create a service connection to aws with the profile/region specified on letme-config
		sesAws, err := session.NewSession(&aws.Config{
			Region:      aws.String(exportedProfileRegion),
			Credentials: credentials.NewSharedCredentials("", exportedProfile),
		})
		utils.CheckAndReturnError(err)
		_, err = sesAws.Config.Credentials.Get()
		utils.CheckAndReturnError(err)

		// prepare a dynamodb query (projection + expression)
		dynamoDBTable := exportedDynamoDBTable
		sesAwsDB := dynamodb.New(sesAws)
		proj := expression.NamesList(expression.Name("id"), expression.Name("name"), expression.Name("role"), expression.Name("region"))
		expr, err := expression.NewBuilder().WithProjection(proj).Build()
		utils.CheckAndReturnError(err)
		inputs := &dynamodb.ScanInput{
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
			TableName:                 aws.String(dynamoDBTable),
		}
		cacheFilePath, err := os.Create(utils.GetHomeDirectory() + "/.letme/.letme-cache")
		utils.CheckAndReturnError(err)
		defer cacheFilePath.Close()
		cacheFileWriter := bufio.NewWriter(cacheFilePath)

		// once the query is prepared, scan the table name and retrieve the fields
		scanTable, err := sesAwsDB.Scan(inputs)
		utils.CheckAndReturnError(err)
		for _, i := range scanTable.Items {
			item := account{}
			err = dynamodbattribute.UnmarshalMap(i, &item)
			utils.CheckAndReturnError(err)

			// save the exported variables into a file (.letme-cache) this will improve performance because common queries will be satisified by the cache file
			_, err = fmt.Fprintf(cacheFileWriter, "%v", utils.TemplateCacheFile(item.Name, item.Id, item.Role, item.Region))
			utils.CheckAndReturnError(err)
			cacheFileWriter.Flush()
		}
		err = os.Chmod(utils.GetHomeDirectory()+"/.letme/.letme-cache", 0600)
		utils.CheckAndReturnError(err)
		fmt.Println("Cache file stored on " + utils.GetHomeDirectory() + "/.letme/.letme-cache")

	},
}

func init() {
	var Remove bool
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&Remove, "remove", "", false, "remove init file")

}
