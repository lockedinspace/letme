package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func CommandExists(a string) bool {
	_, erra := exec.LookPath(a)
	if erra != nil {
		fmt.Println(erra)
		os.Exit(1)
		return false
	}
	return true
}
