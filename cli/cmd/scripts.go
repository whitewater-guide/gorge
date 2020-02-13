package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whitewater-guide/gorge/core"
)

func init() {
	scriptsCmd := &cobra.Command{
		Use:   "scripts [command]",
		Short: "Set of command to view scripts information",
	}
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Lists available scripts",
		Run: func(cmd *cobra.Command, args []string) {
			var result []core.ScriptDescriptor
			err := Client.GetTo("scripts", &result)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			} else {
				printScriptsList(result)
			}
		},
	}
	scriptsCmd.AddCommand(listCmd)
	rootCmd.AddCommand(scriptsCmd)
}
