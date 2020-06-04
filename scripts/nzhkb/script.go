package nzhkb

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsNzhkb struct{}
type scriptNzhkb struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptNzhkb) ListGauges() (core.Gauges, error) {
	return s.fetchGauges()
}

func (s *scriptNzhkb) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(errs)
	defer close(recv)
	s.fetchMeasurements(recv, errs)
}
