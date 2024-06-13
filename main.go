package main

import (
	"fmt"

	utils "github.com/hectorruiz-it/letme-alpha/pkg"
	letme "github.com/hectorruiz-it/letme-alpha/pkg/cmd"
)

func main() {
	if utils.CacheFileExists() {
		fmt.Println("letme: file" + utils.GetHomeDirectory() + "/.letme-alpha/.letme-cache" + " not supported anymore. Please remove it manually.")
	}
	utils.CommandExists("aws")
	letme.Execute()
}
