package nzcan

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNzcan struct{}
type scriptNzcan struct {
	name string
	url  string
	// upstream assumes current year, but for stable tests we have to parametrize it
	year int
	core.LoggingScript
}

func (s *scriptNzcan) ListGauges() (core.Gauges, error) {
	locations, err := s.fetchGeo()
	if err != nil {
		return nil, err
	}
	msmnts := make(chan *core.Measurement)
	go func() {
		defer close(msmnts)
		err = s.fetchList("NORTH", msmnts)
		if err != nil {
			return
		}
		err = s.fetchList("SOUTH", msmnts)
	}()
	var result core.Gauges
	for m := range msmnts {
		loc := locations[m.Code]
		var levelUnit, flowUnit string
		if m.Flow.Valid() {
			flowUnit = "m3/s"
		}
		if m.Level.Valid() {
			levelUnit = "m"
		}
		result = append(result, core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   m.Code,
			},
			Name:      loc.name,
			URL:       "https://ecan.govt.nz/data/riverflow/sitedetails/" + m.Code,
			LevelUnit: levelUnit,
			FlowUnit:  flowUnit,
			Location:  loc.location,
		})
	}
	return result, err
}

func (s *scriptNzcan) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	err := s.fetchList("NORTH", recv)
	if err != nil {
		errs <- err
		return
	}
	err = s.fetchList("SOUTH", recv)
	if err != nil {
		errs <- err
		return
	}
}
