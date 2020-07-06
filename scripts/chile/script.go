package chile

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsChile struct{}

type scriptChile struct {
	name            string
	selectFormURL   string
	webmapIDPageURL string
	webmapURLFormat string
	xlsURL          string
	skipCookies     bool
	core.LoggingScript
}

func (s *scriptChile) ListGauges() (core.Gauges, error) {
	fromKml, err := s.getKMLGauges()
	if err != nil {
		return nil, err
	}
	fromWeb, err := s.getWebGauges()
	if err != nil {
		return nil, err
	}

	result := core.Gauges{}

	for _, v := range fromKml {
		result = append(result, v)
	}
	for k, v := range fromWeb {
		if _, ok := fromKml[k]; !ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (s *scriptChile) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	code, err := codes.Only()
	if err != nil {
		errs <- err
		return
	}
	s.parseXLS(recv, errs, code, since)
}
