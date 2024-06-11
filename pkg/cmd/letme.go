package letme

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/google/go-github/v48/github"
	"github.com/hashicorp/go-version"
	utils "github.com/hectorruiz-it/letme-alpha/pkg"
	"github.com/spf13/cobra"
)

var currentVersion = "v0.2.0-rc4"
var versionPrettyName = "Nuclear dawn"
var rootCmd = &cobra.Command{
	Use:   "letme",
	Short: "Obtain AWS credentials from another account",
	Long: `letme automates the process of authenticating and obtaining credentials
from another AWS account.
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
	fmt.Println("letme " + currentVersion + " (" + versionPrettyName + ") for (" + runtime.GOOS + "/" + runtime.GOARCH + ")")

	client := github.NewClient(nil)
	tags, _, err := client.Repositories.ListReleases(context.Background(), "lockedinspace", "letme", nil)
	utils.CheckAndReturnError(err)
	v1, err := version.NewVersion(currentVersion)
	utils.CheckAndReturnError(err)
	if len(tags) > 0 {
		latestTag := tags[0]
		v2, err := version.NewVersion(*latestTag.Name)
		utils.CheckAndReturnError(err)
		if v1.LessThan(v2) {
			fmt.Printf("\n%s is not longer the latest version. Please consider updating to: %s\n", v1, v2)
		} else {
			fmt.Println("\nYou are using the latest version available.")
		}
	} else {
		fmt.Printf("No tags yet\n")
	}
	fmt.Println("More info: https://github.com/lockedinspace/letme")
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
