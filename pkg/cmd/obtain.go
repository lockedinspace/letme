package letme

import (
        "fmt"
        "github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/credentials"
        "github.com/aws/aws-sdk-go/aws/session"
        "github.com/lockedinspace/letme-go/pkg"
        "github.com/spf13/cobra"
        "os"
)

var obtainCmd = &cobra.Command{
        Use: "obtain",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
                if _, err := os.Stat(utils.GetHomeDirectory() + "/.letme/letme-config"); err == nil {
                        fmt.Println("Config file found.")
                } else {
                        fmt.Println("letme: Could not locate any config file. Please run 'letme config-file' to create one.")
                        os.Exit(1)
                }
        },
        Short: "Obtain aws credentials",
        Long: `Through the AWS Security Token Service, obtain temporal credentials
once the user successfully authenticates itself. Credentials will last 3600 seconds
and can be used with the argument '--profile example1' within the aws cli binary.`,
        Run: func(cmd *cobra.Command, args []string) {
                fmt.Println("obtain called")
                profile := utils.ConfigFileResultString("Aws_source_profile")
                region := utils.ConfigFileResultString("Aws_source_profile_region")
                sesAws, err := session.NewSession(&aws.Config{
                        Region:      aws.String(region),
                        Credentials: credentials.NewSharedCredentials("", profile),
                })
                utils.CheckAndReturnError(err)
                _, err = sesAws.Config.Credentials.Get()
                utils.CheckAndReturnError(err)

        },
}

func init() {
        rootCmd.AddCommand(obtainCmd)
}