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
	Short: "Interact with letme-config file and manage contexts.",
	Long:  `Interact with letme-config file and manage contexts.`,
	// Args: cobra.MinimumNFlags(1),
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("exit")
	},
}
