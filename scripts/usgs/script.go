package usgs

import (
	"context"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

type optionsUSGS struct {
	StateCD   string `desc:"State code"`
	BatchSize int    `desc:"Number of gauges in batch"`
}

// GetBatchSize Implements core.BatchableOptions interface
func (o optionsUSGS) GetBatchSize() int {
	return o.BatchSize
}

type scriptUSGS struct {
	name    string
	url     string
	stateCd string
	core.LoggingScript
}

func (s *scriptUSGS) ListGauges() (core.Gauges, error) {
	gMap := map[string]core.Gauge{}
	// fetch twice to correctly set level and flow units
	err := s.listStations(false, gMap)
	if err != nil {
		return nil, err
	}
	err = s.listStations(true, gMap)
	if err != nil {
		return nil, err
	}
	result := make([]core.Gauge, len(gMap))
	i := 0
	for _, v := range gMap {
		result[i] = v
		i++
	}
	core.CloseTimezoneDb()
	return result, nil
}

func (s *scriptUSGS) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	codez := []string{}
	// send in chunks of 100
	for code := range codes {
		codez = append(codez, code)
		if len(codez) >= 100 {
			s.listInstantaneousValues(strings.Join(codez, ","), recv, errs)
			codez = []string{}
		}
	}
	if len(codez) >= 0 {
		s.listInstantaneousValues(strings.Join(codez, ","), recv, errs)
	}
}
