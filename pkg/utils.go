package utils

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
)

// struct to unmarshal toml (will be modified as new options are demanded)
type GeneralParams struct {
	Aws_source_profile        string
	Aws_source_profile_region string `toml:"aws_source_profile_region,omitempty"`
	Dynamodb_table            string
	Mfa_arn                   string `toml:"mfa_arn,omitempty"`
}

// struct to parse cache data
type CacheFields struct {
	Id     int      `toml:"id"`
	Name   string   `toml:"name"`
	Role   []string `toml:"role"`
	Region []string `toml:"region"`
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

// this function marshalls data into a toml file (letme-config)
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

// this function marshalls data into a toml file (.letme-cache)
func TemplateCacheFile(accountName string, accountID int, accountRole []string, accountRegion []string) string {
	var (
		buf = new(bytes.Buffer)
	)
	err := toml.NewEncoder(buf).Encode(map[string]interface{}{
		accountName: map[string]interface{}{
			"id":     accountID,
			"name":   accountName,
			"role":   accountRole,
			"region": accountRegion,
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

// this function checks if a cache file exists
func CacheFileExists() bool {
	if _, err := os.Stat(GetHomeDirectory() + "/.letme/.letme-cache"); err == nil {
		return true
	} else {
		return false
	}
}

// this function reads the cache file
func CacheFileRead() string {
	readCacheFile, err := ioutil.ReadFile(GetHomeDirectory() + "/.letme/.letme-cache")
	CheckAndReturnError(err)
	s := string(readCacheFile)
	return s
}

// this function maps data on the cache file into a struct
func ParseCacheFile(account string) CacheFields {
	type o = CacheFields
	type general map[string]o
	var generalConfig general
	homeDir := GetHomeDirectory()
	configFilePath := homeDir + "/.letme/.letme-cache"
	_, err := toml.DecodeFile(configFilePath, &generalConfig)
	CheckAndReturnError(err)
	s := generalConfig[account]
	return s
	/* for _, name := range []string{account} {
		s := generalConfig[name]
		fmt.Printf(s.Name)
	} */

}

// TODO: function which validates a toml file against a struct
