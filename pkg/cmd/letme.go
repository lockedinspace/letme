package letme

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var version = "0.1.1-rc2"
var rootCmd = &cobra.Command{
	Use:     "letme",
	Version: version,
	Short:   "Obtain aws cli credentials",
	Long: `letme acts like an API between your DynamoDB table and your local computer.
When a request is successful, letme will update your aws files (credentials/config) in order 
to call resources from the requested account.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("letme: try 'letme --help' or 'letme -h' for more information")
		os.Exit(0)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
