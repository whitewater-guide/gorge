package galicia

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsGalicia struct{}
type scriptGalicia struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptGalicia) ListGauges() (core.Gauges, error) {
	l, err := s.fetchList()
	if err != nil {
		return nil, err
	}
	result := make([]core.Gauge, len(l))
	for i, item := range l {
		result[i] = *item.gauge
	}
	return result, nil
}

func (s *scriptGalicia) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	list, err := s.fetchList()
	if err != nil {
		errs <- err
		return
	}
	for _, i := range list {
		recv <- i.measurement
	}
}
