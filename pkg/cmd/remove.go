package letme

import (
	//"fmt"
	//"github.com/lockedinspace/letme-go/pkg"
	"github.com/spf13/cobra"
	//"os"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an account in your local aws files",
	Long: `Removes the entry referencing the account in the credentials and config files.
Used for sanitazing purposes and keeping readable and clean files.`,
	Run: func(cmd *cobra.Command, args []string) {
  		// var then variable name then variable type
	/* 	var boolean bool
		boolean = false
		var mfaToken string
		// Taking input from user
		if boolean {
			fmt.Println("with mfa")
			mfaToken = "12345"
		} else {
			fmt.Println("without mfa")
			mfaToken = ""
		} */
		
	},
}

func init() {
	//initCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(removeCmd)
}
