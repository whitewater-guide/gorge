package main

import (
	"github.com/whitewater-guide/gorge/cli/cmd"
)

// Version is provided by govvv at compile time
var Version string //nolint

func main() {
	cmd.Execute() //nolint:errcheck
}
