package letme

import (
	"os"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
    "github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test purposes",
	Long:  `testing is vital.`,
	Run: func(cmd *cobra.Command, args []string) {
		// generating toml struct
		type generalParams struct {
			Aws_source_profile        string
			Aws_source_profile_region string
			Dynamodb_table            string
			Mfa_arn                   string `toml:"mfa_arn,omitempty"`
		}
		type general map[string]generalParams
		
		// reading file and decoding it
		a := "letme-config"
		b, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if _, err := os.Stat(a); err != nil {
			a = b+"/letme-config"
		}

		var generalConfig general

		c, err := toml.DecodeFile(a, &generalConfig)
		if err != nil {
			fmt.Printf("ERROR: Something went wrong while trying to decode the config file (letme-config). Verify any typos.\n")
			os.Exit(1)
		}
		d := c.Undecoded()

		if len(d) == 0 {
			fmt.Printf("letme-config is valid.\n")
		} else {
			fmt.Printf("ERROR: Unknown key: %q\n", d)
			os.Exit(1)
		}

		for _, name := range []string{"general"} {
			e := generalConfig[name]
			if len(e.Aws_source_profile) == 0 {
				fmt.Printf("ERROR: Unknown hash table. Verify its '[general]'\n")
				os.Exit(1)
			}
			fmt.Printf("\nUsing profile: %v\nSource profile region: %v\n\n", e.Aws_source_profile, e.Aws_source_profile_region)
			f, err := session.NewSession(&aws.Config{
				Region:      aws.String(e.Aws_source_profile_region),
				Credentials: credentials.NewSharedCredentials("", e.Aws_source_profile),
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = f.Config.Credentials.Get()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			type account struct {
				Id int `json:"id"`
				Name string `json:"name"`
				Description string `json:"description"`
	
			}
			g := e.Dynamodb_table
			h := dynamodb.New(f)
			proj := expression.NamesList(expression.Name("id"), expression.Name("name"))
			expr, err := expression.NewBuilder().WithProjection(proj).Build()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			j := &dynamodb.ScanInput{
				ExpressionAttributeNames:  expr.Names(),
				ExpressionAttributeValues: expr.Values(),
				FilterExpression:          expr.Filter(),
				ProjectionExpression:      expr.Projection(),
				TableName:                 aws.String(g),
			}
			k, err := h.Scan(j)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, i := range k.Items {
				item := account{}
		
				err = dynamodbattribute.UnmarshalMap(i, &item)
				
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
		
					fmt.Println(item.Id,item.Name)

			}
			
			
			
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
