package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// Struct which represents the config-file toml keys
type GeneralParams struct {
	Aws_source_profile        string
	Aws_source_profile_region string `toml:"aws_source_profile_region,omitempty"`
	Dynamodb_table            string
	Mfa_arn                   string `toml:"mfa_arn,omitempty"`
	Session_name              string
	Session_duration          string `toml:"session_duration,omitempty"`
}

type DynamoDbAccountConfig struct {
	Name   string   `dynamodbav:"name"`
	Region []string `dynamodbav:"region"`
	Role   []string `dynamodbav:"role"`
}

type ProfileConfig struct {
	Output string `ini:"output"`
	Region string `ini:"region"`
}

type ProfileCredential struct {
	AccessKey    string `ini:"aws_access_key_id"`
	SecretKey    string `ini:"aws_secret_access_key"`
	SessionToken string `ini:"aws_session_token"`
}

// Verify if the config-file respects the struct LetmeContext
func CheckConfigFile(path string) bool {
	type config struct {
		General struct {
			Aws_source_profile        string
			Aws_source_profile_region string `toml:"aws_source_profile_region,omitempty"`
			Dynamodb_table            string
			Mfa_arn                   string `toml:"mfa_arn,omitempty"`
			Session_name              string
			Session_duration          string `toml:"session_duration,omitempty"`
		}
	}
	var conf config
	md, err := toml.DecodeFile(path, &conf)
	CheckAndReturnError(err)

	sections := config.Sections()

	for _, section := range sections {
		if section.Name() == "DEFAULT" {
			continue
		}
		for _, key := range MandatoryKeys {
			if ok := section.HasKey(key); !ok {
				fmt.Printf("letme: missing mandatory key '%s' in table '%s'. Config file should have the following structure:\n", key, section.Name())
				return false
			}
		}

		for _, key := range section.KeyStrings() {
			if !ExpectedKeys[key] {
				fmt.Printf("Error: Invalid key '%s' in table '%s'. Config file should have the following structure:\n", key, section.Name())
				return false
			}
		}
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

	if len(iamResp.MFADevices) == 0 {
		fmt.Println("letme: no MFA devices configured on you user. MFA configuration ommited.")
		return ""
	}

	var mfaDevices []string
	for _, device := range iamResp.MFADevices {
		mfaDevices = append(mfaDevices, *device.SerialNumber)
	}

	mfaArnExists := false
	for {
		fmt.Print("→ AWS MFA Device arn (optional): ")
		fmt.Scanln(&mfaArn)
		
		if len(mfaArn) == 0 {
			return ""
		}

		re := regexp.MustCompile(mfaArnRegex)
		switch re.MatchString(mfaArn) {
		case true: 
			for _, arn := range mfaDevices {
				if arn == mfaArn {
					mfaArnExists = true
					break
				}
			}
		case false:
			fmt.Println("letme: not a valid mfa device arn. Run 'aws iam list-mfa-devices --query 'MFADevices[*].SerialNumber --profile +'"+awsProfile)
			continue
			}	
		if !mfaArnExists {
			fmt.Println("letme: MFA Device not found. Run 'aws iam list-mfa-devices --query 'MFADevices[*].SerialNumber --profile '"+awsProfile)
			continue
		}
		break
	}
	return mfaArn
}

func sourceProfileRegionInput() string {
	var awsRegion string
	fmt.Print("→ AWS Source Profile region: ")
	fmt.Scanln(&awsRegion)

	return awsRegion
}

func sessionDurationInput() int32 {
	var sessionDuration int32
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("→ Token Session duration in seconds (optional): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if len(input) == 0 {
			input = "3600"
		}

		duration, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("letme: expected integer not a string.")
			continue
		} else if duration < 900 || duration > 43200 {
			fmt.Println("letme: token session duration cannot be lower than 15 minutes or higher than 12 hours.")
			continue
		} else {
			sessionDuration = int32(duration)
			break
		}
	}
	// fmt.Println(sessionDuration)
	return sessionDuration
}

func sessionNameInput() string {
	var sessionName string
	fmt.Print("→ Session Name (optional): ")
	fmt.Scanln(&sessionName)
	
	if len(sessionName) == 0 {
		return "letme_session"
	}

	return sessionName
}

func sourceProfileInput() string {
	config := AwsConfigFileReadV2()
	credentials := AwsCredsFileReadV2()
	var awsProfile string

	for {
		fmt.Print("→ AWS Source Profile Name: ")
		fmt.Scanln(&awsProfile)
		configProfileExists := false
		credentialsProfileExists := false

		if len(awsProfile) == 0 {
			fmt.Println("letme: AWS Profile Name field is required. Please introduce a value.")
			continue
		}

		if config.HasSection("profile "+awsProfile) || config.HasSection(awsProfile) {
			configProfileExists = true
		}

		if credentials.HasSection(awsProfile) {
			credentialsProfileExists = true
		}

		if !configProfileExists {
			fmt.Println("letme: profile name does not exist on your .aws/config files. Please specify a valid profile.")
			continue
		}

		if !credentialsProfileExists {
			fmt.Println("letme: profile name does not exist on your .aws/credentials files. Please specify a valid profile.")
			continue
		}
		break
	}
	return awsProfile
}

