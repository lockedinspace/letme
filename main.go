package main

import (
	"github.com/lockedinspace/letme-go/pkg"
	"github.com/lockedinspace/letme-go/pkg/cmd"
)

func main() {
	utils.CommandExists("aws")
	letme.Execute()
}
