package letme

import (
	"fmt"
	"os"
	"path/filepath"

	utils "github.com/hectorruiz-it/letme-alpha/pkg"
	"github.com/spf13/cobra"
)

var configFileCmd = &cobra.Command{
	Use:   "config-file",
	Short: "Creates the letme configuration file",
	Long: `Creates a configuration file with all the needed key pairs.
The config file is created in your '$HOME/.letme-alpha/' directory, letme reads this file
and performs the operations based from the user-specified values.
        `,
	Run: func(cmd *cobra.Command, args []string) {

		// grab and define force flag & verify flags
		forceFlag, _ := cmd.Flags().GetBool("force")
		verifyFlag, _ := cmd.Flags().GetBool("verify")

		// define file name and grab user home directory
		fileName := "letme-config"
		homeDir := utils.GetHomeDirectory()

		if verifyFlag {
			if _, err := os.Stat(homeDir + "/.letme-alpha/" + fileName); err == nil {
				result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme-alpha/letme-config")
				if !result {
					utils.TemplateConfigFile(true)
				}
				fmt.Println("letme: config file is valid!")
				os.Exit(0)
			} else {
				utils.CheckAndReturnError(err)
			}
		}

		// creates the directory + config file or just the config file if the directory already exists
		// then writes the marshalled values on a toml document (letme-config).
		if _, err := os.Stat(homeDir + "/.letme-alpha/"); err != nil {
			err = os.Mkdir(homeDir+"/.letme-alpha/", 0755)
			utils.CheckAndReturnError(err)

			utils.TemplateConfigFile(false)
			err = os.Chmod(filepath.Join(homeDir+"/.letme-alpha/", filepath.Base(fileName)), 0600)
			utils.CheckAndReturnError(err)
			fmt.Println("letme: edit the config file at " + homeDir + "/.letme-alpha/letme-config with your values.")
		} else if _, err := os.Stat(homeDir + "/.letme-alpha/"); err == nil {
			if _, err = os.Stat(homeDir + "/.letme-alpha/" + fileName); err == nil && !(forceFlag) {
				fmt.Println("letme: letme-config file already exists at: " + homeDir + "/.letme-alpha/" + fileName)
				fmt.Println("letme: to restore the letme-config file, pass the -f, --force flags or delete the letme-config file manually.")
				os.Exit(0)
			}
			utils.TemplateConfigFile(false)
			fmt.Println("letme: edit the config file at " + homeDir + "/.letme-alpha/letme-config with your values.")
		}
	},
}

func init() {

	// define a Region boolean variable
	var Force bool
	var Verify bool
	rootCmd.AddCommand(configFileCmd)

	// create a local force flag
	configFileCmd.Flags().BoolVarP(&Force, "force", "f", false, "bypass safety restrictions and force a command to be run")
	configFileCmd.Flags().BoolVarP(&Verify, "verify", "", false, "verify config file structure and integrity")
}
