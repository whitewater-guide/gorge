package tirol

import (
	"context"

	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/text/encoding/charmap"
)

type optionsTirol struct{}

type scriptTirol struct {
	name   string
	csvURL string
	core.LoggingScript
}

var csvOptions = core.CSVStreamOptions{
	Comma:        ';',
	Decoder:      charmap.Windows1252.NewDecoder(),
	NumColumns:   11,
	HeaderHeight: 1,
}

func (s *scriptTirol) ListGauges() (result core.Gauges, err error) {
	byCode := make(map[string]core.Gauge)
	err = core.Client.StreamCSV(
		s.csvURL,
		func(row []string) error {
			raw := fromRow(row)
			_, ok := byCode[raw.code]
			if ok {
				return nil
			}
			g, err := s.getGauge(raw)
			if err != nil {
				return err
			}
			byCode[raw.code] = g
			return nil
		},
		csvOptions,
	)
	if err != nil {
		return
	}
	result = make([]core.Gauge, len(byCode))
	i := 0
	for _, v := range byCode {
		result[i] = v
		i++
	}
	return
}

func (s *scriptTirol) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	err := core.Client.StreamCSV(
		s.csvURL,
		func(row []string) error {
			m, err := s.getMeasurement(fromRow(row))
			if err != nil {
				s.GetLogger().Error(err)
				return nil
			}
			_, ok := codes[m.Code]
			// special value that indicates broken gauge
			if (ok || len(codes) == 0) && m.Level.Float64Value() != -777.0 {
				recv <- &m
			}
			return nil
		},
		csvOptions,
	)
	if err != nil {
		errs <- err
	}
}
