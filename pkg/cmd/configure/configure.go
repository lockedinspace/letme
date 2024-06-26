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
		if list, _ := cmd.Flags().GetBool("template"); list {
			utils.TemplateConfigFile(true)
		}

		
		homeDir := utils.GetHomeDirectory()

		if validateFlag, _ := cmd.Flags().GetBool("validate-config"); validateFlag {
			if _, err := os.Stat(homeDir + "/.letme/" + "letme-config"); err == nil {
				result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
				if !result {
					utils.TemplateConfigFile(true)
				}
				fmt.Println(utils.GetHomeDirectory() + "/.letme/letme-config is valid")
			} else {
				utils.CheckAndReturnError(err)
			}
		}	
	},
}

func init() {
	letme.RootCmd.AddCommand(ConfigureCmd)
	ConfigureCmd.AddCommand(context.ContextCmd)
	ConfigureCmd.Flags().Bool("template", false, "shows a letme-config context example")
	ConfigureCmd.Flags().Bool("validate-config", false, "validate config file structure and integrity")
}
