package nztrc

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNztrc struct{}
type scriptNztrc struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptNztrc) ListGauges() (core.Gauges, error) {
	gaugesCh := make(chan *core.Gauge)
	errCh := make(chan error)
	go func() {
		defer close(gaugesCh)
		defer close(errCh)
		s.parseList(gaugesCh, nil, errCh)
	}()
	return core.GaugeSinkToSlice(gaugesCh, errCh)
}

func (s *scriptNztrc) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.parseList(nil, recv, errs)
}
