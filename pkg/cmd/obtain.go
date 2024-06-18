package letme

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	utils "github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
	"os"
)

var obtainCmd = &cobra.Command{
	Use:     "obtain",
	Aliases: []string{"ob"},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(utils.GetHomeDirectory() + "/.letme/letme-config"); err == nil {
		} else {
			fmt.Println("letme: could not locate any config file. Please run 'letme config-file' to create one.")
			os.Exit(1)
		}
		result := utils.CheckConfigFile(utils.GetHomeDirectory() + "/.letme/letme-config")
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
		// get the current context
		context := utils.GetCurrentContext()
		
		// grab and save fields from the config file into variables
		profile := utils.ConfigFileResultString(context, "Aws_source_profile").(string)
		region := utils.ConfigFileResultString(context, "Aws_source_profile_region").(string)
		table := utils.ConfigFileResultString(context, "Dynamodb_table").(string)
		sessionName := utils.ConfigFileResultString(context, "Session_name").(string)
		sessionDuration := utils.ConfigFileResultString(context, "Session_duration").(int64)
		if sessionDuration == 0 {
			sessionDuration = 3600
		}
		// grab credentials process flags
		credentialProcess, _ := cmd.Flags().GetBool("credential-process")
		localCredentialProcessFlagV1, _ := cmd.Flags().GetBool("v1")

		// overwrite the session name variable if the user provides it
		if len(sessionName) == 0 && !localCredentialProcessFlagV1 {
			fmt.Println("Using default session name: '" + args[0] + "-letme-session' with context: '" + context + "'")
			sessionName = args[0] + "-letme-session"
		} else if !localCredentialProcessFlagV1 {
			fmt.Println("Assuming role with the following session name: '" + sessionName + "' and context: '" + context + "'")
		}

		// grab the mfa arn from the config, create a new aws session and try to get credentials
		serialMfa := utils.ConfigFileResultString(context, "Mfa_arn").(string)
		var authMethod string
		if len(serialMfa) > 0 && !localCredentialProcessFlagV1 {
			authMethod = "mfa"
		} else if len(serialMfa) > 0 && localCredentialProcessFlagV1 {
			authMethod = "mfa-credential-process-v1"
		} else if localCredentialProcessFlagV1 {
			authMethod = "credential-process-v1"
		} else {
			authMethod = "assume-role"
		}

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
		fmt.Println("letme: use the argument --profile '" + account.Name + "' to interact with the account.")

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
