package main

import (
	"os"
	"ticTacSolved/task/cli/client/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
