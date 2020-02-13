package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whitewater-guide/gorge/core"
)

func init() {
	statusCmd := &cobra.Command{
		Use:   "status [jobId]",
		Short: "Displays information about running jobs",
		Long:  "Displays information about running jobs.\nProvide job id or omit it to list all jobs",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				handleJobs()
			} else {
				handleGauges(args[0])
			}
		},
	}
	rootCmd.AddCommand(statusCmd)
}

func handleJobs() {
	var result []core.JobDescription
	err := Client.GetTo("jobs", &result)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	} else {
		printJobStatuses(result)
	}
}

func handleGauges(jobID string) {
	var result map[string]core.Status
	err := Client.GetTo("jobs/"+jobID+"/gauges", &result)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	} else {
		printGaugeStatuses(result)
	}
}
