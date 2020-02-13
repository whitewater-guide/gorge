package switzerland

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsSwitzerland struct {
	Username string `desc:"Username for basic HTTP auth"`
	Password string `desc:"Password for basic HTTP auth"`
}

type scriptSwitzerland struct {
	name             string
	options          optionsSwitzerland
	xmlURL           string
	gaugePageURLBase string
	core.LoggingScript
}

func (s *scriptSwitzerland) ListGauges() (core.Gauges, error) {
	gauges, err := s.parseXMLGauges()
	if err != nil {
		return nil, err
	}

	// Workers just parse elevation
	numGauges := len(gauges)
	numWorkers := 10
	jobsCh := make(chan *core.Gauge, numGauges)
	resultsCh := make(chan struct{}, numGauges)

	for w := 1; w <= numWorkers; w++ {
		go gaugePageWorker(s.gaugePageURLBase, jobsCh, resultsCh)
	}
	for i := 0; i < numGauges; i++ {
		jobsCh <- &(gauges[i])
	}
	close(jobsCh)
	for i := 0; i < numGauges; i++ {
		<-resultsCh
	}
	close(resultsCh)
	return gauges, nil
}

func (s *scriptSwitzerland) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	dataRoot, err := s.fetchStations()
	if err != nil {
		errs <- err
		return
	}
	for _, station := range dataRoot.Stations {
		m := s.stationToMeasurement(&station)
		if m != nil {
			recv <- m
		}
	}
}
