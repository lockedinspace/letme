package letme

import (
	"github.com/lockedinspace/letme-go/pkg"
	"fmt"
	"os"
	"bufio"
	"path/filepath"
	"github.com/spf13/cobra"

)

var configFileCmd = &cobra.Command{
	Use:   "config-file",
	Short: "Create a config file to specify required parameters for letme",
	Long: `Use the command 'config-file' to create a toml template with all the key-value pairs needed by letme.
The config file is created on '$HOME/.letme/letme-config', letme needs this file
to perform aws calls. Once created, you will need to edit that file and fill it with 
your values.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		
		fileName := "letme-config"
		homeDir, err := os.UserHomeDir()
		utils.CheckAndReturnError(err)
		if _, err := os.Stat(homeDir + "/.letme/"); err != nil {
			err = os.Mkdir(homeDir + "/.letme/", 0660)
			utils.CheckAndReturnError(err)
			configFile, err := os.Create(filepath.Join(homeDir + "/.letme/", filepath.Base(fileName))) 
			utils.CheckAndReturnError(err)
			defer configFile.Close()
		} else if _, err := os.Stat(homeDir + "/.letme/"); err == nil {
			if _, err = os.Stat(homeDir + "/.letme/" + fileName); err == nil {
				fmt.Println("letme: letme-config file already exists at: " + homeDir + "/.letme/" + fileName)
				fmt.Println("letme: to restore the letme-config file, pass the -f, --force flags or delete the letme-config file manually.")
				os.Exit(0)
			}
			configFile, err := os.Create(filepath.Join(homeDir + "/.letme/", filepath.Base(fileName))) 
			utils.CheckAndReturnError(err)
			defer configFile.Close()
			writer := bufio.NewWriter(configFile)
			_, err = fmt.Fprintf(writer, "%v", utils.TemplateConfigFile())
			utils.CheckAndReturnError(err)
			writer.Flush()
		} 
	},
}

func init() {
	rootCmd.AddCommand(configFileCmd)
}
