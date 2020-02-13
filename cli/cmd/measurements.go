package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/whitewater-guide/gorge/core"
)

const timeLayout = "2006-01-02 15:04"

func init() {
	measurementsCmd := &cobra.Command{
		Use:     "measurements <command>",
		Aliases: []string{"m"},
		Short:   "Prints harvested and stored measurements",
	}

	queryCmd := &cobra.Command{
		Use:   "query <script> [code] [--from XXX] [--to YYY]",
		Short: "Queries stored measurements. Default period is 1 day from now. Times are in UTC",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var fromS, toS string
			cmd.Flags().StringVar(&fromS, "from", "", "Start of time window, YYYY-MM-DD HH:MM")
			cmd.Flags().StringVar(&toS, "end", "", "Start of time window, YYYY-MM-DD HH:MM")

			var result []core.Measurement
			q, err := makeTimeQuery(fromS, toS)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
			path := fmt.Sprintf("measurements/%s?%s", strings.Join(args, "/"), q)
			err = Client.GetTo(path, &result)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			} else {
				printMeasurements(result)
			}
		},
	}

	latestCmd := &cobra.Command{
		Use:   "latest --script XXX --script YYY",
		Short: "Lists latest harvested measurements for given scripts",
		Run: func(cmd *cobra.Command, args []string) {
			scripts, err := cmd.Flags().GetStringSlice("script")
			if err != nil {
				fmt.Printf("Error: failed to parse scripts: %v", err)
				os.Exit(1)
			}
			var result []core.Measurement
			path := fmt.Sprintf("measurements/latest?scripts=%s", strings.Join(scripts, ","))
			err = Client.GetTo(path, &result)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			} else {
				printMeasurements(result)
			}
		},
	}
	latestCmd.Flags().StringSliceP("script", "s", []string{}, "script name")
	latestCmd.MarkFlagRequired("script") // nolint:errcheck

	measurementsCmd.AddCommand(queryCmd, latestCmd)
	rootCmd.AddCommand(measurementsCmd)
}

func makeTimeQuery(fromS, toS string) (string, error) {
	q := url.Values{}
	if fromS != "" {
		from, err := time.ParseInLocation(timeLayout, fromS, time.UTC)
		if err != nil {
			return "", fmt.Errorf("failed to parse time window start '%s': %v", fromS, err)
		}
		q.Add("from", fmt.Sprintf("%d", from.Unix()))
	}
	if toS != "" {
		to, err := time.ParseInLocation(timeLayout, toS, time.UTC)
		if err != nil {
			return "", fmt.Errorf("failed to parse time window end '%s': %v", toS, err)
		}
		q.Add("to", fmt.Sprintf("%d", to.Unix()))
	}
	return q.Encode(), nil
}
