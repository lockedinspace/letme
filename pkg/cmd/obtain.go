package letme

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
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
	Short: "Obtain aws credentials",
	Long: `Obtains assumed credentials for the account specified.
Once the user successfully authenticates itself. Credentials will last 3600 seconds
and can be used with the argument '--profile example1' within the aws cli binary.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// grab and save fields from the config file into variables
		profile := utils.ConfigFileResultString("Aws_source_profile")
		region := utils.ConfigFileResultString("Aws_source_profile_region")
		table := utils.ConfigFileResultString("Dynamodb_table")
		sessionName := utils.ConfigFileResultString("Session_name")

		// overwrite the session name variable if the user provides it
		if len(sessionName) == 0 {
			fmt.Println("Using default session name: " + args[0] + "-letme-session")
			sessionName = args[0] + "-letme-session"
		} else {
			fmt.Println("Assuming role with the following session name: " + sessionName)
		}

		// grab the mfa arn from the config, create a new aws session and try to get credentials
		serialMfa := utils.ConfigFileResultString("Mfa_arn")
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
				fmt.Println("letme: account '" + args[0] + "' is already present under your aws files and it is not managed by letme.")
				fmt.Println("letme: this will cause duplicate entries hence parsing errors.")
				fmt.Println("letme: no changes were made.")
				os.Exit(1)
			} else if resultCred {
				fmt.Println("letme: account '" + args[0] + "' is already present under your aws credentials file and it is not managed by letme.")
				fmt.Println("letme: this will cause duplicate entries hence parsing errors.")
				fmt.Println("letme: no changes were made.")
				os.Exit(1)
			} else if resultConfig {
				fmt.Println("letme: account '" + args[0] + "' is already present under your aws config file and it is not managed by letme.")
				fmt.Println("letme: this will cause duplicate entries hence parsing errors.")
				fmt.Println("letme: no changes were made.")
				os.Exit(1)
			}

		}

		// check if the .letme-cache file exists, if not, queries must be satisfied through internet
		if utils.CacheFileExists() {

			// create a struct and a map to iterate over them
			type (
				accountFields struct {
					Id     int      `toml:"id"`
					Name   string   `toml:"name"`
					Region []string `toml:"region"`
					Role   []string `toml:"role"`
				}
				general map[interface{}]accountFields
			)
			var allitems general

			// save into a variable the name of the client parsed from the cache file and check if exists
			_, err := toml.DecodeFile(utils.GetHomeDirectory()+"/.letme/.letme-cache", &allitems)
			utils.CheckAndReturnError(err)
			var accountExists bool
			for _, value := range allitems {
				if value.Name == args[0] {
					accountExists = true
				}
			}

			if accountExists {
				file, err := os.Open(utils.GetHomeDirectory() + "/.letme/.letme-cache")
				utils.CheckAndReturnError(err)
				defer file.Close()
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					a := scanner.Text()
					_, err := regexp.MatchString("\\b"+args[0]+"\\b", a)
					utils.CheckAndReturnError(err)
				}

				// create a new aws session and declare the account and get the first role to assume
				svc := sts.New(sesAws)
				account := utils.ParseCacheFile(args[0])
				roleToAssumeArn := account.Role[0]

				// save into a variable the assume role output and check if mfa authentication is enabled
				var result *sts.AssumeRoleOutput
				if len(serialMfa) > 0 && len(account.Role) == 1 {
					fmt.Println("Enter MFA one time pass code: ")
					var tokenMfa string
					fmt.Scanln(&tokenMfa)
					result, err = svc.AssumeRole(&sts.AssumeRoleInput{
						RoleArn:         &roleToAssumeArn,
						RoleSessionName: &sessionName,
						SerialNumber:    &serialMfa,
						TokenCode:       &tokenMfa,
					})
					utils.CheckAndReturnError(err)

				} else if len(account.Role) > 1 {

					// check if account has more than one role, if true, start hopping between roles
					fmt.Println("More than one role detected. Total hops:", len(account.Role))
					var creds credentials.Value
					for i := 0; i < len(account.Role); i++ {
						fmt.Printf("[%v/%v]\n", i+1, len(account.Role))
						if i == 0 && len(serialMfa) > 0 {
							fmt.Printf("Enter MFA one time pass code: ")
							var tokenMfa string
							fmt.Scanln(&tokenMfa)
							result, err = svc.AssumeRole(&sts.AssumeRoleInput{
								RoleArn:         &account.Role[i],
								RoleSessionName: &sessionName,
								SerialNumber:    &serialMfa,
								TokenCode:       &tokenMfa,
							})
							utils.CheckAndReturnError(err)
							creds.AccessKeyID = *result.Credentials.AccessKeyId
							creds.SecretAccessKey = *result.Credentials.SecretAccessKey
							creds.SessionToken = *result.Credentials.SessionToken
						} else if i == 0 {
							result, err = svc.AssumeRole(&sts.AssumeRoleInput{
								RoleArn:         &account.Role[i],
								RoleSessionName: &sessionName,
							})
							utils.CheckAndReturnError(err)
							creds.AccessKeyID = *result.Credentials.AccessKeyId
							creds.SecretAccessKey = *result.Credentials.SecretAccessKey
							creds.SessionToken = *result.Credentials.SessionToken
						} else {
							chainAws, err := session.NewSession(&aws.Config{
								Credentials: credentials.NewStaticCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken),
							})
							utils.CheckAndReturnError(err)
							svcChain := sts.New(chainAws)
							result, err = svcChain.AssumeRole(&sts.AssumeRoleInput{
								RoleArn:         &account.Role[i],
								RoleSessionName: &sessionName,
							})
							utils.CheckAndReturnError(err)
						}
					}
				} else {
					result, err = svc.AssumeRole(&sts.AssumeRoleInput{
						RoleArn:         &roleToAssumeArn,
						RoleSessionName: &sessionName,
					})
					utils.CheckAndReturnError(err)
				}

				// save credentials outputs to variables
				var creds credentials.Value
				creds.AccessKeyID = *result.Credentials.AccessKeyId
				creds.SecretAccessKey = *result.Credentials.SecretAccessKey
				creds.SessionToken = *result.Credentials.SessionToken

				// open and read aws config & credentials files and create new variables
				credFile, errCred := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_APPEND, 0600)
				confFile, errConf := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_APPEND, 0600)
				str := "#s-" + account.Name
				etr := "#e-" + account.Name
				s := utils.AwsCredsFileRead()
				f := utils.AwsConfigFileRead()

				// check if client is not existing in credentials and config files
				if !(errors.Is(errCred, os.ErrNotExist)) && !(errors.Is(errConf, os.ErrNotExist)) {
					if strings.Contains(s, str) && strings.Contains(s, etr) && strings.Contains(f, str) && strings.Contains(f, etr) {
						credFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_TRUNC, 0600)
						utils.CheckAndReturnError(err)
						confFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_TRUNC, 0600)
						utils.CheckAndReturnError(err)
						fmt.Fprintf(credFile2, "%v", utils.AwsReplaceBlock(s, account.Name))
						fmt.Fprintf(confFile2, "%v", utils.AwsReplaceBlock(f, account.Name))
						if _, err = credFile2.WriteString(utils.AwsCredentialsFile(account.Name, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer credFile2.Close()
						}
						if _, err = confFile2.WriteString(utils.AwsConfigFile(account.Name, account.Region[0])); err != nil {
							utils.CheckAndReturnError(err)
							defer confFile2.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + account.Name + "' to interact with the account.\n")

						// check if client is existing in credentials but not found in config
					} else if strings.Contains(s, str) && strings.Contains(s, etr) && !(strings.Contains(f, str) && strings.Contains(f, etr)) {
						fmt.Fprintf(confFile, "%v", utils.AwsReplaceBlock(f, account.Name))
						if _, err = confFile.WriteString(utils.AwsConfigFile(account.Name, account.Region[0])); err != nil {
							utils.CheckAndReturnError(err)
							defer confFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + account.Name + "' to interact with the account.\n")
						fmt.Printf("letme: only modified '$HOME/.aws/config'. If you face problems while using the argument, please check your config file.\n")

						// check if client is not existing in credentials but found in config
					} else if !(strings.Contains(s, str) && strings.Contains(s, etr)) && strings.Contains(f, str) && strings.Contains(f, etr) {
						fmt.Fprintf(credFile, "%v", utils.AwsReplaceBlock(s, account.Name))
						if _, err = credFile.WriteString(utils.AwsCredentialsFile(account.Name, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer credFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + account.Name + "' to interact with the account.\n")
						fmt.Printf("letme: only modified '$HOME/.aws/credentials'. If you face problems while using the argument, please check your credentials files.\n")
					} else {
						if _, err = credFile.WriteString(utils.AwsCredentialsFile(account.Name, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer credFile.Close()
						}
						if _, err = confFile.WriteString(utils.AwsConfigFile(account.Name, account.Region[0])); err != nil {
							utils.CheckAndReturnError(err)
							defer confFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + account.Name + "' to interact with the account.\n")
					}
				} else {
					fmt.Println("letme: please check if the aws credentials and config files exists.")
					os.Exit(1)
				}
			} else {
				fmt.Printf("letme: account '" + args[0] + "' not found on your cache file. Try running 'letme init' to create a new updated cache file\n")
				os.Exit(1)
			}
		} else {
			// struct to map data
			type account struct {
				Id     int      `json:"id"`
				Name   string   `json:"name"`
				Role   []string `json:"role"`
				Region []string `json:"region"`
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
				if len(serialMfa) > 0 && len(roleToAssumeArn) > 1 {
					fmt.Println("More than one role detected. Total hops:", len(roleToAssumeArn))
					for i := 0; i < len(roleToAssumeArn); i++ {
						fmt.Printf("[%v/%v]\n", i+1, len(roleToAssumeArn))
						if i == 0 {
							fmt.Printf("Enter MFA one time pass code: ")
							var tokenMfa string
							fmt.Scanln(&tokenMfa)
							result, err = svc.AssumeRole(&sts.AssumeRoleInput{
								RoleArn:         &roleToAssumeArn[i],
								RoleSessionName: &sessionName,
								SerialNumber:    &serialMfa,
								TokenCode:       &tokenMfa,
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
						}
					}
				} else if len(roleToAssumeArn) > 1 {
					fmt.Println("More than one role detected. Total hops:", len(roleToAssumeArn))
					for i := 0; i < len(roleToAssumeArn); i++ {
						fmt.Printf("[%v/%v]\n", i+1, len(roleToAssumeArn))
						if i == 0 {
							result, err = svc.AssumeRole(&sts.AssumeRoleInput{
								RoleArn:         &roleToAssumeArn[i],
								RoleSessionName: &sessionName,
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
						}
					}
				} else {
					result, err = svc.AssumeRole(&sts.AssumeRoleInput{
						RoleArn:         &singleRoleToAssumeArn,
						RoleSessionName: &sessionName,
					})
				}
				utils.CheckAndReturnError(err)

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

						// check if client is existing in credentials but not found in config
					} else if strings.Contains(s, str) && strings.Contains(s, etr) && !(strings.Contains(f, str) && strings.Contains(f, etr)) {
						fmt.Fprintf(confFile, "%v", utils.AwsReplaceBlock(f, accountName))
						if _, err = confFile.WriteString(utils.AwsConfigFile(accountName, accountRegion)); err != nil {
							utils.CheckAndReturnError(err)
							defer confFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + accountName + "' to interact with the account.\n")
						fmt.Printf("letme: only modified '$HOME/.aws/config'. If you face problems while using the argument, please check your config file.\n")

						// check if client is not existing in credentials but found in config
					} else if !(strings.Contains(s, str) && strings.Contains(s, etr)) && strings.Contains(f, str) && strings.Contains(f, etr) {
						fmt.Fprintf(credFile, "%v", utils.AwsReplaceBlock(s, accountName))
						if _, err = credFile.WriteString(utils.AwsCredentialsFile(accountName, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer credFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + accountName + "' to interact with the account.\n")
						fmt.Printf("letme: only modified '$HOME/.aws/credentials'. If you face problems while using the argument, please check your credentials file.\n")
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
					fmt.Println("letme: please check if the aws credentials and config files exists.")
					os.Exit(1)
				}
			} else {
				fmt.Printf("letme: account '" + args[0] + "' not found on your dynamodb table '" + table + "'. Are you pointing to the correct table?\n")
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(obtainCmd)
}
