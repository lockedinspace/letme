package letme

import (
	"fmt"
	"os"

	utils "github.com/lockedinspace/letme/pkg"

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
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if result {
		} else {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Configure letme options, such as context",
	Long:  `-`,
	Run: func(cmd *cobra.Command, args []string) {

		// Get context flag value
		context, _ := cmd.Flags().GetString("context")

		contexts := utils.GetAvalaibleContexts()

		switch {
		case len(context) > 0: // Detect if prompted context exists & update it
			contextExists := false
			for _, section := range contexts {
				if section == context {
					contextExists = true
				}
			}
			if !contextExists {
				fmt.Println("letme: context '" + context + "' does not exist in your .letme-config file")
				os.Exit(1)
			}
			utils.UpdateContext(context)
			fmt.Println("letme: using context '" + context + "'")

		default:
			fmt.Println("Available contexts and active context marked with *: ")
			currentContext := utils.GetCurrentContext()
			for _, context := range contexts {
				if context == currentContext {
					fmt.Println("* " + context)
				} else {
					fmt.Println("  " + context)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.Flags().String("context", "", "switch the current context")
}

