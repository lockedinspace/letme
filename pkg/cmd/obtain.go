package letme

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	utils "github.com/hectorruiz-it/letme-alpha/pkg"
	"github.com/spf13/cobra"
)

type ProfileConfig struct {
	Output string `ini:"output"`
	Region string `ini:"region"`
}

type ProfileCredential struct {
	AccessKey    string `ini:"aws_access_key_id"`
	SecretKey    string `ini:"aws_secret_access_key"`
	SessionToken string `ini:"aws_session_token"`
}

type DynamoDbAccountConfig struct {
	Id          int      `dynamodbav:"id"`
	Description string   `dynamodbav:"description"`
	Name        string   `dynamodbav:"name"`
	Region      []string `dynamodbav:"region"`
	Role        []string `dynamodbav:"role"`
}

var obtainCmd = &cobra.Command{
	Use:     "obtain",
	Aliases: []string{"ob"},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(utils.GetHomeDirectory() + "/.letme-alpha/letme-config"); err == nil {
		} else {
			fmt.Println("letme: could not locate any config file. Please run 'letme config-file' to create one.")
			os.Exit(1)
		}
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme-alpha/letme-config")
		if result {
		} else {
			fmt.Println("letme: run 'letme config-file --verify' to obtain a template for your config file.")
			os.Exit(1)
		}
	},
	Short: "Obtain account credentials",
	Long: `Obtains credentials once the user authenticates itself.
Credentials will last 3600 seconds by default and can be used with the argument '--profile $ACCOUNT_NAME'
within the AWS cli binary.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// get flags
		inlineTokenMfa, _ := cmd.Flags().GetString("inline-mfa")
		renew, _ := cmd.Flags().GetBool("renew")
		localCredentialProcessFlagV1, _ := cmd.Flags().GetBool("v1")

		// get the current context
		currentContext := utils.GetCurrentContext()
		letmeContext := utils.GetContextData(currentContext)
		if letmeContext.AwsSessionDuration == 0 {
			letmeContext.AwsSessionDuration = 3600
		}

		// overwrite the session name variable if the user provides it
		if len(letmeContext.AwsSessionName) == 0 && !localCredentialProcessFlagV1 {
			fmt.Println("Using default session name: '" + args[0] + "-letme-session' with context: '" + currentContext + "'")
			letmeContext.AwsSessionName = args[0] + "-letme-session"
		} else if !localCredentialProcessFlagV1 {
			fmt.Println("Assuming role with the following session name: '" + letmeContext.AwsSessionName + "' and context: '" + currentContext + "'")
		}

		// grab the mfa arn from the config, create a new aws session and try to get credentials
		var authMethod string
		if len(letmeContext.AwsMfaArn) > 0 && !localCredentialProcessFlagV1 {
			authMethod = "mfa"
		} else if len(letmeContext.AwsMfaArn) > 0 && localCredentialProcessFlagV1 {
			authMethod = "mfa-credential-process-v1"
		} else if localCredentialProcessFlagV1 {
			authMethod = "credential-process-v1"
		} else {
			authMethod = "assume-role"
		}

		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(letmeContext.AwsSourceProfile), config.WithRegion(letmeContext.AwsSourceProfileRegion))
		utils.CheckAndReturnError(err)

		account := utils.GetAccount(letmeContext.AwsDynamoDbTable, cfg, args[0])
		var profileCredential utils.ProfileCredential
		var profileConfig utils.ProfileConfig
		switch {
		case len(account.Role) > 1:
			profileCredential, profileConfig = utils.AssumeRoleChained(letmeContext, cfg, inlineTokenMfa, account, renew, localCredentialProcessFlagV1, authMethod)
		default:
			profileCredential, profileConfig = utils.AssumeRole(letmeContext, cfg, inlineTokenMfa, account, renew, localCredentialProcessFlagV1, authMethod)
		}

		utils.LoadAwsCredentials(account.Name, profileCredential)
		utils.LoadAwsConfig(account.Name, profileConfig)
	},
}

func init() {
	var credentialProcess bool
	var v1 bool
	var renew bool
	rootCmd.AddCommand(obtainCmd)
	obtainCmd.Flags().String("inline-mfa", "", "pass the mfa token without user prompt")
	obtainCmd.Flags().BoolVarP(&renew, "renew", "", false, "force new credentials to be assumed")
	obtainCmd.Flags().BoolVarP(&credentialProcess, "credential-process", "", false, "obtain credentials using the credential_process entry in your aws config file.")
	obtainCmd.Flags().BoolVarP(&v1, "v1", "", false, "output credentials following the credential_process version 1 standard.")

}
