package letme

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var version = "0.1"
var rootCmd = &cobra.Command{
	Use:     "letme",
	Version: version,
	Short:   "Obtain AWS credentials from another account",
	Long: `letme will query the DynamoDB table or cache file for the specified account and
load the temporal credentials onto your aws files.
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
