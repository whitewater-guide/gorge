package nzstl

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNzstl struct{}
type scriptNzstl struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptNzstl) ListGauges() (core.Gauges, error) {
	gaugesCh := make(chan *core.Gauge)
	errCh := make(chan error)
	go func() {
		defer close(gaugesCh)
		defer close(errCh)
		s.fetchList(gaugesCh, nil, errCh)
	}()
	return core.GaugeSinkToSlice(gaugesCh, errCh)
}

func (s *scriptNzstl) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.fetchList(nil, recv, errs)
}
