package configure

import (
	"fmt"
	utils "github.com/lockedinspace/letme/pkg"
	"github.com/lockedinspace/letme/pkg/cmd"
	"github.com/lockedinspace/letme/pkg/cmd/configure/context"
	"github.com/spf13/cobra"
	"os"
)

var ConfigureCmd = &cobra.Command{
	Use: "configure",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LetmeConfigCreate()
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if !result {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Interact with letme-config file and manage contexts.",
	Long:  `Interact with letme-config file and manage contexts.`,
	// Args: cobra.MinimumNFlags(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get context flag value
		contexts := utils.GetAvalaibleContexts()
		if list, _ := cmd.Flags().GetBool("template"); list {
			utils.TemplateConfigFile(true)
			os.Exit(0)
		}

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

		homeDir := utils.GetHomeDirectory()

		if validateFlag, _ := cmd.Flags().GetBool("validate-config"); validateFlag {
			if _, err := os.Stat(homeDir + "/.letme/" + "letme-config"); err == nil {
				result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
				if !result {
					utils.TemplateConfigFile(true)
				}
				fmt.Println("letme: config file is valid!")
				os.Exit(0)
			} else {
				utils.CheckAndReturnError(err)
			}
		}

		context, _ := cmd.Flags().GetString("context")
		renew, _ := cmd.Flags().GetBool("renew")

		contextExists := false
		for _, section := range contexts {
			if section == context {
				contextExists = true
			}
		}

		switch {
		case contextExists && !renew:
			utils.UpdateContext(context)
			fmt.Println("letme: DSswitched to letme '" + context + "' context.")
		case !contextExists || renew:
			utils.NewContext(context)
			fmt.Println("letme: created letme '" + context + "' context.")
		}
	},
}

func init() {
	letme.RootCmd.AddCommand(ConfigureCmd)
	ConfigureCmd.AddCommand(context.ContextCmd)
	ConfigureCmd.Flags().StringP("context", "c", "general", "switch to the specified context or creates it")
	ConfigureCmd.Flags().BoolP("list", "l", false, "list active and current context")
	ConfigureCmd.Flags().Bool("renew", false, "edits an existing context if exists")
	ConfigureCmd.Flags().Bool("template", false, "shows a letme-config context example")
	ConfigureCmd.Flags().Bool("validate-config", false, "validate config file structure and integrity")
}
