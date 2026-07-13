package main

import (
	"os"

	"ticTacSolved/task/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
