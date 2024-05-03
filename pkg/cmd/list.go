package letme

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
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
	Short: "List accounts",
	Long:  `Lists all the AWS accounts and their main region.`,
	Run: func(cmd *cobra.Command, args []string) {

		// get local flag
		localFlag, _ := cmd.Flags().GetBool("local")

		// grab and save fields from the config file into variables
		profile := utils.ConfigFileResultString("general", "Aws_source_profile").(string)
		region := utils.ConfigFileResultString("general", "Aws_source_profile_region").(string)
		table := utils.ConfigFileResultString("general", "Dynamodb_table").(string)

		// create a new aws session
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(letmeContext.AwsSourceProfile), config.WithRegion(letmeContext.AwsSourceProfileRegion))
		utils.CheckAndReturnError(err)
		fmt.Println("Listing accounts using '" + currentContext + "' context:\n")

		utils.GetSortedTable(letmeContext.AwsDynamoDbTable, cfg)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
