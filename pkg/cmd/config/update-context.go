package config

import (
	"fmt"
	"os"

	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
)

var UpdateContext = &cobra.Command{
	Use: "update-context",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LetmeConfigCreate()
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if !result {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Updates an existing context.",
	Long:  `Interactively update an existing context configuration on your letme-config file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		contexts := utils.GetAvalaibleContexts()
		letmeContext := args[0]

		for _, section := range contexts {
			if section == letmeContext {
				utils.NewContext(letmeContext)
				fmt.Println("letme: updated letme '" + letmeContext + "' context.")
				os.Exit(0)
			}
		}
		fmt.Println("letme: '" + letmeContext + "' context does not exist in your letme-config file.")
		fmt.Println("letme: run 'letme config new-context " + letmeContext + "' to create it.")
		os.Exit(1)
	},
}

func init() {
	ConfigCmd.AddCommand(UpdateContext)
}
