package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Struct which represents the config-file keys
type GeneralParams struct {
	Aws_source_profile        string
	Aws_source_profile_region string `toml:"aws_source_profile_region,omitempty"`
	Dynamodb_table            string
	Mfa_arn                   string `toml:"mfa_arn,omitempty"`
	Session_name              string
	Session_duration          int64 `toml:"session_duration,omitempty"`
}

// Expected keys in letme-config file
var ExpectedKeys = map[string]bool{
	"aws_source_profile":       true,
	"aws_source_profile_region": true,
	"dynamodb_table":           true,
	"mfa_arn":                  true,
	"session_name":             true,
	"session_duration":         true,
}

// Mandatory keys in letme-config file
var MandatoryKeys = []string{
	"aws_source_profile",
	"aws_source_profile_region",
	"dynamodb_table",
}

// Verify if the config-file respects the struct GeneralParams
func CheckConfigFile(path string) bool {

	type Config map[string]map[string]interface{}
	var config Config

	// Decode the TOML file into the config map
	md, err := toml.DecodeFile(path, &config)
	if err != nil {
		fmt.Println("letme: error decoding TOML file:", err)
		return false
	}

	// Iterate over each table and validate its keys
	for tableName, table := range config {
		// Check for mandatory keys
		for _, key := range MandatoryKeys {
			if _, ok := table[key]; !ok {
				fmt.Printf("letme: missing mandatory key '%s' in table '%s'\n", key, tableName)
				return false
			}
		}
		for key := range table {
			if !ExpectedKeys[key] {
				fmt.Printf("Error: Invalid key '%s' in table '%s'\n", key, tableName)
				return false
			}
		}
	}

	// Check for any undecoded keys
	undecoded := md.Undecoded()
	if len(undecoded) > 0 {
		fmt.Println("Undecoded keys found:")
		for _, key := range undecoded {
			fmt.Println(key)
		}
		return false
	}
	return true
}

// Check if a command exists
func CommandExists(command string) {
	_, err := exec.LookPath(command)
	CheckAndReturnError(err)
}

// Checks the error, if the error contains a message, stop the execution and show the error to the user
func CheckAndReturnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Marshalls data into a toml file (config-file)
func TemplateConfigFile() string {
	var (
		buf = new(bytes.Buffer)
	)
	err := toml.NewEncoder(buf).Encode(map[string]interface{}{
		"contextName": map[string]interface{}{
			"aws_source_profile":        "default",
			"aws_source_profile_region": "eu-west-3",
			"dynamodb_table":            "customers",
			"mfa_arn":                   "arn:aws:iam::4002019901:mfa/user",
			"session_name":              "user_letme",
			"session_duration":          3600,
		},
	})
	CheckAndReturnError(err)
	return buf.String()
}

// Gets the user $HOME directory
func GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	CheckAndReturnError(err)
	return homeDir
}

// Parses letme-config file and returns one field at a time
func ConfigFileResultString(profile string, field string) interface{} {
	var generalConfig map[string]GeneralParams
	_, err := toml.DecodeFile(GetHomeDirectory()+"/.letme/letme-config", &generalConfig)
	CheckAndReturnError(err)
	switch field {
	case "Aws_source_profile":
		return generalConfig[profile].Aws_source_profile
	case "Aws_source_profile_region":
		return generalConfig[profile].Aws_source_profile_region
	case "Mfa_arn":
		return generalConfig[profile].Mfa_arn
	case "Session_name":
		return generalConfig[profile].Session_name
	case "Dynamodb_table":
		return generalConfig[profile].Dynamodb_table
	case "Session_duration":
		return generalConfig[profile].Session_duration
	default:
		fmt.Println("letme: error while retrieving field '" + field + "' could not be found in " + GetHomeDirectory() + "/.letme/letme-config")
		os.Exit(1)
	}
	return generalConfig[profile]
}

// Checks if the .letme-cache file exists, this file is not supported starting from versions 0.2.0 and above
func CacheFileExists() bool {
	if _, err := os.Stat(GetHomeDirectory() + "/.letme/.letme-cache"); err == nil {
		return true
	} else {
		return false
	}
}

