package main

import (
	"os"

	"ticTacSolved/task/cli/client"
)

func main() {
	if err := client.Execute(); err != nil {
		os.Exit(1)
	}
}
