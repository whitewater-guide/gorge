package nzniwa

import (
	"context"
	"sync"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNzniwa struct{}
type scriptNzniwa struct {
	name        string
	flowURL     string
	locationURL string
	numWorkers  int
	core.LoggingScript
	// levelURL    string
}

func (s *scriptNzniwa) ListGauges() (core.Gauges, error) {
	gaugesCh := make(chan *core.Gauge)
	err := s.fetchGauges(gaugesCh)
	if err != nil {
		return nil, err
	}
	resultsCh := make(chan *core.Gauge)

	var wg sync.WaitGroup
	for w := 1; w <= s.numWorkers; w++ {
		wg.Add(1)
		go s.gaugePageWorker(gaugesCh, resultsCh, &wg)
	}
	go func() {
		wg.Wait()
		close(resultsCh)
	}()
	result := core.Gauges{}
	for g := range resultsCh {
		result = append(result, *g)
	}
	return result, nil
}

func (s *scriptNzniwa) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(errs)
	defer close(recv)
	s.fetchMeasurements(recv, errs)
}
