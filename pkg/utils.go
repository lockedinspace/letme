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

// Struct which represents the config-file toml keys
type GeneralParams struct {
	Aws_source_profile        string
	Aws_source_profile_region string `toml:"aws_source_profile_region,omitempty"`
	Dynamodb_table            string
	Mfa_arn                   string `toml:"mfa_arn,omitempty"`
	Session_name              string
	Session_duration          int64 `toml:"session_duration,omitempty"`
}

// Struct which represents the cache file toml keys
type CacheFields struct {
	Id     int      `toml:"id"`
	Name   string   `toml:"name"`
	Role   []string `toml:"role"`
	Region []string `toml:"region"`
}

// Verify config-file integrity
func CheckConfigFile(path string) bool {
	type config struct {
		General struct {
			Aws_source_profile        string
			Aws_source_profile_region string `toml:"aws_source_profile_region,omitempty"`
			Dynamodb_table            string
			Mfa_arn                   string `toml:"mfa_arn,omitempty"`
			Session_name              string
			Session_duration          int64 `toml:"session_duration,omitempty"`
		}
	}
	var conf config
	md, err := toml.DecodeFile(path, &conf)
	CheckAndReturnError(err)
	if len(md.Undecoded()) > 0 {
		fmt.Printf("letme: config file is corrupted. Following values might be misspelled:\n")
		fmt.Printf("* %v \n", md.Undecoded())
		return false
	} else {
		return true
	}
}

// Check if a command exists on the host machine
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
		"general": map[string]interface{}{
			"aws_source_profile":        "default",
			"aws_source_profile_region": "eu-west-3",
			"dynamodb_table":            "customers",
			"mfa_arn":                   "arn:aws:iam::3301048219:mfa/user",
			"session_name":              "user_letme",
			"session_duration":          "3600",
		},
	})
	CheckAndReturnError(err)
	return buf.String()
}

// Marshalls data into a toml file (.letme-cache)
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

// Gets user's $HOME directory
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
		fmt.Println("letme: error while retrieving field \""+ field + "\" could not be found in " + GetHomeDirectory()+"/.letme/letme-config")
		os.Exit(1)
    }
	return generalConfig[profile]
	// var exportedField interface{}
	// for _, name := range []string{profile} {
	// 	a := generalConfig[name]
	// 	r := reflect.ValueOf(a)
	// 	f := reflect.Indirect(r).FieldByName(field)
	// 	exportedField = string(f.String())

	// }
	// return exportedField
}

// Checks if a cache file exists
func CacheFileExists() bool {
	if _, err := os.Stat(GetHomeDirectory() + "/.letme/.letme-cache"); err == nil {
		return true
	} else {
		return false
	}
}

// Reads the cache file
func CacheFileRead() string {
	readCacheFile, err := ioutil.ReadFile(GetHomeDirectory() + "/.letme/.letme-cache")
	CheckAndReturnError(err)
	s := string(readCacheFile)
	return s
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

// Maps data from the cache file into a struct
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
}

// Marshalls data into a string
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

// Marshalls data into a string
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

// Marshalls data into a string
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

// Return aws credentials following the credentials_process standard
// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html

func CredentialsProcessOutput(accessKeyID string, secretAccessKey string, sessionToken string, expirationTime time.Time) string {
	type CredentialsProcess struct {
		Version         int
		AccessKeyId     string
		SecretAccessKey string
		SessionToken    string
		Expiration      time.Time
	}
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

// Create a file which stores the last time when credentials where requested. Then query if the account exists,
// if not, it will create its first entry. If it already exists, it will either return true (if credemtials are still within the session_duration)
// and false if credentials have already been expired.
func CheckAccountDatabaseFile(accountName string, sessionDuration int64) {
	databaseFileWriter, err := os.OpenFile(GetHomeDirectory()+"/.letme/.letme-db", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	CheckAndReturnError(err)
	databaseFileReader, err := os.ReadFile(GetHomeDirectory() + "/.letme/.letme-db")
	CheckAndReturnError(err)
	fi, err := os.Stat(GetHomeDirectory() + "/.letme/.letme-db")
	CheckAndReturnError(err)
	if !json.Valid([]byte(databaseFileReader)) && fi.Size() > 0 {
		fmt.Printf("letme: " + GetHomeDirectory() + "/.letme/.letme-db" + " is not JSON valid.\n")
		os.Exit(1)
	}
	type Dataset struct {
		Name		string `json:"name"`
		LastRequest int64 `json:"lastRequest"`
	 	Expiry		int64 `json:"expiry"`
	}
	type Account struct {
		Account     Dataset `json:"account"`
	}
	
	var idents []Account
	if fi.Size() > 0 {
		err = json.Unmarshal(databaseFileReader, &idents)
		CheckAndReturnError(err)
		err = os.Truncate(GetHomeDirectory()+"/.letme/.letme-db", 0)
		CheckAndReturnError(err)
		for i := range idents {
			if idents[i].Account.Name == accountName {
				idents[i].Account.LastRequest = time.Now().Unix()
				idents[i].Account.Expiry = time.Now().Add(time.Second * time.Duration(sessionDuration)).Unix()
				b, err := json.MarshalIndent(idents, "", "  ")
				CheckAndReturnError(err)

				if _, err = databaseFileWriter.WriteString(string(b)); err != nil {
					CheckAndReturnError(err)
					defer databaseFileWriter.Close()
				}
				return
			} 
		}
		idents = append(idents, Account{Dataset{accountName, time.Now().Unix(), time.Now().Add(time.Second * time.Duration(sessionDuration)).Unix()}})
		b, err := json.MarshalIndent(idents, "", "  ")
		CheckAndReturnError(err)

		if _, err = databaseFileWriter.WriteString(string(b)); err != nil {
			CheckAndReturnError(err)
			defer databaseFileWriter.Close()
		}
	} else if fi.Size() == 0 {
		idents = append(idents, Account{Dataset{accountName, time.Now().Unix(), time.Now().Unix()}})
		b, err := json.MarshalIndent(idents, "", "  ")
		CheckAndReturnError(err)

		if _, err = databaseFileWriter.WriteString(string(b)); err != nil {
			CheckAndReturnError(err)
			defer databaseFileWriter.Close()
		}
	} 
}
