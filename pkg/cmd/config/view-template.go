package config

import (
	"fmt"
	"os"

	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
)

var ViewTemplate = &cobra.Command{
	Use: "view-template",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LetmeConfigCreate()
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if !result {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Validate config file structure and integrity ",
	Long:  `validate config file structure and integrity of your letme-config file.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		utils.TemplateConfigFile(true)
		os.Exit(0)
	},
}

func init() {
	ConfigCmd.AddCommand(ViewTemplate)
}
