package letme

import (
	"github.com/spf13/cobra"
	"errors"
	"github.com/lockedinspace/letme/pkg"
	"os"
	"fmt"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a profile from your local AWS files",
	Long: `Removes the account entry on both of your AWS files.
This will not remove anything on the DynamoDB side. Use it for
cleanup purposes and sanitizing your '$HOME/.aws/credentials' 
and '$HOME/.aws/config' files.
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := utils.AwsCredsFileRead()
		f := utils.AwsConfigFileRead()
		_, errCred := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_APPEND, 0600)
		_, errConf := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_APPEND, 0600)
		if !(errors.Is(errCred, os.ErrNotExist)) && !(errors.Is(errConf, os.ErrNotExist)) && utils.CheckAccountLocally(args[0]) == "true,true" {
			credFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_TRUNC, 0600)
			utils.CheckAndReturnError(err)
			confFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_TRUNC, 0600)
			utils.CheckAndReturnError(err)
			fmt.Fprintf(credFile2, "%v", utils.AwsReplaceBlock(s, args[0]))
			fmt.Fprintf(confFile2, "%v", utils.AwsReplaceBlock(f, args[0]))
			fmt.Println("Removed profile '" + args[0] + "' entries from credentials and config files.")
		} else if utils.CheckAccountLocally(args[0]) == "false,true" {
			confFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_TRUNC, 0600)
			utils.CheckAndReturnError(err)
			fmt.Fprintf(confFile2, "%v", utils.AwsReplaceBlock(f, args[0]))
			fmt.Println("Removed profile '" + args[0] + "' entry from config file.")
		} else if utils.CheckAccountLocally(args[0]) == "true,false" {
			credFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_TRUNC, 0600)
			utils.CheckAndReturnError(err)
			fmt.Fprintf(credFile2, "%v", utils.AwsReplaceBlock(s, args[0]))
			fmt.Println("Removed profile '" + args[0] + "' entry from credentials file.")
		} else {
			fmt.Println("letme: unable to remove profile '" + args[0] + "', not found on your local aws files")
			os.Exit(1)
		}
		
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

}
