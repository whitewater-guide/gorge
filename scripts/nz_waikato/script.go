package nz_waikato

import (
	"context"
	"sync"

	"github.com/whitewater-guide/gorge/core"
)

type optionsWaikato struct{}
type scriptWaikato struct {
	name       string
	listURL    string
	pageURL    string
	numWorkers int
	core.LoggingScript
}

func (s *scriptWaikato) ListGauges() (core.Gauges, error) {
	listCh := make(chan *core.Measurement)
	resultsCh := make(chan *core.Gauge)
	var err error
	var results core.Gauges

	var wg sync.WaitGroup
	for i := 1; i <= s.numWorkers; i++ {
		wg.Add(1)
		go s.gaugePageWorker(listCh, resultsCh, &wg)
	}
	go func() {
		defer close(listCh)
		err = s.parseMeasurements(listCh)
	}()
	if err != nil {
		close(resultsCh)
		return nil, err
	}
	go func() {
		wg.Wait()
		close(resultsCh)
	}()
	for g := range resultsCh {
		results = append(results, *g)
	}
	return results, nil
}

func (s *scriptWaikato) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	err := s.parseMeasurements(recv)
	if err != nil {
		errs <- err
	}
}