// Reads the aws credentials file
func AwsCredsFileRead() string {
	readCacheFile, err := ioutil.ReadFile(GetHomeDirectory() + "/.aws/credentials")
	CheckAndReturnError(err)
	s := string(readCacheFile)
	return s
}

// Reads the aws config file
func AwsConfigFileRead() string {
	readCacheFile, err := ioutil.ReadFile(GetHomeDirectory() + "/.aws/config")
	CheckAndReturnError(err)
	s := string(readCacheFile)
	return s
}

// Marshalls data into a string used for the aws credentials file
func AwsCredentialsFile(accountName string, accessKeyID string, secretAccessKey string, sessionToken string) string {
	now := time.Now()
	a := now.Format("Jan 2, 2006 15:04:05")
	return fmt.Sprintf(
		`#s-%v
#%v;t
[%v]
aws_access_key_id = %v
aws_secret_access_key = %v
aws_session_token = %v
#e-%v
`, accountName, a, accountName, accessKeyID, secretAccessKey, sessionToken, accountName)
}

// Marshalls data into a string used for the aws config file but with the v1 output protocol
func AwsConfigFileCredentialsProcessV1(accountName string, region string) string {
	return fmt.Sprintf(
		`#s-%v
[profile %v]
credential_process = letme obtain %v --v1
region = %v
output = json
#e-%v
`, accountName, accountName, accountName, region, accountName)
}

// Marshalls data into a string used for the aws credentials file
func AwsConfigFile(accountName string, region string) string {
	return fmt.Sprintf(
		`#s-%v
[profile %v]
region = %v
output = json
#e-%v
`, accountName, accountName, region, accountName)
}

// Removes from a file all text in between two indentificators (accountName)
func AwsReplaceBlock(file string, accountName string) string {
	str := "#s-" + accountName
	etr := "#e-" + accountName
	empty := ""
	if strings.Contains(file, str) && strings.Contains(file, etr) {
		startIndex := strings.Index(file, str)
		stopIndex := strings.Index(file, etr) + len(etr)
		res := file[:startIndex] + file[stopIndex:]
		res = strings.ReplaceAll(res, "\n\n", "\n")
		return res
	}
	return empty
}

// Returns only the text entry which statisfies the accountName
func AwsSingleReplaceBlock(file string, accountName string) string {
	str := "#s-" + accountName
	etr := "#e-" + accountName
	empty := ""
	if strings.Contains(file, str) && strings.Contains(file, etr) {
		startIndex := strings.Index(file, str)
		stopIndex := strings.Index(file, etr) + len(etr)
		res := file[startIndex:stopIndex]
		res = strings.ReplaceAll(res, "\n\n", "\n")
		return res
	}
	return empty
}

// Return the latest requested time from a block of text
func GetLatestRequestedTime(content string) string {
	pat := regexp.MustCompile(`#.*\;t`)
	s := pat.FindString(content)
	out := strings.TrimLeft(strings.TrimRight(s, ";t"), "#")
	return out
}

// Check if an account is present on the local aws credentials/config files
func CheckAccountLocally(account string) string {
	accountCredExists, err := regexp.MatchString("(?sm)#s-"+account+"$.*?#e-"+account+"$", AwsCredsFileRead())
	CheckAndReturnError(err)
	accountConfExists, err := regexp.MatchString("(?sm)#s-"+account+"$.*?#e-"+account+"$", AwsConfigFileRead())
	CheckAndReturnError(err)
	if accountCredExists && accountConfExists {
		return fmt.Sprintf("%t,%t", accountCredExists, accountConfExists)
	} else if !(accountCredExists) && accountConfExists {
		return fmt.Sprintf("%t,%t", accountCredExists, accountConfExists)
	} else if accountCredExists && !(accountConfExists) {
		return fmt.Sprintf("%t,%t", accountCredExists, accountConfExists)
	}
	return ""
}

