package main

import (
	"fmt"
	"github.com/lockedinspace/letme/pkg"
	"github.com/lockedinspace/letme/pkg/cmd"
)

func main() {
	if utils.CacheFileExists() {
		fmt.Println("letme: file" + utils.GetHomeDirectory() + "/.letme/.letme-cache" + " not supported anymore. Please remove it manually.")
	}
	utils.CommandExists("aws")
	letme.Execute()
}
