# letme
[![Go Report Card](https://goreportcard.com/badge/github.com/lockedinspace/letme-go)](https://goreportcard.com/report/github.com/lockedinspace/letme-go)

## What letme achieves
letme is a tool to obtain AWS credentials from another account without tampering your OS. 
It only requires a central AWS account with a DynamoDB table to store all of the other accounts information.

It is also mantained and developed under the following statement:

- A simple automation which writes the new AWS credentials under your AWS files.

This achieves a lightweight, fast and not-intrusive tool that only reads from a DynamoDB database, authenticates the user (if done with MFA) and adds the successful credentials into (``$HOME/.aws/credentials`` and ``$HOME/.aws/config``).

Later on, you can append the  ``--profile example1`` to your cli operations and call resources from within example1's AWS account.

## What letme is not
As you can see, letme just updates your AWS files so you can use different accounts.

What this software does not achieve is:
- Securing your AWS files, letme just reads and writes to them. You should be responsible to keep them secure under your OS.
- Securing the AWS side (requiring MFA in your trust relationships, using a role with the least amount of privileges...)

## Using letme
There are two possible scenarios, either you just want to install letme or you need to set up the aws infrastructure required. 

### Installing letme