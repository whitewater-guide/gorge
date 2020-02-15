package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whitewater-guide/gorge/version"
)

var (
	// Used for flags.
	endpointURL string

	rootCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [command] [arguments]", filepath.Base(os.Args[0])),
		Short:   "command line interface to gorge server",
		Version: version.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !strings.HasSuffix(endpointURL, "/") {
				endpointURL = endpointURL + "/"
			}
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&endpointURL, "url", "http://localhost:7080", "endpoint url")
}
