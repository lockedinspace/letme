package letme

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

var version = "v0.1.5"
var rootCmd = &cobra.Command{
	Use:   "letme",
	Short: "Obtain AWS credentials from another account",
	Long: `letme will query the DynamoDB table or cache file for the specified account and
load the temporal credentials onto your aws files.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			getVersions()
			os.Exit(0)
		}
		fmt.Println("letme: try 'letme --help' or 'letme -h' for more information")
		os.Exit(0)
	},
}

func getVersions() string {
	fmt.Println("letme "+version+" ("+runtime.GOOS+"/"+runtime.GOARCH+")")
	return " "
}
func init() {
	var Version bool
	rootCmd.PersistentFlags().BoolVarP(&Version, "version", "v", false, "list current version for letme")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
