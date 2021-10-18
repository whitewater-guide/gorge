package quebec

import (
	"context"
	"math"

	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
)

type optionsQuebec struct{}
type scriptQuebec struct {
	name              string
	codesURL          string
	referenceListURL  string
	stationURLFormat  string
	readingsURLFormat string
	core.LoggingScript
}

func (s *scriptQuebec) ListGauges() (result core.Gauges, err error) {
	federal, err := s.getReferenceList()
	var local []stationInfo
	if err != nil {
		return nil, err
	}
	codes, err := s.getCodes()
	if err != nil {
		return nil, err
	}

	jobsCh := make(chan string, len(codes))
	resultsCh := make(chan stationInfo, len(codes))
	numWorkers := int(math.Min(5, float64(len(codes))))
	for i := 0; i < numWorkers; i++ {
		go s.stationWorker(jobsCh, resultsCh)
	}
	for _, code := range codes {
		jobsCh <- code
	}
	close(jobsCh)
	for range codes {
		gauge := <-resultsCh
		if gauge.isLocal {
			local = append(local, gauge)
		}
	}
	close(resultsCh)

	for _, gauge := range local {
		var gauges core.Gauge
		if fedInfo, ok := federal[gauge.federalCode]; ok {
			gauges = fedInfo
			gauges.Name = gauge.name + " (" + gauge.code + ")"
			gauges.Code = gauge.code
			gauges.URL = "https://www.cehq.gouv.qc.ca/suivihydro/graphique.asp?NoStation=" + gauge.code
		} else {
			s.GetLogger().WithFields(logrus.Fields{
				"script":      s.name,
				"command":     "harvest",
				"federalCode": gauge.federalCode,
				"code":        gauge.code,
			}).Warn("no federal reference found")
			gauges = core.Gauge{
				GaugeID: core.GaugeID{
					Code:   gauge.code,
					Script: s.name,
				},
				Name:      gauge.name + " (" + gauge.code + ")",
				FlowUnit:  "m3/s",
				LevelUnit: "m",
				URL:       "https://www.cehq.gouv.qc.ca/suivihydro/graphique.asp?NoStation=" + gauge.code,
				Timezone:  "America/Toronto",
			}
		}
		result = append(result, gauges)
	}
	return
}

func (s *scriptQuebec) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	code, err := codes.Only()
	if err != nil {
		errs <- err
		return
	}
	s.getReadings(recv, errs, code)
}

func (s *scriptQuebec) stationWorker(codes <-chan string, results chan<- stationInfo) {
	for code := range codes {
		gauge, err := s.parsePage(code)
		if err != nil {
			s.GetLogger().WithFields(logrus.Fields{
				"script":  s.name,
				"command": "harvest",
				"code":    code,
			}).Error(err)
		}
		results <- *gauge
	}
}
