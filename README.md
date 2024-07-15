# letme [![Go Report Card](https://goreportcard.com/badge/github.com/lockedinspace/letme-go)](https://goreportcard.com/report/github.com/lockedinspace/letme-go) [![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/lockedinspace/letme) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) 

![](./docs/letme-banner.webp)
<p align="center">A <b>reliable</b>, <b>secure</b> and <b>fast</b> way to switch between AWS accounts from the CLI. </p>

## Documentation

To learn more about Letme [go to the complete documentation](https://lockedinspace.com/letme/index.html
).

## Why letme
As current AWS administrators, we've found that:

1. **It wasn't easy to manage** credentials from multiple accounts and **follow AWS best practices** at the same time.
2. Every team had a different tool to do the same task, we wanted a centralized way to manage credentials.

3. _"On my local computer works."_

## Requirements

- Go (recommended 1.22 or later).
- AWS CLI (recommended v2).

## Install Letme

- [Through go cli (recommended)](#go-cli)
- [Building from source](#building-from-source)
  
### Go CLI

Install the latest letme version with:

```bash
go install github.com/lockedinspace/letme@latest
```
> [Where does go install the binary?](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)

You can also install a specific version swapping ``@latest`` with your desired version.

Available versions can be found as tags in the [letme official repo](https://github.com/lockedinspace/letme). 


### Building from source

Clone the repository

```bash
git clone git@github.com:lockedinspace/letme.git
```

Change directory to letme and build the binary:

```bash
cd letme/
go build 
```

Move the ``letme`` binary to one of your ``$PATH`` (linux-macos) / ``$env:PATH`` (windows-poweshell) locations.

