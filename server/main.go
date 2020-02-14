package main

// Version is provided by govvv at compile time
var Version string //nolint

func main() {
	rootCmd.Execute() //nolint:errcheck
}
