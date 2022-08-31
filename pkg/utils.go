package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func CommandExists(a string)  {
	_, err := exec.LookPath(a)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func CheckAndReturnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}