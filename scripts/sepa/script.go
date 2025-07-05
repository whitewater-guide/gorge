package sepa

import (
	"context"
	"fmt"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

type optionsSepa struct{}
type scriptSepa struct {
	name    string
	listURL string
	apiURL  string

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

	var resp SEPAStationMeasurements
	// This harvests levels only. To harvest flows, we need to make another reauest, because two wildcard ts_path parameters are not supported
	if err := core.Client.GetAsJSON(fmt.Sprintf("%s?service=kisters&type=queryServices&datasource=0&request=getTimeseriesValues&returnfields=Timestamp,Value&metadata=true&md_returnfields=station_no,ts_unitsymbol,stationparameter_name&format=dajson&ts_path=1/*/SG/15m.Cmd", s.apiURL), &resp, nil); err != nil {
		errs <- err
		return
	}

	for _, station := range resp {
		if station.StationParameterName != "Level" {
			continue
		}
		for _, d := range station.Data {
			recv <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   station.StationNo,
				},
				Timestamp: d.Timestamp,
				Level:     d.Value,
			}
		}
	}
}
