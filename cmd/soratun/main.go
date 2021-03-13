package main

import (
	"github.com/soracom/soratun/cmd"
	"os"
)

func main() {
	os.Exit(run())
}

func run() int {
	if err := cmd.RootCmd.Execute(); err != nil {
		return -1
	}
	return 0
}
