package ireland2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

type optionsIreland2 struct{}

type scriptIreland2 struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptIreland2) fetch() ([]river, error) {
	raw, err := core.Client.GetAsString(s.url, nil)
	if err != nil {
		return nil, err
	}
	i := strings.LastIndex(raw, "<body>")
	raw = raw[i+len("<body>"):]

	var resp riverspyResp
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}

	return resp.Rivers, nil
}

func (s *scriptIreland2) ListGauges() (core.Gauges, error) {
	rivers, err := s.fetch()
	if err != nil {
		return nil, err
	}
	result := make([]core.Gauge, len(rivers))
	for i, r := range rivers {
		lUnit, fUnit := "cm", ""
		if r.Yunit == "cumecs" {
			fUnit, lUnit = "m3/s", ""
		}
		result[i] = core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   r.Code,
			},
			Name:      fmt.Sprintf("%s - %s", r.Rivername, r.Sitename),
			URL:       strings.Replace(r.Graphlink, "http:", "https:", 1),
			LevelUnit: lUnit,
			FlowUnit:  fUnit,
			Location: &core.Location{
				Latitude:  r.Latitude,
				Longitude: r.Longitude,
			},
			Timezone: "Europe/Dublin",
		}
	}

	return result, nil
}

func (s *scriptIreland2) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	rivers, err := s.fetch()
	if err != nil {
		errs <- err
		return
	}

	for _, r := range rivers {
		var l, f nulltype.NullFloat64
		if r.Yunit == "cumecs" {
			f = r.Lastlevel
		} else {
			l = r.Lastlevel
		}

		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   r.Code,
			},
			Timestamp: core.HTime{
				Time: time.Unix(r.Updated, 0).UTC(),
			},
			Level: l,
			Flow:  f,
		}
	}
}
