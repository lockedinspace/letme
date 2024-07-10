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
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if !result {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Validate the config file.",
	Long:  `Validate the config file structure and integrity.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		homeDir := utils.GetHomeDirectory()

		if _, err := os.Stat(homeDir + "/.letme/" + "letme-config"); err == nil {
			result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
			if !result {
				utils.TemplateConfigFile(true)
			}
			fmt.Println("letme: config file valid.")
			os.Exit(0)
		} else {
			utils.CheckAndReturnError(err)
		}
	},
}

func init() {
	ConfigCmd.AddCommand(Validate)
}
