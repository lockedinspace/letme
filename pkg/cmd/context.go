package letme

import (
	"fmt"
	"os"
	"github.com/lockedinspace/letme/pkg"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use: "config",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(utils.GetHomeDirectory() + "/.letme/letme-config"); err == nil {
		} else {
			fmt.Println("letme: Could not locate any config file. Please run 'letme config-file' to create one.")
			os.Exit(1)
		}
	},
	Short: "Configure letme options, such as context",
	Long:  `-`,
	Run: func(cmd *cobra.Command, args []string) {
			// Create a map to hold the raw TOML data
	var rawConfig map[string]interface{}

	// Decode the TOML file into the map
	if _, err := toml.DecodeFile(utils.GetHomeDirectory() + "/.letme/letme-config", &rawConfig); err != nil {
		utils.CheckAndReturnError(err)
	}

	// Iterate over the keys (table names) in the rawCon(fig map
	context, _ := cmd.Flags().GetString("context")
	currentContext := utils.GetCurrentContext()
	fmt.Println("Listing all contexts and marking active context with *:")
	for tableName, _  := range rawConfig {
		if tableName == context {
			fmt.Println("* "+tableName)
			utils.UpdateContext(context)
		} else {
			if tableName == currentContext {
				fmt.Println("* "+tableName)
			} else {
				fmt.Println("  "+tableName)
			}
			
		}
		// If you need to access specific fields within each table, you can assert the type
		// if table, ok := tableData.(map[string]interface{}); ok {
		// 	if dynamoDBTable, exists := table["dynamodb_table"]; exists {
		// 		fmt.Printf("  DynamoDB Table: %s\n", dynamoDBTable)
		// 	}
		// }
	}
	},
}
func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.Flags().String("context", "", "switch the current context")
}