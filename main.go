package main

import (
	"github.com/lockedinspace/letme/pkg"
	"github.com/lockedinspace/letme/pkg/cmd"
)

func main() {
	utils.CommandExists("aws")
	letme.Execute()
}
