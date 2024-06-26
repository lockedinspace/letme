package context

import (
	"fmt"
	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
	"os"
)

var ContextCmd = &cobra.Command{
	Use: "context",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.LetmeConfigCreate()
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
		if !result {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Manage contexts",
	Long:  `Manage contexts inside your letme-config file`,
	Run: func(cmd *cobra.Command, args []string) {
		contexts := utils.GetAvalaibleContexts()

		// if no flags passed, list contexts
		if cmd.Flags().NFlag() == 0 {
			fmt.Println("Active context marked with '*': ")
			for _, context := range contexts {
				currentContext := utils.GetCurrentContext()
				if context == currentContext {
					fmt.Println("* " + context)
				} else {
					fmt.Println("  " + context)
				}
			}
		} else if cmd.Flags().Changed("switch") {
			contextToSwitch, _ := cmd.Flags().GetString("switch")
			for _, section := range contexts {
				if section == contextToSwitch {
					utils.UpdateContext(contextToSwitch)
					fmt.Println("Using: '" + contextToSwitch + "' context.")
				} else {
					fmt.Println("letme: '" + contextToSwitch + "' context does not exist in your letme-config file.")
					fmt.Println("run 'letme configuration context --create " + contextToSwitch  + "' to create it.")
					os.Exit(1)
				}
			} 
		} else if cmd.Flags().Changed("create") {
			contextToCreate, _ := cmd.Flags().GetString("create")
			utils.NewContext(contextToCreate)
		}
		
	},
}
func init() {
	ContextCmd.Flags().StringP("switch", "s", "", "switch to the specified context")
	ContextCmd.Flags().StringP("create", "c", "", "create the specified context")
}