package kuban

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsKuban struct{}
type scriptKuban struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptKuban) ListGauges() (core.Gauges, error) {
	gaugesCh := make(chan *core.Gauge)
	errCh := make(chan error)
	go func() {
		defer close(gaugesCh)
		defer close(errCh)
		s.parseTable(gaugesCh, nil, errCh)
	}()
	return core.GaugeSinkToSlice(gaugesCh, errCh)
}

func (s *scriptKuban) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.parseTable(nil, recv, errs)
}
