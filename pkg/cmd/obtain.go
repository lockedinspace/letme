package letme

import (
	"os"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/lockedinspace/letme-go/pkg"
)

var obtainCmd = &cobra.Command{
	Use:   "obtain",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		homeDir, err := os.UserHomeDir()
		utils.CheckAndReturnError(err)
		if _, err := os.Stat(homeDir + "/.letme/letme-config"); err == nil {
			fmt.Println("Config file found.")
		} else {
			fmt.Println("letme: Could not locate any config file. Please run 'letme config-file' to create one.")
			os.Exit(1)
		}
	},
	Short: "Obtain aws credentials",
	Long: `Through the AWS Security Token Service, obtain temporal credentials
once the user successfully authenticates itself. Credentials will last 3600 seconds
and can be used with the argument '--profile example1' within the aws cli binary.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("obtain called")
	},
}

func init() {
	rootCmd.AddCommand(obtainCmd)
}