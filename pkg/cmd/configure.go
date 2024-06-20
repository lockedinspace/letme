package letme

import (
	"fmt"
	"os"

	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
	// "golang.org/x/text/cases"
)

var configure = &cobra.Command{
	Use: "configure",
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
	Short: "Interact with letme-config file and manage contexts.",
	Long:  `-`,
	// Args: cobra.MinimumNFlags(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get context flag value
		contexts := utils.GetAvalaibleContexts()
		if list, _ := cmd.Flags().GetBool("list"); list {
			fmt.Println("Available contexts and active context marked with *: ")
			for _, context := range contexts {
				currentContext := utils.GetCurrentContext()
				if context == currentContext {
					fmt.Println("* " + context)
				} else {
					fmt.Println("  " + context)
				}
			}
			os.Exit(0)
		}

		context, _ := cmd.Flags().GetString("context")

		contextExists := false
		for _, section := range contexts {
			if section == context {
				fmt.Println(section, context)
				contextExists = true
			}
		}

		switch {
		case contextExists && len(context) != 0:
			fmt.Println("Context  introduced exists")
			// utils.UpdateContext(context, renew)
		case contextExists && len(context) == 0:
			fmt.Println("Using existing general context")
			// utils.UpdateContext("general", renew)
		case !contextExists && len(context) != 0:
			utils.NewContext(context)
		}
	},
}

func init() {
	rootCmd.AddCommand(configure)
	configure.Flags().StringP("context", "c", "general", "switch the current context")
	configure.Flags().BoolP("list", "l", false, "list active and current context")
	configure.Flags().Bool("renew", false, "force creation of a context even if exists")
}
