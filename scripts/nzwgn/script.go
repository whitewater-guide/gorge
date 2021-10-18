package nzwgn

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNzwgn struct{}
type scriptNzwgn struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptNzwgn) ListGauges() (core.Gauges, error) {
	locations, err := s.fetchLocations()
	if err != nil {
		return nil, err
	}
	vals, err := s.fetchValues()
	if err != nil {
		return nil, err
	}
	result := core.Gauges{}
	for _, v := range vals {
		loc := locations[v.Code]
		result = append(result, core.Gauge{
			GaugeID:   v.GaugeID,
			Name:      v.name,
			URL:       "https://graphs.gw.govt.nz/",
			LevelUnit: v.levelUnit,
			FlowUnit:  v.flowUnit,
			Location:  &loc,
			Timezone:  "Pacific/Auckland",
		})
	}
	return result, nil
}

func (s *scriptNzwgn) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	data, err := s.fetchValues()
	if err != nil {
		errs <- err
		return
	}
	for _, v := range data {
		vv := v
		recv <- &vv.Measurement
	}
}