// Struct which states the credential process output for the v1 protocol
type CredentialsProcess struct {
	Version         int
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// Return aws credentials following the credentials_process standard
// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html
func CredentialsProcessOutput(accessKeyID string, secretAccessKey string, sessionToken string, expirationTime time.Time) string {
	group := CredentialsProcess{
		Version:         1,
		AccessKeyId:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
		Expiration:      expirationTime,
	}
	b, err := json.Marshal(group)
	CheckAndReturnError(err)
	return string(b)
}

type Dataset struct {
	Name          string `json:"name"`
	LastRequest   int64  `json:"lastRequest"`
	Expiry        int64  `json:"expiry"`
	AuthMethod    string `json:"authMethod"`
	V1Credentials string `json:"v1Credentials,omitempty"`
}
type Account struct {
	Account Dataset `json:"account"`
}

// Create a file which stores the last time when credentials where requested. Then query if the account exists,
// if not, it will create its first entry.
func DatabaseFile(accountName string, sessionDuration int64, v1Credentials string, authMethod string) {
	databaseFileWriter, err := os.OpenFile(GetHomeDirectory()+"/.letme/.letme-db", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	CheckAndReturnError(err)
	databaseFileReader, err := os.ReadFile(GetHomeDirectory() + "/.letme/.letme-db")
	CheckAndReturnError(err)
	fi, err := os.Stat(GetHomeDirectory() + "/.letme/.letme-db")
	CheckAndReturnError(err)
	var idents []Account
	if fi.Size() > 0 {
		//check if the json is valid, but ensure that the file has content
		if !json.Valid([]byte(databaseFileReader)) && fi.Size() > 0 {
			fmt.Printf("letme: " + GetHomeDirectory() + "/.letme/.letme-db" + " is not JSON valid. Remove the file and try again.\n")
			os.Exit(1)
		}
		err = json.Unmarshal(databaseFileReader, &idents)
		CheckAndReturnError(err)
		err = os.Truncate(GetHomeDirectory()+"/.letme/.letme-db", 0)
		CheckAndReturnError(err)
		for i := range idents {
			//when file is populated and client exist, just update fields
			if idents[i].Account.Name == accountName {
				idents[i].Account.LastRequest = time.Now().Unix()
				idents[i].Account.Expiry = time.Now().Add(time.Second * time.Duration(sessionDuration)).Unix()
				idents[i].Account.V1Credentials = v1Credentials
				idents[i].Account.AuthMethod = authMethod
				b, err := json.MarshalIndent(idents, "", "  ")
				CheckAndReturnError(err)

				if _, err = databaseFileWriter.WriteString(string(b)); err != nil {
					CheckAndReturnError(err)
					defer databaseFileWriter.Close()
				}
				return
			}
		}
		//when file is populated but client does not exist
		idents = append(idents, Account{Dataset{accountName, time.Now().Unix(), time.Now().Add(time.Second * time.Duration(sessionDuration)).Unix(), authMethod, v1Credentials}})
		b, err := json.MarshalIndent(idents, "", "  ")
		CheckAndReturnError(err)

		if _, err = databaseFileWriter.WriteString(string(b)); err != nil {
			CheckAndReturnError(err)
			defer databaseFileWriter.Close()
		}
		//when file does not exist neither the client
	} else if fi.Size() == 0 {
		idents = append(idents, Account{Dataset{accountName, time.Now().Unix(), time.Now().Add(time.Second * time.Duration(sessionDuration)).Unix(), authMethod, v1Credentials}})
		b, err := json.MarshalIndent(idents, "", "  ")
		CheckAndReturnError(err)

		if _, err = databaseFileWriter.WriteString(string(b)); err != nil {
			CheckAndReturnError(err)
			defer databaseFileWriter.Close()
		}
	}
}

// Compare the current local time with the expiry field in the .letme-db file. If current time has not yet surpassed
// expiry time, return true. Else, return false indicating new credentials need to be requested.
func CheckAccountAvailability(accountName string) bool {
	if _, err := os.Stat(GetHomeDirectory() + "/.letme/.letme-db"); err == nil {
		databaseFileReader, err := os.ReadFile(GetHomeDirectory() + "/.letme/.letme-db")
		CheckAndReturnError(err)
		fi, err := os.Stat(GetHomeDirectory() + "/.letme/.letme-db")
		CheckAndReturnError(err)
		if !json.Valid([]byte(databaseFileReader)) && fi.Size() > 0 {
			fmt.Printf("letme: " + GetHomeDirectory() + "/.letme/.letme-db" + " is not JSON valid. Remove the file and try again.\n")
			os.Exit(1)
		}
		var idents []Account
		json.Unmarshal(databaseFileReader, &idents) //should really check with _, err
		for i := range idents {
			if idents[i].Account.Name == accountName {
				t1 := time.Now().Unix()
				t2 := idents[i].Account.Expiry
				t3 := t2 - t1
				if t3 <= 0 {
					return false
				} else {
					return true
				}
			}
		}
	} else {
		_, err := os.OpenFile(GetHomeDirectory()+"/.letme/.letme-db", os.O_CREATE, 0600)
		CheckAndReturnError(err)
	}
	return false
}

// Check if the account to retrieve stored credentials exist, if true, return the credentials to stdout
func ReturnAccountCredentials(accountName string) map[string]string {
	databaseFileReader, err := os.ReadFile(GetHomeDirectory() + "/.letme/.letme-db")
	CheckAndReturnError(err)
	var idents []Account
	var result string
	m := make(map[string]string)
	err = json.Unmarshal(databaseFileReader, &idents)
	CheckAndReturnError(err)
	for i := range idents {
		if idents[i].Account.Name == accountName {
			result = idents[i].Account.V1Credentials
			data := CredentialsProcess{}
			json.Unmarshal([]byte(result), &data)
			m["AccessKeyId"] = data.AccessKeyId
			m["SecretAccessKey"] = data.SecretAccessKey
			m["SessionToken"] = data.SessionToken
		}
	}
	return m
}

// Remove an account from the database file
func RemoveAccountFromDatabaseFile(accountName string) {
	jsonData, err := ioutil.ReadFile(GetHomeDirectory() + "/.letme/.letme-db")
	CheckAndReturnError(err)
	// Unmarshal JSON data into a slice of maps
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		CheckAndReturnError(err)
	}

	// Iterate over each object in the slice
	for i, obj := range data {
		// Check if the "name" field of the "account" object is "adaral"
		if name, ok := obj["account"].(map[string]interface{})["name"].(string); ok && name == accountName {
			// Remove the object from the slice
			data = append(data[:i], data[i+1:]...)
			break // Break after removing to avoid index out of range error
		}
	}

	updatedJsonData, err := json.MarshalIndent(data, "", "  ")
	CheckAndReturnError(err)

	// Write the prettified JSON data to the file /test.json
	if err := ioutil.WriteFile(GetHomeDirectory()+"/.letme/.letme-db", updatedJsonData, 0600); err != nil {
		CheckAndReturnError(err)
	}
}

// UserSettings represents the structure of the .letme-usersettings file.
type UserSettings struct {
	ActiveContext string `json:"activeContext"`
}

// Create the .letme-usersettings file which holds the current context and more
func UpdateContext(context string) {
	filePath := GetHomeDirectory() + "/.letme/.letme-usersettings"

	var settings UserSettings

	// Check if the file exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, read the current content
		content, err := ioutil.ReadFile(filePath)
		CheckAndReturnError(err)

		// Unmarshal the current content into the settings struct
		err = json.Unmarshal(content, &settings)
		CheckAndReturnError(err)

	} else if !os.IsNotExist(err) {
		// An unexpected error occurred
		CheckAndReturnError(err)
	}

	// Update the activeContext field
	settings.ActiveContext = context

	// Marshal the settings struct to JSON
	newContent, err := json.Marshal(settings)
	CheckAndReturnError(err)


	// Write the updated JSON back to the file
	err = ioutil.WriteFile(filePath, newContent, 0644)
	CheckAndReturnError(err)

	err = os.Chmod(filePath, 0600)
	CheckAndReturnError(err)
}

// Get the current context used by letme

func GetCurrentContext() string {
	filePath := GetHomeDirectory() + "/.letme/.letme-usersettings"

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		CheckAndReturnError(err)
	}

	// Read the content of the file
	content, err := ioutil.ReadFile(filePath)
	CheckAndReturnError(err)

	// Unmarshal the content into the UserSettings struct
	var settings UserSettings
	if err := json.Unmarshal(content, &settings); err != nil {
		CheckAndReturnError(err)
	}

	// Return the value of the activeContext field
	return settings.ActiveContext
}