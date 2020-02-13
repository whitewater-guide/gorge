package catalunya

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type scriptCatalunya struct {
	name            string
	gaugesURL       string
	measurementsURL string
	core.LoggingScript
}

type optionsCatalunya struct{}

func (s *scriptCatalunya) ListGauges() (core.Gauges, error) {
	return s.parseList()
}

func (s *scriptCatalunya) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	s.parseObservations(recv, errs)
}
