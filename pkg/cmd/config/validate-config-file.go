package config

import (
	"fmt"
	"os"

	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
)

var Validate = &cobra.Command{
	Use: "validate",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LetmeConfigCreate()
		utils.ConfigFileHealth()
	},
	Short: "Validates the config file structure and context parameters.",
	Long:  `Validates if the config file has the right structure. Pass the --context flag if you want to validate the context endpoint reachability.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		context, err := cmd.Flags().GetString("context")
		homeDir := utils.GetHomeDirectory()
		if _, err := os.Stat(homeDir + "/.letme/" + "letme-config"); err == nil {
			result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
			if !result {
				utils.TemplateConfigFile(true)
			}
			fmt.Println("letme: the letme config file structure is valid.")
			if context {
				fmt.Print("context endpoint validation called")
			}
			os.Exit(0)
		} else {
			utils.CheckAndReturnError(err)
		}
	},
}

func init() {
	ConfigCmd.AddCommand(Validate)
	ConfigCmd.Flags().StringP("context", "", "", "validates the context endpoint's reachability")

}
