module github.com/lockedinspace/letme

go 1.19

require (
	github.com/BurntSushi/toml v1.2.0
	github.com/aws/aws-sdk-go v1.44.86
	github.com/spf13/cobra v1.5.0
)

require (
	github.com/google/go-github/v48 v48.2.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
)

retract (
	v0.1.3 // Not suitable for a stable production release
	v0.1.2 // Not suitable for a stable production release
	v0.1.1 // Not suitable for a stable production release
	v0.1.0 // Not suitable for a stable production release
)
