package wales

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsWales struct{
	Key string `desc:"Auth key"`
}

type scriptWales struct {
	name string
	url  string
	options             optionsWales
	core.LoggingScript
}

func (s *scriptWales) ListGauges() (core.Gauges, error) {
	gaugesCh := make(chan *core.Gauge)
	errCh := make(chan error)
	go func() {
		defer close(gaugesCh)
		defer close(errCh)
		s.fetchList(gaugesCh, nil, errCh)
	}()
	return core.GaugeSinkToSlice(gaugesCh, errCh)
}

func (s *scriptWales) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.fetchList(nil, recv, errs)
}
