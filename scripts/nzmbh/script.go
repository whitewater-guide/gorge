package nzmbh

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNzmbh struct{}
type scriptNzmbh struct {
	name        string
	reportURL   string
	siteListURL string
	core.LoggingScript
}

func (s *scriptNzmbh) ListGauges() (core.Gauges, error) {
	sites, err := s.fetchSiteList()
	if err != nil {
		return nil, err
	}
	gauges := core.Gauges{}

	msmntsCh := make(chan *core.Measurement)
	errCh := make(chan error)
	go func() {
		defer close(msmntsCh)
		defer close(errCh)
		s.fetchReport(msmntsCh, errCh)
	}()

outer:
	for {
		select {
		case err := <-errCh:
			if err != nil {
				return nil, err
			}
		case m, ok := <-msmntsCh:
			if !ok {
				break outer
			}
			g := s.genGauge(sites, m)
			if g != nil {
				gauges = append(gauges, *g)
			}
		}

	}

	return gauges, nil
}

func (s *scriptNzmbh) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.fetchReport(recv, errs)
}
