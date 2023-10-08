package smhi

import (
	"context"
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

type optionsSmhi struct{}

type scriptSmhi struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptSmhi) ListGauges() (core.Gauges, error) {
	var resp response
	if err := core.Client.GetAsJSON(fmt.Sprintf("%s/api/version/latest/parameter/2/station-set/all/period/latest-hour/data.json", s.url), &resp, nil); err != nil {
		return nil, err
	}
	var result []core.Gauge

	for _, f := range resp.Station {
		g := core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   fmt.Sprint(f.ID),
			},
			Name: fmt.Sprintf("%s (%d)", f.Name, f.ID),
			Location: &core.Location{
				Latitude:  core.TruncCoord(f.Latitude),
				Longitude: core.TruncCoord(f.Longitude),
			},
			FlowUnit: "m3/s",
			URL:      fmt.Sprintf("https://www.smhi.se/en/weather/observations/observations#ws=wpt-a,proxy=wpt-a,tab=vatten,param=waterflow,stationid=%d,type=water", f.ID),
			Timezone: "Europe/Stockholm",
		}
		result = append(result, g)
	}

	return result, nil
}

func (s *scriptSmhi) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	var resp response
	if err := core.Client.GetAsJSON(fmt.Sprintf("%s/api/version/latest/parameter/2/station-set/all/period/latest-hour/data.json", s.url), &resp, nil); err != nil {
		return
	}

	for _, f := range resp.Station {
		for _, v := range f.Value {
			recv <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   fmt.Sprint(f.ID),
				},
				Timestamp: core.HTime{Time: v.Date},
				Flow:      v.Value,
			}
		}
	}
}
