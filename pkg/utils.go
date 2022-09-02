package utils

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"os/exec"
	"reflect"
)

// struct to unmarshall toml (wiill be modified as new options are demanded)
type GeneralParams struct {
	Aws_source_profile        string
	Aws_source_profile_region string `toml:"aws_source_profile_region,omitempty"`
	Dynamodb_table            string
	Mfa_arn                   string `toml:"mfa_arn,omitempty"`
}

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
			"aws_source_profile":        "",
			"aws_source_profile_region": "",
			"dynamodb_table":            "",
			"mfa_arn":                   "",
		},
	})
	CheckAndReturnError(err)
	return buf.String()
}

// this function returns the caller $HOME directory
func GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	CheckAndReturnError(err)
	return homeDir
}

// this function parsers the struct and returns one field (string only) at a time
func ConfigFileResultString(field string) string {
	type structUnmarshall = GeneralParams
	type general map[string]structUnmarshall
	var generalConfig general
	_, err := toml.DecodeFile(GetHomeDirectory()+"/.letme/letme-config", &generalConfig)
	CheckAndReturnError(err)
	var exportedField string
	for _, name := range []string{"general"} {
		a := generalConfig[name]
		r := reflect.ValueOf(a)
		f := reflect.Indirect(r).FieldByName(field)
		exportedField = string(f.String())

	}
	return exportedField
}

// TODO: function which validates a toml file against a struct