func dynamoDbTableInput(awsProfile string, awsRegion string) string {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(awsProfile), config.WithRegion(awsRegion))
	CheckAndReturnError(err)

	sesAwsDynamoDb := dynamodb.NewFromConfig(cfg)

	resp, err := sesAwsDynamoDb.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	CheckAndReturnError(err)
	var dynamoDbTableName string

	for {
		fmt.Print("→ AWS DynamoDB Table Name: ")
		fmt.Scanln(&dynamoDbTableName)

		if len(dynamoDbTableName) == 0 {
			fmt.Println("letme: DynamoDB Table Name field is required. Please introduce a value.")
			continue
		}

		tableExists := false
		for _, table := range resp.TableNames {
			if table == dynamoDbTableName {
				tableExists = true
			}
		}

		if !tableExists {
			fmt.Println("letme: DynamoDB Table not found. Please introduce an existing table.")
			continue
		}

		break
	}
	return dynamoDbTableName
}

func NewContext(context string) {
	fmt.Println("letme: creating context '" + context + "'. Optional fields can be left empty.")
	var letmeContext LetmeContext

	letmeContext.AwsSourceProfile = sourceProfileInput()
	letmeContext.AwsSourceProfileRegion = sourceProfileRegionInput()
	letmeContext.AwsDynamoDbTable = dynamoDbTableInput(letmeContext.AwsSourceProfile, letmeContext.AwsSourceProfileRegion)
	letmeContext.AwsMfaArn = mfaArnInput(letmeContext.AwsSourceProfile, letmeContext.AwsSourceProfileRegion)
	letmeContext.AwsSessionDuration = sessionDurationInput()
	letmeContext.AwsSessionName = sessionNameInput()

	letmeConfig := LetmeConfigRead()

	section := letmeConfig.Section(context)

	if err := section.ReflectFrom(&letmeContext); err != nil {
		CheckAndReturnError(err)
	}
	if len(letmeContext.AwsMfaArn) == 0 {
		section.DeleteKey("mfa_arn")
	}
	letmeConfig.SaveTo(GetHomeDirectory() + "/.letme/letme-config")
}

// Gets the user $HOME directory
func GetHomeDirectory() string {
	homeDir, err := os.UserHomeDir()
	CheckAndReturnError(err)
	return homeDir
}

// Checks if the .letme-cache file exists, this file is not supported starting from versions 0.2.0 and above
func CacheFileExists() bool {
	if _, err := os.Stat(GetHomeDirectory() + "/.letme/.letme-cache"); err == nil {
		return true
	} else {
		return false
	}
}

// Marshalls data into a string used for the aws config file but with the v1 output protocol
func AwsConfigFileCredentialsProcessV1(accountName string, region string) {
	credentials := AwsCredsFileReadV2()
	config := AwsConfigFileReadV2()

	accountInFile := CheckAccountLocally(accountName)
	switch {
	case accountInFile["credentials"]:
		credentials.DeleteSection(accountName)
		err := credentials.SaveTo(GetHomeDirectory() + "/.aws/credentials")
		fmt.Println("letme: removed profile '" + accountName + "' entry from credentials file.")
		CheckAndReturnError(err)
		fallthrough
	case accountInFile["config"] && !config.Section("profile "+accountName).HasKey("credential_process"):
		_, err := config.Section("profile "+accountName).NewKey("credential_process", "letme obtain "+accountName+" --v1")
		CheckAndReturnError(err)
		err = config.SaveTo(GetHomeDirectory() + "/.aws/config")
		CheckAndReturnError(err)
	default:
		section, errSection := config.NewSection("profile " + accountName)
		CheckAndReturnError(errSection)
		_, errCredentialProcess := section.NewKey("credential_process", "letme obtain "+accountName+" --v1")
		CheckAndReturnError(errCredentialProcess)
		_, errRegion := section.NewKey("region", region)
		CheckAndReturnError(errRegion)
		_, errOutput := section.NewKey("output", "json")
		CheckAndReturnError(errOutput)
		err := config.SaveTo(GetHomeDirectory() + "/.aws/config")
		CheckAndReturnError(err)
	}

	fmt.Println("letme: configured credential proces V1 for account " + accountName)
	os.Exit(0)
}

// Check if an account is present on the local aws credentials/config files
func CheckAccountLocally(account string) map[string]bool {
	credentials := AwsCredsFileReadV2()
	config := AwsConfigFileReadV2()

	accountInFile := make(map[string]bool)

	accountInFile["credentials"] = false
	accountInFile["config"] = false

	if credentials.HasSection(account) {
		accountInFile["credentials"] = true
	}

	if config.HasSection("profile " + account) {
		accountInFile["config"] = true
	}

	return accountInFile
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
func DatabaseFile(accountName string, sessionDuration int32, v1Credentials string, authMethod string) {
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
		LastRequest int64 `json:"lastRequest"`
	 	Expiry		int64 `json:"expiry"`
	}
	type Account struct {
		Account     string `json:"account"`
		Dataset	    Dataset `json:"data"`
	}

	var idents []Account

	CheckAndReturnError(err)
	if fi.Size() > 0 {
		err = json.Unmarshal(databaseFileReader, &idents)
		CheckAndReturnError(err)
		err = os.Truncate(GetHomeDirectory()+"/.letme/.letme-db", 0)
		CheckAndReturnError(err)
	}
	idents = append(idents, Account{accountName, Dataset{time.Now().Unix(), time.Now().Unix()}})
	b, err := json.MarshalIndent(idents, "", "  ")
	CheckAndReturnError(err)

	if _, err = databaseFileWriter.WriteString(string(b)); err != nil {
		CheckAndReturnError(err)
		defer databaseFileWriter.Close()
	}
}
