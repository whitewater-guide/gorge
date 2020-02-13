package cantabria

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
)

type optionsCantabria struct{}

type scriptCantabria struct {
	name         string
	listURL      string
	gaugeURLBase string
	core.LoggingScript
}

func (s *scriptCantabria) ListGauges() (core.Gauges, error) {
	resCh, errCh, err := s.parseTable()
	if err != nil {
		return nil, err
	}
	var result []*tableEntry
outer:
	for {
		select {
		case err = <-errCh:
			if err != nil {
				return nil, err
			}
		case entry, ok := <-resCh:
			if !ok {
				break outer
			}
			result = append(result, entry)
		}
	}

	jobsCh := make(chan *core.Gauge, len(result))
	resultsCh := make(chan *core.Gauge, len(result))
	gauges := make([]core.Gauge, len(result))

	for w := 1; w <= 10; w++ {
		go s.gaugePageWorker(jobsCh, resultsCh)
	}
	for i := range result {
		jobsCh <- result[i].gauge
	}
	close(jobsCh)
	for range gauges {
		<-resultsCh
	}
	for i := range result {
		gauges[i] = *result[i].gauge
	}
	close(resultsCh)
	return gauges, nil
}

func (s *scriptCantabria) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	resCh, errCh, err := s.parseTable()
	if err != nil {
		errs <- err
		return
	}
	select {
	case <-ctx.Done():
		return
	case err = <-errCh:
		if err != nil {
			errs <- err
			return
		}
	default:
		for i := range resCh {
			recv <- i.measurement
		}
	}

}

func (s *scriptCantabria) gaugePageWorker(gauges <-chan *core.Gauge, results chan<- *core.Gauge) {
	for gauge := range gauges {
		loc := s.parseGaugeLocation((*gauge).Code)
		(*gauge).Location = &loc
		results <- gauge
	}
}
