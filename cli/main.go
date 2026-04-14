package main

import (
	"os"

	"github.com/whitewater-guide/gorge/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
