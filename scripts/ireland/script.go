package ireland

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

type optionsIreland struct{}

type scriptIreland struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptIreland) ListGauges() (core.Gauges, error) {
	var resp geojson
	if err := core.Client.GetAsJSON(fmt.Sprintf("%s?%d", s.url, time.Now().Unix()), &resp, nil); err != nil {
		return nil, err
	}
	var result []core.Gauge

	for _, f := range resp.Features {
		if f.Properties.SensorRef != "0001" {
			continue
		}
		ref := f.Properties.StationRef
		g := core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   ref[len(ref)-5:],
			},
			Name: f.Properties.StationName,
			Location: &core.Location{
				Latitude:  f.Geometry.Coordinates[1],
				Longitude: f.Geometry.Coordinates[0],
			},
			LevelUnit: "m",
			URL:       fmt.Sprintf("https://waterlevel.ie/%s/0001/", ref),
			Timezone:  "Europe/Dublin",
		}
		result = append(result, g)
	}

	return result, nil
}

func (s *scriptIreland) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	var resp geojson
	if err := core.Client.GetAsJSON(fmt.Sprintf("%s?%d", s.url, time.Now().Unix()), &resp, nil); err != nil {
		errs <- err
		return
	}

	for _, f := range resp.Features {
		if f.Properties.SensorRef != "0001" {
			continue
		}
		ref := f.Properties.StationRef

		var val nulltype.NullFloat64
		json.Unmarshal([]byte(f.Properties.Value), &val)

		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   ref[len(ref)-5:],
			},
			Timestamp: core.HTime{
				Time: f.Properties.Datetime,
			},
			Level: val,
		}
	}
}
