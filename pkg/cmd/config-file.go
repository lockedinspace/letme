package letme

import (
        "bufio"
        "fmt"
        "github.com/lockedinspace/letme-go/pkg"
        "github.com/spf13/cobra"
        "os"
        "path/filepath"
)

var configFileCmd = &cobra.Command{
        Use:   "config-file",
        Short: "Create a config file needed by letme where parameters such as MFA arn are stored.",
        Long: `Use the command 'config-file' to create a toml template with all the key-value pairs needed by letme.
The config file is created on '$HOME/.letme/letme-config', letme needs this file
to perform aws calls. Once created, you will need to edit that file and fill it with 
your values.
        `,
        Run: func(cmd *cobra.Command, args []string) {
                // grab and define force flag
                forceFlag, _ := cmd.Flags().GetBool("force")

                // define file name and grab user home directory
                fileName := "letme-config"
                homeDir, err := os.UserHomeDir()
                utils.CheckAndReturnError(err)

                // conditional statement which creats either the directory + config file or just the config file if the directory already exists
                // then writes  the marshalled values on a toml document (letme-config).
                if _, err := os.Stat(homeDir + "/.letme/"); err != nil {
                        err = os.Mkdir(homeDir+"/.letme/", 0700)
                        utils.CheckAndReturnError(err)

                        configFile, err := os.Create(filepath.Join(homeDir+"/.letme/", filepath.Base(fileName)))
                        utils.CheckAndReturnError(err)
                        defer configFile.Close()

                        writer := bufio.NewWriter(configFile)
                        _, err = fmt.Fprintf(writer, "%v", utils.TemplateConfigFile())
                        utils.CheckAndReturnError(err)
                        writer.Flush()
                        fmt.Println("letme: edit the config file at " + homeDir + "/.letme/letme-config with your values.")
                } else if _, err := os.Stat(homeDir + "/.letme/"); err == nil {
                        if _, err = os.Stat(homeDir + "/.letme/" + fileName); err == nil && !(forceFlag) {
                                fmt.Println("letme: letme-config file already exists at: " + homeDir + "/.letme/" + fileName)
                                fmt.Println("letme: to restore the letme-config file, pass the -f, --force flags or delete the letme-config file manually.")
                                os.Exit(0)
                        }
                        configFile, err := os.Create(filepath.Join(homeDir+"/.letme/", filepath.Base(fileName)))
                        utils.CheckAndReturnError(err)
                        defer configFile.Close()

                        writer := bufio.NewWriter(configFile)
                        _, err = fmt.Fprintf(writer, "%v", utils.TemplateConfigFile())
                        utils.CheckAndReturnError(err)
                        writer.Flush()
                        fmt.Println("letme: edit the config file at " + homeDir + "/.letme/letme-config with your values.")
                }
        },
}

func init() {
        // define a Region boolean variable
        var Region bool

        rootCmd.AddCommand(configFileCmd)

        // create a local force flag
        configFileCmd.Flags().BoolVarP(&Region, "force", "f", false, "bypass safety restrictions and force a command to be run")
}