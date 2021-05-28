package ukraine

import (
	"context"
	"time"

	"github.com/whitewater-guide/gorge/core"
)

type optionsUkraine struct{}
type scriptUkraine struct {
	core.LoggingScript
	name           string
	urlDaily       string
	urlHourly      string
	addStation2url bool
	timezone       *time.Location
	station2code   map[string]string
}

func (s *scriptUkraine) ListGauges() (core.Gauges, error) {
	rivers, err := s.getAllRivers()
	if err != nil {
		return nil, err
	}
	gauges := make(core.Gauges, 0, len(rivers))
	for _, river := range rivers {
		gauges = append(gauges, river.Gauge)
	}
	return gauges, nil
}

func (s *scriptUkraine) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.harvest(recv, errs)
}
