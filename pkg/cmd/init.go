package letme

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Use a cache to improve response times",
	Long: `A cache file will be created on the $HOME directory.
Account names, IDs, and roles to be assumed will be present on the file.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		a, erra := os.UserHomeDir()
		if erra != nil {
			fmt.Println(erra)
			os.Exit(1)
		}
		b := []byte("2013-07-11 17:08:50 mybucket\n2013-07-24 14:55:44 mybucket2")
		c := os.WriteFile(a+"/.letme-cache", b, 0660)
		if c != nil {
			fmt.Println(c)
			os.Exit(1)
		}
		fmt.Println("Cache file stored on " + a + "/.letme-cache")

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
