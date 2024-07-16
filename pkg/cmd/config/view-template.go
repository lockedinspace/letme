package config

import (
	"os"

	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
)

var ViewTemplate = &cobra.Command{
	Use: "view-template",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LetmeConfigCreate()
		utils.ConfigFileHealth()
	},
	Short: "View the a sample configuration file template.",
	Long:  `View the a sample configuration file template, use that as an scaffolding for your needs.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		utils.TemplateConfigFile(true)
		os.Exit(0)
	},
}

func init() {
	ConfigCmd.AddCommand(ViewTemplate)
}
