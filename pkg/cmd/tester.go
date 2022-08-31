package letme

import (
	"os"
	"fmt"
	"bufio"
	"github.com/lockedinspace/letme-go/pkg"
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
			Aws_source_profile_region string `toml:"aws_source_profile_region,omitempty"`
			Dynamodb_table            string
			Mfa_arn                   string `toml:"mfa_arn,omitempty"`
		}
		type general map[string]generalParams
		
		// reading file and decoding it
		a := "letme-config"
		b, err := os.UserHomeDir()
		utils.CheckAndReturnError(err)

		if _, err := os.Stat(a); err != nil {
			a = b+"/letme-config"
		}

		var generalConfig general

		c, err := toml.DecodeFile(a, &generalConfig)
		utils.CheckAndReturnError(err)
		
		d := c.Undecoded()

		if len(d) == 0 {
			
		} else {
			fmt.Printf("ERROR: Unknown key: %q\n", d)
			os.Exit(1)
		}
		var exportedProfile string
		var exportedProfileRegion string
		var exportedDynamoDBTable string
		for _, name := range []string{"general"} {
			e := generalConfig[name]
			fmt.Printf("\nProfile: %v\nProfile region: %v\n\n", e.Aws_source_profile, e.Aws_source_profile_region)
			exportedProfile = e.Aws_source_profile
			exportedProfileRegion = e.Aws_source_profile_region
			exportedDynamoDBTable = e.Dynamodb_table
		}
		type account struct {
			Id int `json:"id"`
			Name string `json:"name"`
			Description string `json:"description"`

		}
		f, err := session.NewSession(&aws.Config{
			Region:      aws.String(exportedProfileRegion),
			Credentials: credentials.NewSharedCredentials("", exportedProfile),
		})
		utils.CheckAndReturnError(err)
		_, err = f.Config.Credentials.Get()
		utils.CheckAndReturnError(err)
		g := exportedDynamoDBTable
		h := dynamodb.New(f)
		proj := expression.NamesList(expression.Name("id"), expression.Name("name"))
		expr, err := expression.NewBuilder().WithProjection(proj).Build()
		utils.CheckAndReturnError(err)
		j := &dynamodb.ScanInput{
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			FilterExpression:          expr.Filter(),
			ProjectionExpression:      expr.Projection(),
			TableName:                 aws.String(g),
		}
		k, err := h.Scan(j)
		utils.CheckAndReturnError(err)
		var exportedId int
		for _, i := range k.Items {
			item := account{}
			err = dynamodbattribute.UnmarshalMap(i, &item)
			utils.CheckAndReturnError(err)
			exportedId = item.Id
			
		
		}
		m, err := os.Create(b+"/.letme-cache")
			utils.CheckAndReturnError(err)
			defer m.Close()
			fmt.Println("Cache file stored on " + b + "/.letme-cache")
			n := bufio.NewWriter(m)
			_, err = fmt.Fprintf(n, "%v;\n", exportedId)
			utils.CheckAndReturnError(err)
			n.Flush()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
