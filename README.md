# letme [![Go Report Card](https://goreportcard.com/badge/github.com/lockedinspace/letme-go)](https://goreportcard.com/report/github.com/lockedinspace/letme-go) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/lockedinspace/letme)

## Requirements
- Go (recommended 1.19 or >= 1.16) installed in your system.
## What letme achieves
letme is a tool to obtain AWS credentials from another account without tampering your OS. 
It only requires a central AWS account with a DynamoDB table to store all of the other accounts information.

It is also mantained and developed under the following statement:

- A simple automation which writes/updates AWS credentials under your AWS files.

This achieves a lightweight, fast and not-intrusive tool that only reads from a DynamoDB database, authenticates the user (if MFA is enabled and AWS authorizes the assume role request) and adds the successful credentials into (``$HOME/.aws/credentials`` and ``$HOME/.aws/config``).

Later on, you can append the  ``--profile example1`` to your AWS cli operations and call resources from within example1's AWS account.

## What letme is not
As you can see, letme updates your AWS files enabling you to use call from multiple accounts.

What this software is not intended for:
- Securing your AWS files, letme just reads and writes to them. You should be responsible to keep them secure under your OS.
- Securing the AWS side (requiring MFA in your trust relationships, using a role with the least amount of privileges, etc.)

## Using letme
There are two possible scenarios, either you just want to install letme or you need to set up the aws infrastructure required. 

### Installing letme
Install letme with the command ``go install github.com/lockedinspace/letme@latest``. Go will automatically install it in your ``$GOPATH/bin`` directory which should be in your ``$PATH``.

### Installing letme from source
If you wish to install from source, clone the repository and build the executable with ``go build -o letme``. Afterwards, you must place the binary into your ``$PATH``.
