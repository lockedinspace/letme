package letme

import (
        "fmt"
        "github.com/lockedinspace/letme-go/pkg"
        "github.com/spf13/cobra"
        "os"
)

var removeCmd = &cobra.Command{
        Use:   "remove",
        Short: "Remove the cache file",
        Long: `Removes the cache file created by the 'init' command.
Deleting this file could cause higher waiting times when retrieving information.`,
        Run: func(cmd *cobra.Command, args []string) {
                // get user home dir and delete the file if exists
                homeDir := utils.GetHomeDirectory()
                if _, err := os.Stat(homeDir + "/.letme/.letme-cache"); err == nil {
                        err := os.Remove(homeDir + "/.letme/.letme-cache")
                        utils.CheckAndReturnError(err)
                        fmt.Println("Cache file successfully removed.")
                } else {
                        fmt.Println("letme: Could not find nor remove cache file.")
                        os.Exit(1)
                }
        },
}

func init() {
        initCmd.AddCommand(removeCmd)
}