package main

import (
	"os"
	"ticTacSolved/task/cli/server"
)

func main() {
	if err := server.Execute(); err != nil {
		os.Exit(1)
	}
}
