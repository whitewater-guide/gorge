package galicia2

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsGalicia2 struct{}

type scriptGalicia2 struct {
	name           string
	listURL        string
	gaugeURLFormat string
	core.LoggingScript
}

func (s *scriptGalicia2) ListGauges() (core.Gauges, error) {
	items, err := s.parseTable()
	if err != nil {
		return nil, err
	}

	jobsCh := make(chan *core.Gauge, len(items))
	resultsCh := make(chan *core.Gauge, len(items))

	for w := 1; w <= 10; w++ {
		go s.gaugePageWorker(jobsCh, resultsCh)
	}
	for i := range items {
		jobsCh <- &items[i].gauge
	}
	close(jobsCh)
	result := core.Gauges{}
	for range items {
		g := <-resultsCh
		result = append(result, *g)
	}
	close(resultsCh)
	return result, nil
}

func (s *scriptGalicia2) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	gauges, err := s.parseTable()
	if err != nil {
		errs <- err
		return
	}
	for _, g := range gauges {
		m := g.measurement
		recv <- &m
	}
}

func (s *scriptGalicia2) gaugePageWorker(gauges <-chan *core.Gauge, results chan<- *core.Gauge) {
	for g := range gauges {
		latitude, longitude, altitude := s.parseGaugePage(g.Code)
		g.Location = &core.Location{
			Latitude:  latitude,
			Longitude: longitude,
			Altitude:  altitude,
		}
		results <- g
	}
}
