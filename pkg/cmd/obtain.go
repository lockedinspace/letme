package letme

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/lockedinspace/letme-go/pkg"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
)

var obtainCmd = &cobra.Command{
	Use: "obtain",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(utils.GetHomeDirectory() + "/.letme/letme-config"); err == nil {
		} else {
			fmt.Println("letme: Could not locate any config file. Please run 'letme config-file' to create one.")
			os.Exit(1)
		}
	},
	Short: "Obtain aws credentials",
	Long: `Through the AWS Security Token Service, obtain temporal credentials
once the user successfully authenticates itself. Credentials will last 3600 seconds
and can be used with the argument '--profile example1' within the aws cli binary.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.ConfigFileResultString("Aws_source_profile")
		region := utils.ConfigFileResultString("Aws_source_profile_region")
		table := utils.ConfigFileResultString("Dynamodb_table")
		sesAws, err := session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewSharedCredentials("", profile),
		})
		utils.CheckAndReturnError(err)
		_, err = sesAws.Config.Credentials.Get()
		utils.CheckAndReturnError(err)
		if utils.CacheFileExists() {
			accountExists, err := regexp.MatchString("\\b"+args[0]+"\\b", utils.CacheFileRead())
			utils.CheckAndReturnError(err)
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
				svc := sts.New(sesAws)
				testvar := utils.ParseCacheFile(args[0])
				roleToAssumeArn := testvar.Role[0]
				sessionName := testvar.Name + "-letme-session"
				result, err := svc.AssumeRole(&sts.AssumeRoleInput{
					RoleArn:         &roleToAssumeArn,
					RoleSessionName: &sessionName,
				})
				utils.CheckAndReturnError(err)

				var creds credentials.Value
				creds.AccessKeyID = *result.Credentials.AccessKeyId
				creds.SecretAccessKey = *result.Credentials.SecretAccessKey
				creds.SessionToken = *result.Credentials.SessionToken

				credFile, errCred := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_APPEND, 0600)
				confFile, errConf := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_APPEND, 0600)
				str := "#s-" + testvar.Name
				etr := "#e-" + testvar.Name
				s := utils.AwsCredsFileRead()
				f := utils.AwsConfigFileRead()
				if !(errors.Is(errCred, os.ErrNotExist)) && !(errors.Is(errConf, os.ErrNotExist)) {
					if strings.Contains(s, str) && strings.Contains(s, etr) && strings.Contains(f, str) && strings.Contains(f, etr) {
						credFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_TRUNC, 0600)
						utils.CheckAndReturnError(err)
						confFile2, err := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_TRUNC, 0600)
						utils.CheckAndReturnError(err)
						fmt.Fprintf(credFile2, "%v", utils.AwsReplaceBlock(s, testvar.Name))
						fmt.Fprintf(confFile2, "%v", utils.AwsReplaceBlock(f, testvar.Name))
						if _, err = credFile2.WriteString(utils.AwsCredentialsFile(testvar.Name, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer credFile2.Close()
						}
						if _, err = confFile2.WriteString(utils.AwsConfigFile(testvar.Name, testvar.Region[0])); err != nil {
							utils.CheckAndReturnError(err)
							defer confFile2.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + testvar.Name + "' to interact with the account.\n")
					} else if strings.Contains(s, str) && strings.Contains(s, etr) && !(strings.Contains(f, str) && strings.Contains(f, etr)) {
						fmt.Fprintf(confFile, "%v", utils.AwsReplaceBlock(f, testvar.Name))
						if _, err = confFile.WriteString(utils.AwsConfigFile(testvar.Name, testvar.Region[0])); err != nil {
							utils.CheckAndReturnError(err)
							defer confFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + testvar.Name + "' to interact with the account.\n")
						fmt.Printf("letme: only modified '$HOME/.aws/config'. If you face problems while using the argument, please check your config file.\n")
					} else if !(strings.Contains(s, str) && strings.Contains(s, etr)) && strings.Contains(f, str) && strings.Contains(f, etr) {
						fmt.Fprintf(credFile, "%v", utils.AwsReplaceBlock(s, testvar.Name))
						if _, err = credFile.WriteString(utils.AwsCredentialsFile(testvar.Name, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer credFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + testvar.Name + "' to interact with the account.\n")
						fmt.Printf("letme: only modified '$HOME/.aws/credentials'. If you face problems while using the argument, please check your credentials files.\n")
					} else {
						if _, err = credFile.WriteString(utils.AwsCredentialsFile(testvar.Name, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer credFile.Close()
						}
						if _, err = confFile.WriteString(utils.AwsConfigFile(testvar.Name, testvar.Region[0])); err != nil {
							utils.CheckAndReturnError(err)
							defer confFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + testvar.Name + "' to interact with the account.\n")
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
			// create a struct to hold the data that will be passed into .letme-cache file
			type account struct {
				Id     int      `json:"id"`
				Name   string   `json:"name"`
				Role   []string `json:"role"`
				Region []string `json:"region"`
			}

			sesAwsDB := dynamodb.New(sesAws)
			proj := expression.NamesList(expression.Name("id"), expression.Name("name"), expression.Name("role"), expression.Name("region"))
			expr, err := expression.NewBuilder().WithProjection(proj).Build()
			utils.CheckAndReturnError(err)
			inputs := &dynamodb.ScanInput{
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ProjectionExpression:      expr.Projection(),
				TableName:                 aws.String(table),
			}
			// once the query is prepared, scan the table name (specified on letme-config) to retrieve the fields and loop through the results
			scanTable, err := sesAwsDB.Scan(inputs)
			utils.CheckAndReturnError(err)
			var accountName interface{}
			var roleToAssumeArn interface{}
			var accountRegion interface{}
			for _, i := range scanTable.Items {
				item := account{}
				err = dynamodbattribute.UnmarshalMap(i, &item)
				utils.CheckAndReturnError(err)
				accountName = item.Name
				roleToAssumeArn = item.Role[0]
				accountRegion = item.Region[0]

			}
			if accountName == args[0] {
				svc := sts.New(sesAws)
				sessionName := accountName.(string) + "-letme-session"
				roleToAssumeArnString := roleToAssumeArn.(string)
				result, err := svc.AssumeRole(&sts.AssumeRoleInput{
					RoleArn:         &roleToAssumeArnString,
					RoleSessionName: &sessionName,
				})
				utils.CheckAndReturnError(err)
				accountName := accountName.(string)
				accountRegion := accountRegion.(string)

				var creds credentials.Value
				creds.AccessKeyID = *result.Credentials.AccessKeyId
				creds.SecretAccessKey = *result.Credentials.SecretAccessKey
				creds.SessionToken = *result.Credentials.SessionToken

				credFile, errCred := os.OpenFile(utils.GetHomeDirectory()+"/.aws/credentials", os.O_RDWR|os.O_APPEND, 0600)
				confFile, errConf := os.OpenFile(utils.GetHomeDirectory()+"/.aws/config", os.O_RDWR|os.O_APPEND, 0600)
				str := "#s-" + accountName
				etr := "#e-" + accountName
				s := utils.AwsCredsFileRead()
				f := utils.AwsConfigFileRead()
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
					} else if strings.Contains(s, str) && strings.Contains(s, etr) && !(strings.Contains(f, str) && strings.Contains(f, etr)) {
						fmt.Fprintf(confFile, "%v", utils.AwsReplaceBlock(f, accountName))
						if _, err = confFile.WriteString(utils.AwsConfigFile(accountName, accountRegion)); err != nil {
							utils.CheckAndReturnError(err)
							defer confFile.Close()
						}
						fmt.Printf("letme: use the argument '--profile " + accountName + "' to interact with the account.\n")
						fmt.Printf("letme: only modified '$HOME/.aws/config'. If you face problems while using the argument, please check your config file.\n")
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
