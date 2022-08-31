package utils

import (
	"fmt"
	"os"
	"os/exec"
	"bytes"
	"github.com/BurntSushi/toml"	
)

func CommandExists(command string)  {
	_, err := exec.LookPath(command)
	CheckAndReturnError(err)
}
func CheckAndReturnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func CheckConfigExists() string {
	a, err := os.UserHomeDir()
	CheckAndReturnError(err)
	if _, err := os.Stat(a + "/.letme/letme-config"); err != nil {
		fmt.Println("letme: Could not find config file. Please run 'letme config-file'")
	}
	return a + "/.letme/letme-config"

}

func TemplateConfigFile() string {
	var (
		buf     = new(bytes.Buffer)
	)
	err := toml.NewEncoder(buf).Encode(map[string]interface{}{
		"general": map[string]string{
			"aws_source_profile": "",
			"aws_source_profile_region": "",
			"dynamodb_table": "",
			"mfa_arn": "",
		},
	})
	CheckAndReturnError(err)
	return buf.String()
}