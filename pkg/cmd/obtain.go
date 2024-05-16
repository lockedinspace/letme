package letme

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/lockedinspace/letme/pkg"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
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
Credentials will last 3600 seconds and can be used with the argument '--profile ACCOUNT_NAME' 
within the AWS cli binary.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// grab and save fields from the config file into variables
		profile := utils.ConfigFileResultString("general", "Aws_source_profile").(string)
		region := utils.ConfigFileResultString("general", "Aws_source_profile_region").(string)
		table := utils.ConfigFileResultString("general", "Dynamodb_table").(string)
		sessionName := utils.ConfigFileResultString("general", "Session_name").(string)
		sessionDuration := utils.ConfigFileResultString("general", "Session_duration").(int64)
		if sessionDuration == 0 {
			sessionDuration = 3600
		}
		// grab credentials process flags
		credentialProcess, _ := cmd.Flags().GetBool("credential-process")
		renew, _ := cmd.Flags().GetBool("renew")
		localCredentialProcessFlagV1, _ := cmd.Flags().GetBool("v1")

		// overwrite the session name variable if the user provides it
		if len(sessionName) == 0 && !localCredentialProcessFlagV1 {
			fmt.Println("Using default session name: " + args[0] + "-letme-session")
			sessionName = args[0] + "-letme-session"
		} else if !localCredentialProcessFlagV1 {
			fmt.Println("Assuming role with the following session name: " + sessionName)
		}

		// grab the mfa arn from the config, create a new aws session and try to get credentials
		serialMfa := utils.ConfigFileResultString("general", "Mfa_arn").(string)
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
		sesAws, err := session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewSharedCredentials("", profile),
		})
		utils.CheckAndReturnError(err)
		_, err = sesAws.Config.Credentials.Get()
		utils.CheckAndReturnError(err)

		// check if the requested account is already found in the aws files
		if len(utils.CheckAccountLocally(args[0])) == 0 {
			resultCred, err := regexp.MatchString("\\["+args[0]+"]", utils.AwsCredsFileRead())
			utils.CheckAndReturnError(err)
			resultConfig, err := regexp.MatchString("\\[profile "+args[0]+"]", utils.AwsConfigFileRead())
			utils.CheckAndReturnError(err)
			if resultCred && resultConfig {
				fmt.Println("letme: account '" + args[0] + "' is already present under your AWS files and it is not managed by letme.")
				fmt.Println("letme: this will cause duplicate entries hence parsing errors.")
				fmt.Println("letme: no changes were made.")
				os.Exit(1)
			} else if resultCred {
				fmt.Println("letme: account '" + args[0] + "' is already present under your AWS credentials file and it is not managed by letme.")
				fmt.Println("letme: this will cause duplicate entries hence parsing errors.")
				fmt.Println("letme: no changes were made.")
				os.Exit(1)
			} else if resultConfig {
				fmt.Println("letme: account '" + args[0] + "' is already present under your AWS config file and it is not managed by letme.")
				fmt.Println("letme: this will cause duplicate entries hence parsing errors.")
				fmt.Println("letme: no changes were made.")
				os.Exit(1)
			}

		}

		// struct to map data
		type account struct {
			Id               int      `json:"id"`
			Name             string   `json:"name"`
			Role             []string `json:"role"`
			Region           []string `json:"region"`
			Session_duration int64    `json:"session_duration"`
		}

		// creating a new aws session and prepare a dynamodb query
		sesAwsDB := dynamodb.New(sesAws)
		proj := expression.NamesList(expression.Name("id"), expression.Name("name"), expression.Name("role"), expression.Name("region"))
		filt := expression.Name("name").Equal(expression.Value(args[0]))
		expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
		utils.CheckAndReturnError(err)
		inputs := &dynamodb.ScanInput{
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
			TableName:                 aws.String(table),
		}

		// once the query is prepared, scan the table name to retrieve the fields and loop through the results
		scanTable, err := sesAwsDB.Scan(inputs)
		utils.CheckAndReturnError(err)
		var accountName interface{}
		var roleToAssumeArn []string
		var singleRoleToAssumeArn string
		var accountRegion interface{}
		for _, i := range scanTable.Items {
			item := account{}
			err = dynamodbattribute.UnmarshalMap(i, &item)
			utils.CheckAndReturnError(err)
			if len(item.Role) > 1 {
				accountName = item.Name
				roleToAssumeArn = item.Role
				accountRegion = item.Region[0]
			} else {
				accountName = item.Name
				singleRoleToAssumeArn = item.Role[0]
				accountRegion = item.Region[0]
			}
		}

		// check if the account is the same as the provided by the user
		if accountName == args[0] {
			svc := sts.New(sesAws)
			var result *sts.AssumeRoleOutput
			var tempCreds credentials.Value
			// check if mfa authentication is enabled
			if len(serialMfa) > 0 && len(roleToAssumeArn) > 1 && !credentialProcess {
				if localCredentialProcessFlagV1 {
				} else {
					fmt.Println("More than one role detected. Total hops:", len(roleToAssumeArn))
				}
				for i := 0; i < len(roleToAssumeArn); i++ {
					if localCredentialProcessFlagV1 {
					} else {
						fmt.Printf("[%v/%v]\n", i+1, len(roleToAssumeArn))
					}
					if i == 0 {
						inlineTokenMfa, _ := cmd.Flags().GetString("inline-mfa")
						if len(inlineTokenMfa) > 0 {
							result, err = svc.AssumeRole(&sts.AssumeRoleInput{
								RoleArn:         &roleToAssumeArn[i],
								RoleSessionName: &sessionName,
								SerialNumber:    &serialMfa,
								TokenCode:       &inlineTokenMfa,
								DurationSeconds: &sessionDuration,
							})
							utils.CheckAndReturnError(err)
							tempCreds.AccessKeyID = *result.Credentials.AccessKeyId
							tempCreds.SecretAccessKey = *result.Credentials.SecretAccessKey
							tempCreds.SessionToken = *result.Credentials.SessionToken
						} else {
							if localCredentialProcessFlagV1 {
							} else {
								fmt.Printf("Enter MFA one time pass code: ")
							}
							var tokenMfa string
							fmt.Scanln(&tokenMfa)
							result, err = svc.AssumeRole(&sts.AssumeRoleInput{
								RoleArn:         &roleToAssumeArn[i],
								RoleSessionName: &sessionName,
								SerialNumber:    &serialMfa,
								TokenCode:       &tokenMfa,
								DurationSeconds: &sessionDuration,
							})
							utils.CheckAndReturnError(err)
							tempCreds.AccessKeyID = *result.Credentials.AccessKeyId
							tempCreds.SecretAccessKey = *result.Credentials.SecretAccessKey
							tempCreds.SessionToken = *result.Credentials.SessionToken
						}
					} else {
						chainAws, err := session.NewSession(&aws.Config{
							Credentials: credentials.NewStaticCredentials(tempCreds.AccessKeyID, tempCreds.SecretAccessKey, tempCreds.SessionToken),
						})
						utils.CheckAndReturnError(err)
						svcChain := sts.New(chainAws)
						result, err = svcChain.AssumeRole(&sts.AssumeRoleInput{
							RoleArn:         &roleToAssumeArn[i],
							RoleSessionName: &sessionName,
							DurationSeconds: &sessionDuration,
						})
						utils.CheckAndReturnError(err)
						if localCredentialProcessFlagV1 {
							fmt.Printf(utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration))
							os.Exit(0)
						}
					}
				}
			} else if len(roleToAssumeArn) > 1 && !credentialProcess {
				if localCredentialProcessFlagV1 {
				} else {
					fmt.Println("More than one role detected. Total hops:", len(roleToAssumeArn))
				}
				for i := 0; i < len(roleToAssumeArn); i++ {
					if localCredentialProcessFlagV1 {
					} else {
						fmt.Printf("[%v/%v]\n", i+1, len(roleToAssumeArn))
					}
					if i == 0 {
						result, err = svc.AssumeRole(&sts.AssumeRoleInput{
							RoleArn:         &roleToAssumeArn[i],
							RoleSessionName: &sessionName,
							DurationSeconds: &sessionDuration,
						})
						utils.CheckAndReturnError(err)
						tempCreds.AccessKeyID = *result.Credentials.AccessKeyId
						tempCreds.SecretAccessKey = *result.Credentials.SecretAccessKey
						tempCreds.SessionToken = *result.Credentials.SessionToken
					} else {
						chainAws, err := session.NewSession(&aws.Config{
							Credentials: credentials.NewStaticCredentials(tempCreds.AccessKeyID, tempCreds.SecretAccessKey, tempCreds.SessionToken),
						})
						utils.CheckAndReturnError(err)
						svcChain := sts.New(chainAws)
						result, err = svcChain.AssumeRole(&sts.AssumeRoleInput{
							RoleArn:         &roleToAssumeArn[i],
							RoleSessionName: &sessionName,
						})
						utils.CheckAndReturnError(err)
						if localCredentialProcessFlagV1 {
							fmt.Printf(utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration))
							os.Exit(0)
						}
					}
				}
			} else if len(serialMfa) > 0 && !credentialProcess {
				inlineTokenMfa, _ := cmd.Flags().GetString("inline-mfa")
				if len(inlineTokenMfa) > 0 {
					result, err = svc.AssumeRole(&sts.AssumeRoleInput{
						RoleArn:         &singleRoleToAssumeArn,
						RoleSessionName: &sessionName,
						SerialNumber:    &serialMfa,
						TokenCode:       &inlineTokenMfa,
						DurationSeconds: &sessionDuration,
					})
					utils.CheckAndReturnError(err)
					if localCredentialProcessFlagV1 {
						fmt.Printf(utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration))
						os.Exit(0)
					}
				} else {
					if renew {
						fmt.Printf("Enter MFA one time pass code: ")
						var tokenMfa string
						fmt.Scanln(&tokenMfa)
						result, err = svc.AssumeRole(&sts.AssumeRoleInput{
							RoleArn:         &singleRoleToAssumeArn,
							RoleSessionName: &sessionName,
							SerialNumber:    &serialMfa,
							TokenCode:       &tokenMfa,
							DurationSeconds: &sessionDuration,
						})
						utils.CheckAndReturnError(err)
						utils.DatabaseFile(args[0], sessionDuration, utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration), authMethod) //only when we really authenticate against aws
					} else if !utils.CheckAccountAvailability(args[0]) && len(serialMfa) > 0 { // if CheckAccountAvailability is false, request new MFA token, else, do not ask
						fmt.Printf("Enter MFA one time pass code: ")
						var tokenMfa string
						fmt.Scanln(&tokenMfa)
						result, err = svc.AssumeRole(&sts.AssumeRoleInput{
							RoleArn:         &singleRoleToAssumeArn,
							RoleSessionName: &sessionName,
							SerialNumber:    &serialMfa,
							TokenCode:       &tokenMfa,
							DurationSeconds: &sessionDuration,
						})
						utils.CheckAndReturnError(err)
						utils.DatabaseFile(args[0], sessionDuration, utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration), authMethod) //only when we really authenticate against aws
					} else if len(serialMfa) > 0 { //we do not ask mfa token when its on grace period, using credentials saved into keyring/encrypted into env vars
						v1 := utils.ReturnAccountCredentials(args[0])
						result, err = svc.AssumeRole(&sts.AssumeRoleInput{
							RoleArn:         &singleRoleToAssumeArn,
							RoleSessionName: &sessionName,
							DurationSeconds: &sessionDuration,
						})
						v1 = utils.ReturnAccountCredentials(args[0])
						*result.Credentials.AccessKeyId = v1["AccessKeyId"]
						*result.Credentials.SecretAccessKey = v1["SecretAccessKey"]
						*result.Credentials.SessionToken = v1["SessionToken"]
						if !localCredentialProcessFlagV1 {
							fmt.Println("letme: using cached credentials. Use argument --renew to obtain new credentials.")
						}
					}
					if localCredentialProcessFlagV1 {
						fmt.Printf(utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration))
						os.Exit(0)
					}
				}
			} else {
				if renew {
					result, err = svc.AssumeRole(&sts.AssumeRoleInput{
						RoleArn:         &singleRoleToAssumeArn,
						RoleSessionName: &sessionName,
						DurationSeconds: &sessionDuration,
					})
					utils.CheckAndReturnError(err)
					utils.DatabaseFile(args[0], sessionDuration, utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration), authMethod)
				} else if !utils.CheckAccountAvailability(args[0]) {
					result, err = svc.AssumeRole(&sts.AssumeRoleInput{
						RoleArn:         &singleRoleToAssumeArn,
						RoleSessionName: &sessionName,
						DurationSeconds: &sessionDuration,
					})
					utils.CheckAndReturnError(err)
					utils.DatabaseFile(args[0], sessionDuration, utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration), authMethod)
				} else if localCredentialProcessFlagV1 {
					fmt.Printf(utils.CredentialsProcessOutput(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, *result.Credentials.Expiration))
					os.Exit(0)
				} else {
					v1 := utils.ReturnAccountCredentials(args[0])
					result, err = svc.AssumeRole(&sts.AssumeRoleInput{
						RoleArn:         &singleRoleToAssumeArn,
						RoleSessionName: &sessionName,
						DurationSeconds: &sessionDuration,
					})
					v1 = utils.ReturnAccountCredentials(args[0])
					*result.Credentials.AccessKeyId = v1["AccessKeyId"]
					*result.Credentials.SecretAccessKey = v1["SecretAccessKey"]
					*result.Credentials.SessionToken = v1["SessionToken"]
					fmt.Println("letme: using cached credentials. Use argument --renew to obtain new credentials.")

				}
			}

			// save results into variables
			accountName := accountName.(string)
			accountRegion := accountRegion.(string)
			var creds credentials.Value
			creds.AccessKeyID = *result.Credentials.AccessKeyId
			creds.SecretAccessKey = *result.Credentials.SecretAccessKey
			creds.SessionToken = *result.Credentials.SessionToken

			// open and read aws config & credentials files and create new variables
			credFile, errCred := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_APPEND, 0600)
			confFile, errConf := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_APPEND, 0600)
			str := "#s-" + accountName
			etr := "#e-" + accountName
			s := utils.AwsCredsFileRead()
			f := utils.AwsConfigFileRead()

			if !(errors.Is(errCred, os.ErrNotExist)) && !(errors.Is(errConf, os.ErrNotExist)) && credentialProcess {
				if strings.Contains(f, str) && strings.Contains(f, etr) && strings.Contains(s, str) && strings.Contains(s, etr) {
					credFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_TRUNC, 0600)
					fmt.Fprintf(credFile2, "%v", utils.AwsReplaceBlock(s, accountName))
					confFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_TRUNC, 0600)
					fmt.Fprintf(confFile2, "%v", utils.AwsReplaceBlock(f, accountName))
					if _, err = confFile2.WriteString(utils.AwsConfigFileCredentialsProcessV1(accountName, accountRegion)); err != nil {
						utils.CheckAndReturnError(err)
						defer confFile2.Close()
					}
					fmt.Printf("letme: 123credentials ready to be sourced. Use '--profile " + accountName + "' to interact with the account.\n")
					os.Exit(0)
				} else if strings.Contains(f, str) && strings.Contains(f, etr) && !(strings.Contains(s, str) && strings.Contains(s, etr)) {
					confFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_TRUNC, 0600)
					fmt.Fprintf(confFile2, "%v", utils.AwsReplaceBlock(f, accountName))
					if _, err = confFile.WriteString(utils.AwsConfigFileCredentialsProcessV1(accountName, accountRegion)); err != nil {
						utils.CheckAndReturnError(err)
						defer confFile.Close()
					}
					fmt.Printf("letme: 212credentials ready to be sourced. Use '--profile " + accountName + "' to interact with the account.\n")
					os.Exit(0)
				} else {
					if _, err = confFile.WriteString(utils.AwsConfigFileCredentialsProcessV1(accountName, accountRegion)); err != nil {
						utils.CheckAndReturnError(err)
						defer confFile.Close()
					}
					fmt.Printf("letme: 787credentials ready to be sourced. Use '--profile " + accountName + "' to interact with the account.\n")
					os.Exit(0)
				}
			}

			// check if client is not existing in credentials and config files
			if !(errors.Is(errCred, os.ErrNotExist)) && !(errors.Is(errConf, os.ErrNotExist)) {
				if strings.Contains(s, str) && strings.Contains(s, etr) && strings.Contains(f, str) && strings.Contains(f, etr) {
					credFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_TRUNC, 0600)
					utils.CheckAndReturnError(err)
					confFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_TRUNC, 0600)
					utils.CheckAndReturnError(err)
					fmt.Fprintf(credFile2, "%v", utils.AwsReplaceBlock(s, accountName))
					fmt.Fprintf(confFile2, "%v", utils.AwsReplaceBlock(f, accountName))
					if _, err = credFile2.WriteString(utils.AwsCredentialsFile(accountName, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
						utils.CheckAndReturnError(err)
						defer credFile2.Close()
					}
					if _, err = confFile2.WriteString(utils.AwsConfigFile(accountName, accountRegion)); err != nil {
						utils.CheckAndReturnError(err)
						defer confFile2.Close()
					}
					fmt.Printf("letme: use the argument '--profile " + accountName + "' to interact with the account.\n")

				} else if !(strings.Contains(s, str) && strings.Contains(s, etr)) && strings.Contains(f, str) && strings.Contains(f, etr) {
					confFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_TRUNC, 0600)
					utils.CheckAndReturnError(err)
					fmt.Fprintf(confFile2, "%v", utils.AwsReplaceBlock(f, accountName))
					if _, err = confFile2.WriteString(utils.AwsConfigFile(accountName, accountRegion)); err != nil {
						utils.CheckAndReturnError(err)
						defer confFile2.Close()
					}
					fmt.Fprintf(credFile, "%v", utils.AwsReplaceBlock(s, accountName))
					if _, err = credFile.WriteString(utils.AwsCredentialsFile(accountName, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
						utils.CheckAndReturnError(err)
						defer credFile.Close()
					}
					fmt.Printf("123letme: use the argument '--profile " + accountName + "' to interact with the account.\n")
				} else {
					if _, err = credFile.WriteString(utils.AwsCredentialsFile(accountName, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
						utils.CheckAndReturnError(err)
						defer credFile.Close()
					}
					if _, err = confFile.WriteString(utils.AwsConfigFile(accountName, accountRegion)); err != nil {
						utils.CheckAndReturnError(err)
						defer confFile.Close()
					}
					fmt.Printf("letme: use the argument '--profile " + accountName + "' to interact with the account.\n")
				}
			} else {
				fmt.Println("letme: please check if aws credentials and config files exist.")
				os.Exit(1)
			}
		} else {
			fmt.Printf("letme: account '" + args[0] + "' not found in your dynamodb table '" + table + "'. Are you pointing to the correct table?\n")
			os.Exit(1)
		}

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
