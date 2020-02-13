package ecuador

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsEcuador struct{}
type scriptEcuador struct {
	name           string
	listURL1       string
	listURL2       string
	gaugeURLFormat string
	core.LoggingScript
}

func (s *scriptEcuador) ListGauges() (core.Gauges, error) {
	// this list has no coordinates
	list1, err := s.parseList()
	if err != nil {
		return nil, err
	}
	// this list has coordinates
	list2, err := s.parseList2()
	if err != nil {
		return nil, err
	}
	merged := make(map[string]core.Gauge)
	for _, v := range list1 {
		merged[v.Code] = v
	}
	for _, v := range list2 {
		merged[v.Code] = v
	}
	result, i := make([]core.Gauge, len(merged)), 0
	for _, v := range merged {
		result[i] = v
		i++
	}
	return result, nil
}

func (s *scriptEcuador) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	code, err := codes.Only()
	if err != nil {
		errs <- err
		return
	}
	s.parseGauge(recv, errs, code)
}
