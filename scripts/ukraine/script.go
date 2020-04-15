package ukraine

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsUkraine struct{}
type scriptUkraine struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptUkraine) ListGauges() (core.Gauges, error) {
	gaugesCh := make(chan *core.Gauge)
	errCh := make(chan error)
	go func() {
		defer close(gaugesCh)
		defer close(errCh)
		s.parseTable(gaugesCh, nil, errCh)
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

func (s *scriptUkraine) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.parseTable(nil, recv, errs)
}
