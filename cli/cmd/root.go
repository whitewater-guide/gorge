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
	vers        bool

	rootCmd = &cobra.Command{
		Use:   fmt.Sprintf("%s [command] [arguments]", filepath.Base(os.Args[0])),
		Short: "command line interface to gorge server",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !strings.HasSuffix(endpointURL, "/") {
				endpointURL = endpointURL + "/"
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if vers {
				fmt.Printf("gorge-cli version %s\n", version.Version)

				vServer := map[string]string{}
				err := Client.GetTo("version", &vServer)
				if err != nil {
					fmt.Printf("Error: %v", err)
					os.Exit(1)
				} else {
					fmt.Printf("endpoint version %s\n", vServer["version"])
				}
			} else {
				cmd.Help() //nolint:errcheck
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
	rootCmd.Flags().BoolVar(&vers, "version", false, "prints cli and server version")
}
