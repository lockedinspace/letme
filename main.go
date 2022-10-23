package main

import (
	"github.com/lockedinspace/letme/pkg"
	"github.com/lockedinspace/letme/pkg/cmd"
)

func main() {
	// check if aws binary is on the $PATH variable of the user
	utils.CommandExists("aws")

	letme.Execute()
}
