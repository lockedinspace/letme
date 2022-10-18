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
	"github.com/lockedinspace/letme-go/pkg"
	"github.com/spf13/cobra"
	"os"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a cache to speed up response times",
	Long: `Creates a cache file in the users $HOME directory.
IDs, account names, roles to be assumed and regions will be present on the cache file. 
This will improve performance because common queries will be satisified by the cache file and will not
be routed to the DynamoDB service from AWS. 

If the end user prefers to satisfy all their queries through internet, they can remove the cache file
with the command 'letme init remove' or just deleting the .letme-cache manually.
        `,
	Run: func(cmd *cobra.Command, args []string) {
		// import a struct to unmarshal the letme-config (toml) document.
		type structUnmarshal = utils.GeneralParams
		type general map[string]structUnmarshal
		var generalConfig general

		// check user home directory and save it into a variable.
		homeDir := utils.GetHomeDirectory()
		configFilePath := homeDir + "/.letme/letme-config"
		if _, err := os.Stat(configFilePath); err != nil {
			fmt.Println("letme: Could not locate any config file. Please run 'letme config-file' to create one.")
			os.Exit(1)
		}

		// once letme-config exists decode it and alert the user for any strange key field which is not present on the struct
		decodedFile, err := toml.DecodeFile(configFilePath, &generalConfig)
		utils.CheckAndReturnError(err)
		oddFields := decodedFile.Undecoded()
		if len(oddFields) == 0 {
		} else {
			fmt.Printf("letme: Unknown key: %q\n", oddFields)
			os.Exit(1)
		}

		// parse toml structure and make unmarshalled variables global
		var exportedProfile string
		var exportedProfileRegion string
		var exportedDynamoDBTable string
		for _, name := range []string{"general"} {
			a := generalConfig[name]
			fmt.Printf("\nProfile: %v\nProfile region: %v\n\n", a.Aws_source_profile, a.Aws_source_profile_region)
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
		cacheFilePath, err := os.Create(homeDir + "/.letme/.letme-cache")
		utils.CheckAndReturnError(err)
		defer cacheFilePath.Close()
		cacheFileWriter := bufio.NewWriter(cacheFilePath)
		// once the query is prepared, scan the table name (specified on letme-config) to retrieve the fields and loop through the results
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
		fmt.Println("Cache file stored on " + homeDir + "/.letme/.letme-cache")

	},
}

func init() {
	var Remove bool
	rootCmd.AddCommand(initCmd)
	configFileCmd.Flags().BoolVarP(&Remove, "remove", "r", false, "remove init file")

}
