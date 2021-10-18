package canada

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/tz"
	"golang.org/x/text/encoding/charmap"
)

type optionsCanada struct {
	Provinces string `desc:"Comma-separated list of province codes"`
}

type scriptCanada struct {
	name      string
	baseURL   string
	numWokers int
	provinces map[string]bool
	core.LoggingScript
}

func (s *scriptCanada) ListGauges() (result core.Gauges, err error) {
	err = core.Client.StreamCSV(
		s.baseURL+"/doc/hydrometric_StationList.csv",
		func(row []string) error {
			g, err := s.gaugeFromRow(row)
			if err != nil {
				s.GetLogger().Error(fmt.Errorf("failed to convert row to gauge: %w", err))
			}
			if _, ok := s.provinces[row[4]]; err == nil && ok {
				result = append(result, *g)
			}
			return nil
		},
		core.CSVStreamOptions{
			NumColumns:   6,
			HeaderHeight: 1,
			Decoder:      charmap.Windows1252.NewDecoder(),
		},
	)
	tz.CloseTimezoneDb()
	return
}

func (s *scriptCanada) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	// Sometimes gauge list would contain main gauge, but measurements list would contain auxiliary gauge for it or vise versa.
	// Auxiliary gauges are marked with X character, main gauges have 0. Some gauges have 1, I dont't know what it means
	// I could not find any documentation on this, so I know this by trial and error
	remapCodes := make(map[string]string)
	if len(codes) > 0 {
		for code := range codes {
			code2 := getPairedGauge(code)
			if _, ok := codes[code2]; !ok && code2 != code {
				remapCodes[code2] = code
			}
		}
	}

	jobsCh := make(chan string, len(s.provinces))
	resultsCh := make(chan *core.Measurement, len(s.provinces))

	var wg sync.WaitGroup
	numWorkers := s.numWokers
	if numWorkers == 0 {
		numWorkers = int(math.Min(3, float64(len(s.provinces))))
	}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go s.provinceWorker(ctx, jobsCh, resultsCh, &wg)
	}
	for prov := range s.provinces {
		jobsCh <- prov
	}
	close(jobsCh)
	go func() {
		wg.Wait()
		close(resultsCh)
	}()
	for m := range resultsCh {
		if code2, ok := remapCodes[m.GaugeID.Code]; ok {
			m.GaugeID.Code = code2
		}
		recv <- m
	}
}

func (s *scriptCanada) provinceWorker(ctx context.Context, provinces <-chan string, results chan<- *core.Measurement, wg *sync.WaitGroup) {
	for province := range provinces {
		logger := s.GetLogger().WithField("province", province)

		var m *core.Measurement
		err := core.Client.StreamCSV(
			fmt.Sprintf("%s/csv/%s/hourly/%s_hourly_hydrometric.csv", s.baseURL, province, province),
			func(row []string) error {
				select {
				case <-ctx.Done():
					return fmt.Errorf("context canceled: %w", ctx.Err())
				default:
					var e error
					m, e = s.measurementFromRow(row)
					if e == nil {
						results <- m
					}
					if e != nil {
						logger.Error(core.WrapErr(e, "province worker error"))
					}
				}

				return nil
			},
			core.CSVStreamOptions{
				NumColumns:   10,
				HeaderHeight: 1,
				Decoder:      charmap.Windows1252.NewDecoder(),
			},
		)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			break
		}
	}
	wg.Done()
}
