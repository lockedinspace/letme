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
		utils.ConfigFileHealth()
	},
	Short: "Create a new context.",
	Long:  `Interactively creates a new context in your letme-config file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		contexts := utils.GetAvalaibleContexts()
		letmeContext := args[0]

		for _, section := range contexts {
			if section == letmeContext {
				fmt.Println("letme: context '" + letmeContext + "' already exists. Update it with 'letme update-context " + letmeContext + "'.")
				os.Exit(1)
			}
		}
		utils.NewContext(letmeContext)
		fmt.Println("Created letme '" + letmeContext + "' context.")
	},
}

func init() {
	ConfigCmd.AddCommand(NewContext)
}
