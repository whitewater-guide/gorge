package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/scripts"
)

type jobSurveyAnswer struct {
	Script  string
	Cron    string
	Options string
}

type gaugeSurveyAnswer struct {
	Code    string
	Options string
	More    bool
}

func init() {
	jobsCmd := &cobra.Command{
		Use:   "jobs <command>",
		Short: "Set of commands to list, create and delete harvest jobs",
	}
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Lists currently running jobs",
		Run: func(cmd *cobra.Command, args []string) {
			var result []core.JobDescription
			err := Client.GetTo("jobs", &result)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			} else {
				printJobs(result)
			}
		},
	}
	addCmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "Add new job",
		Run: func(cmd *cobra.Command, args []string) {
			var scriptNames []string
			for _, d := range scripts.Registry.List() {
				scriptNames = append(scriptNames, d.Name)
			}

			var jobAnswer jobSurveyAnswer
			err := survey.Ask([]*survey.Question{
				{
					Name:   "script",
					Prompt: &survey.Select{Message: "Select script:", Options: scriptNames},
				},
				{
					Name:     "cron",
					Prompt:   &survey.Input{Message: "Cron schedule:"},
					Validate: validateCron,
				},
				{
					Name:     "options",
					Prompt:   &survey.Input{Message: "Script-level options (json string):", Default: "{}"},
					Validate: validateJSON,
				},
			}, &jobAnswer)
			if err != nil {
				fmt.Printf("Job survey error: %v", err)
				os.Exit(1)
			}
			id := uuid.Must(uuid.NewUUID())

			descr := core.JobDescription{
				ID:      id.String(),
				Script:  jobAnswer.Script,
				Cron:    jobAnswer.Cron,
				Options: json.RawMessage(jobAnswer.Options),
				Gauges:  map[string]json.RawMessage{},
			}

			for {
				var gAnswer gaugeSurveyAnswer
				err := survey.Ask([]*survey.Question{
					{
						Name:   "code",
						Prompt: &survey.Input{Message: "Gauge code:"},
					},
					{
						Name:     "options",
						Prompt:   &survey.Input{Message: "Gauge options:", Default: "{}"},
						Validate: validateJSON,
					},
					{
						Name:   "more",
						Prompt: &survey.Confirm{Message: "Add another gauge?", Default: true},
					},
				}, &gAnswer)
				if err != nil {
					fmt.Printf("Gauge survey error: %v", err)
					os.Exit(1)
				}
				descr.Gauges[gAnswer.Code] = json.RawMessage(gAnswer.Options)
				if !gAnswer.More {
					break
				}
			}
			var res core.JobDescription
			err = Client.PostTo("jobs", &descr, &res)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			} else {
				fmt.Println("Success")
			}
		},
	}
	deleteCmd := &cobra.Command{
		Use:     "remove <jobId>",
		Short:   "Stops and removes job by its id",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := Client.Delete("jobs/" + args[0])
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			} else {
				fmt.Println("Success")
			}
		},
	}
	jobsCmd.AddCommand(listCmd, addCmd, deleteCmd)
	rootCmd.AddCommand(jobsCmd)
}
