package utils

import (
	"os"
	"fmt"
	"bytes"
	"os/exec"
	"github.com/BurntSushi/toml"
)
// this function checks if a command exists
func CommandExists(command string) {
	_, err := exec.LookPath(command)
	CheckAndReturnError(err)
}
// this function checks the error, if the error contains a message, stop the execution and show the error to the user
func CheckAndReturnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
// marshall data into a toml file
func TemplateConfigFile() string {
	var (
		buf = new(bytes.Buffer)
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
