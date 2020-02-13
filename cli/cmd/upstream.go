package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/scripts"
)

func init() {
	upstreamCmd := &cobra.Command{
		Use:   "upstream <script> <command> [arguments]",
		Short: "Proxies commands to upstream data sources",
	}
	for _, descriptor := range scripts.Registry.List() {
		scriptCmd := &cobra.Command{
			Use:   fmt.Sprintf("%s <command> <arguments>", descriptor.Name),
			Short: fmt.Sprintf("Proxies commands to %s data source", descriptor.Name),
		}
		gaugesCmd := createGaugesCmd(descriptor)
		measurementsCmd := createMeasurementsCmd(descriptor)
		scriptCmd.AddCommand(gaugesCmd)
		scriptCmd.AddCommand(measurementsCmd)
		upstreamCmd.AddCommand(scriptCmd)
	}
	rootCmd.AddCommand(upstreamCmd)
}

func createGaugesCmd(descriptor core.ScriptDescriptor) *cobra.Command {
	cfg := descriptor.DefaultOptions()
	cmd := &cobra.Command{
		Use:   "gauges [flags]",
		Short: fmt.Sprintf("Lists all available gauges for script %s", descriptor.Name),
		Run: func(cmd *cobra.Command, args []string) {
			var result []core.Gauge
			err := Client.PostTo(fmt.Sprintf("upstream/%s/gauges", descriptor.Name), cfg, &result)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			} else {
				printGauges(result)
			}
		},
	}
	gFlags, err := gpflag.Parse(cfg)
	if err != nil {
		fmt.Printf("Failed to setup flags for script %s: %v\n", descriptor.Name, err)
		os.Exit(1)
	}
	cmd.Flags().AddFlagSet(gFlags)
	return cmd
}

func createMeasurementsCmd(descriptor core.ScriptDescriptor) *cobra.Command {
	cfg := descriptor.DefaultOptions()
	use := "measurements {codes}... [flags]"
	var args cobra.PositionalArgs
	if descriptor.Mode == core.OneByOne {
		use = "measurements <code> [flags]"
		args = cobra.ExactArgs(1)
	}
	cmd := &cobra.Command{
		Use:   use,
		Args:  args,
		Short: fmt.Sprintf("Gets latest measurements from upstream of script %s", descriptor.Name),
		Run: func(cmd *cobra.Command, args []string) {
			var result []core.Measurement
			q := url.Values{}
			if len(args) > 0 {
				q.Set("codes", strings.Join(args, ","))
			}
			url := fmt.Sprintf("upstream/%s/measurements?%s", descriptor.Name, q.Encode())
			err := Client.PostTo(url, cfg, &result)
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			} else {
				printMeasurements(result)
			}
		},
	}
	gFlags, err := gpflag.Parse(cfg)
	if err != nil {
		fmt.Printf("Failed to setup flags for script %s: %v\n", descriptor.Name, err)
		os.Exit(1)
	}
	cmd.Flags().AddFlagSet(gFlags)
	return cmd
}
