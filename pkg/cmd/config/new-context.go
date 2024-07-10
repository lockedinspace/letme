package config

import (
	"fmt"
	"os"

	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
)

var NewContext = &cobra.Command{
	Use: "new-context",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LetmeConfigCreate()
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if !result {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Create a new context.",
	Long:  `Interactively creates a new context configuration on your letme-config file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		contexts := utils.GetAvalaibleContexts()
		letmeContext := args[0]

		for _, section := range contexts {
			if section == letmeContext {
				fmt.Println("letme: context '" + letmeContext + "' already exists. Modify it with 'letme update-context " + letmeContext + "'.")
				os.Exit(1)
			}
		}
		utils.NewContext(letmeContext)
		fmt.Println("Created letme '" + letmeContext + "' context.")
		os.Exit(0)
	},
}

func init() {
	ConfigCmd.AddCommand(NewContext)
}
