package russia1

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type scriptRussia1 struct {
	name      string
	gaugesURL string
	core.LoggingScript
}

type optionsRussia1 struct{}

func (s *scriptRussia1) ListGauges() (core.Gauges, error) {
	gaugesCh := make(chan *core.Gauge)
	errCh := make(chan error)
	go func() {
		defer close(gaugesCh)
		defer close(errCh)
		s.fetchList(gaugesCh, nil, errCh)
	}()
	var gauges core.Gauges
outer:
	for {
		select {
		case err := <-errCh:
			if err != nil {
				return nil, err
			}
		case g, ok := <-gaugesCh:
			if g != nil {
				gauges = append(gauges, *g)
			}
			if !ok {
				break outer
			}
		}
	}
	return gauges, nil
}

func (s *scriptRussia1) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.fetchList(nil, recv, errs)
}
