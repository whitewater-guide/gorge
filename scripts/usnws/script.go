package usnws

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsUsnws struct{}
type scriptUsnws struct {
	core.LoggingScript
	name   string
	kmzUrl string
}

func (s *scriptUsnws) ListGauges() (core.Gauges, error) {
	gaugesCh := make(chan *core.Gauge)
	errCh := make(chan error)
	go func() {
		defer close(gaugesCh)
		defer close(errCh)
		s.parseKmz(gaugesCh, nil, errCh)
	}()
	return core.GaugeSinkToSlice(gaugesCh, errCh)
}

func (s *scriptUsnws) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.parseKmz(nil, recv, errs)
}
