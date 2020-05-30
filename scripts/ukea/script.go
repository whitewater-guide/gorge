package ukea

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsUkea struct {
	RLOI int `desc:"0 to include all stations, 1 to only include stations with RLOI, 2 to only include stations without RLOI"`
}
type scriptUkea struct {
	name string
	url  string
	rloi int
	core.LoggingScript
}

func (s *scriptUkea) ListGauges() (core.Gauges, error) {
	return s.fetchList()
}

func (s *scriptUkea) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.getReadings(recv, errs)
}
