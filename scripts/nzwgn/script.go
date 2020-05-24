package nzwgn

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNzwgn struct{}
type scriptNzwgn struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptNzwgn) ListGauges() (core.Gauges, error) {
	return nil, nil
}

func (s *scriptNzwgn) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	data, err := s.fetchValues()
	if err != nil {
		errs <- err
		return
	}
	for _, v := range data {
		recv <- &v
	}
}
