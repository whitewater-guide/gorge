package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/whitewater-guide/gorge/core"
)

func truncateString(str string, maxChars int) string {
	short := str
	if len(str) > maxChars {
		if maxChars > 3 {
			maxChars -= 3
		}
		short = str[0:maxChars] + "..."
	}
	return short
}

func printScriptsList(data []core.ScriptDescriptor) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Name", "Mode"})
	for i, s := range data {
		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			s.Name,
			s.Mode.String(),
		})
	}
	table.Render()
}

func printJobStatuses(data []core.JobDescription) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Job ID", "Timestamp", "Count/Error"})
	for _, j := range data {
		row := []string{j.ID, "", ""}
		if j.Status != nil {
			row[1] = j.Status.Timestamp.Format("2006-01-02T15:04:05")
			if j.Status.Success {
				row[2] = fmt.Sprintf("%d", j.Status.Count)
			} else {
				row[2] = j.Status.Error
			}
		}
		table.Append(row)
	}
	table.Render()
}

func printGaugeStatuses(data map[string]core.Status) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Gauge code", "Timestamp", "Count/Error"})

	for code, status := range data {
		row := []string{code, status.Timestamp.Format("2006-01-02T15:04:05"), ""}
		if status.Success {
			row[2] = fmt.Sprintf("%d", status.Count)
		} else {
			row[2] = status.Error
		}
		table.Append(row)
	}
	table.Render()
}

func printGauges(data []core.Gauge, truncURLs bool) {
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"#", "Code", "Name", "Flow unit", "Level unit", "Location", "URL"}
	table.SetHeader(header)
	for i, g := range data {
		loc := ""
		if g.Location != nil && g.Location.Longitude != 0 && g.Location.Latitude != 0 {
			loc = fmt.Sprintf("%.4f, %.4f", g.Location.Latitude, g.Location.Longitude)
			if g.Location.Altitude != 0 {
				loc = fmt.Sprintf("%s (%.f)", loc, g.Location.Altitude)
			}
		}
		url := g.URL
		if truncURLs {
			url = truncateString(url, 50)
		}
		row := []string{
			fmt.Sprintf("%d", i),
			g.Code,
			g.Name,
			g.FlowUnit,
			g.LevelUnit,
			loc,
			url,
		}
		table.Append(row)
	}
	table.Render()
}

func printMeasurements(measurements []core.Measurement) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Script", "Code", "Timestamp", "Flow", "Level"})
	table.SetFooter([]string{fmt.Sprintf("%d measurements total", len(measurements)), "", "", "", ""})
	for _, m := range measurements {
		flow, level := "", ""
		if m.Flow.Valid() {
			flow = fmt.Sprintf("%.2f", m.Flow.Float64Value())
		}
		if m.Level.Valid() {
			level = fmt.Sprintf("%.2f", m.Level.Float64Value())
		}
		table.Append([]string{
			m.GaugeID.Script,
			m.GaugeID.Code,
			m.Timestamp.UTC().Format("02/01/2006 15:04 MST"),
			flow,
			level,
		})
	}
	table.Render()
}

func printJobs(jobs []core.JobDescription) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Script", "Cron", "Options", "Gauge Code", "Gauge Options"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	for _, j := range jobs {
		for code, gOpts := range j.Gauges {
			table.Append([]string{
				j.ID,
				j.Script,
				j.Cron,
				string(j.Options),
				code,
				string(gOpts),
			})
		}
	}
	table.Render()
}
