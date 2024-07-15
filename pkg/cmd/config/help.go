package config

import (
	"fmt"
	"os"

	utils "github.com/lockedinspace/letme/pkg"
	letme "github.com/lockedinspace/letme/pkg/cmd"

	// "github.com/lockedinspace/letme/pkg/cmd/config/context"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use: "config",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LetmeConfigCreate()
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if !result {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Configure letme.",
	Long:  `Personalize your letme experience, manage contexts and more.`,
	// Args: cobra.MinimumNFlags(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	letme.RootCmd.AddCommand(ConfigCmd)
}
