package nz_bop

import (
	"context"
	"sync"

	"github.com/whitewater-guide/gorge/core"
)

type optionsBop struct{}
type scriptBop struct {
	name       string
	listURL    string
	pageURL    string
	numWorkers int
	core.LoggingScript
}

func (s *scriptBop) ListGauges() (core.Gauges, error) {
	codes, err := s.parseList()
	if err != nil {
		return nil, err
	}
	resultsCh := make(chan *core.Gauge)
	codesCh := make(chan string)
	var results core.Gauges

	var wg sync.WaitGroup
	for i := 1; i <= s.numWorkers; i++ {
		wg.Add(1)
		go s.gaugePageWorker(codesCh, resultsCh, &wg)
	}
	go func() {
		for _, code := range codes {
			codesCh <- code
		}
		close(codesCh)
	}()
	go func() {
		wg.Wait()
		close(resultsCh)
	}()
	for g := range resultsCh {
		results = append(results, *g)
	}
	return results, nil
}

func (s *scriptBop) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	code, err := codes.Only()
	if err != nil {
		errs <- err
		return
	}
	s.parsePage(code, nil, recv)
}
