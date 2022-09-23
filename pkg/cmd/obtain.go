package letme

import (
	"bufio"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
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
		sesAws, err := session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewSharedCredentials("", profile),
		})
		utils.CheckAndReturnError(err)
		_, err = sesAws.Config.Credentials.Get()
		utils.CheckAndReturnError(err)
		if utils.CacheFileExists() {
			//fmt.Println(strings.Split(utils.CacheFileRead(), ","))
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
				if _, err := os.Stat(utils.GetHomeDirectory() + "/.aws/credentials"); err == nil {
					str := "#s-" + testvar.Name
					etr := "#e-" + testvar.Name
					s := utils.AwsCredsFileRead()
					if strings.Contains(s, str) && strings.Contains(s, etr) {
						fmt.Println("It is already present, replacing...")
						startIndex := strings.Index(s, str)
						stopIndex := strings.Index(s, etr) + len(etr)
						res := s[:startIndex] + s[stopIndex:]
						res = strings.ReplaceAll(res, "\n\n", "\n")
						f, err := os.OpenFile(utils.GetHomeDirectory() + "/.aws/credentials", os.O_RDWR|os.O_TRUNC, 0600)
						utils.CheckAndReturnError(err)
						fmt.Fprintf(f, "%v", res)
						if _, err = f.WriteString(utils.AwsCredentialsFile(testvar.Name, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer f.Close()
						}
					} else {
						fmt.Println("It is not present, creating...")
						f, err := os.OpenFile(utils.GetHomeDirectory() + "/.aws/credentials", os.O_APPEND|os.O_WRONLY, 0600)
						utils.CheckAndReturnError(err)
						if _, err = f.WriteString(utils.AwsCredentialsFile(testvar.Name, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)); err != nil {
							utils.CheckAndReturnError(err)
							defer f.Close()
						}
					}

					

				} else {
					fmt.Println("letme: Could not locate '$HOME/.aws/credentials' file.")
					os.Exit(1)
				}

			} else {
				fmt.Printf("letme: account '" + args[0] + "' not found on your cache file. Try running 'letme init' to create a new updated cache file\n")
				os.Exit(1)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(obtainCmd)
}
