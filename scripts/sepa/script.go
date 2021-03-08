package sepa

import (
	"context"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

type optionsSepa struct{}
type scriptSepa struct {
	name         string
	listURL      string
	gaugeURLBase string

	core.LoggingScript
}

func (s *scriptSepa) ListGauges() (result core.Gauges, err error) {
	err = core.Client.StreamCSV(
		s.listURL+"/SEPA_River_Levels_Web.csv",
		func(row []string) error {
			if row[3] == "---" {
				return nil
			}
			raw := gaugeFromRow(row)
			g, err := s.getGauge(raw)
			if err == nil {
				result = append(result, g)
			} else {
				s.GetLogger().WithField("row", strings.Join(row, ", ")).Errorf("failed to convert row: %v", err)
			}
			return nil
		},
		core.CSVStreamOptions{
			NumColumns:   20,
			HeaderHeight: 1,
		},
	)
	return
}

func (s *scriptSepa) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	code, err := codes.Only()
	if err != nil {
		errs <- err
		return
	}

	err = core.Client.StreamCSV(
		s.gaugeURLBase+"/"+code+"?csv=true",
		func(row []string) error {
			m, err := measurementFromRow(row)
			if err == nil {
				(*m).GaugeID = core.GaugeID{
					Script: s.name,
					Code:   code,
				}
				recv <- m
			} else {
				s.GetLogger().WithField("row", strings.Join(row, ", ")).Errorf("failed to convert row: %v", err)
			}
			return nil
		},
		core.CSVStreamOptions{
			NumColumns:   2,
			HeaderHeight: 7,
		},
	)
	if err != nil {
		errs <- err
	}
}
