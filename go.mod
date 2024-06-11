module github.com/hectorruiz-it/letme-alpha

go 1.22

require (
	github.com/aws/aws-sdk-go-v2 v1.27.2
	github.com/aws/aws-sdk-go-v2/config v1.27.18
	github.com/aws/aws-sdk-go-v2/credentials v1.17.18
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.14.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.32.8
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.12
	github.com/google/go-github/v48 v48.2.0
	github.com/hashicorp/go-version v1.6.0
	github.com/spf13/cobra v1.5.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.20.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.5 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
)

retract (
	v0.1.3 // Not suitable for a stable production release
	v0.1.2 // Not suitable for a stable production release
	v0.1.1 // Not suitable for a stable production release
	v0.1.0 // Not suitable for a stable production release
)
