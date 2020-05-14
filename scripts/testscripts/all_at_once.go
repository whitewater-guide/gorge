package testscripts

import (
	"context"
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

type optionsAllAtOnce struct {
	Gauges     int     `desc:"Number of gauges"`
	Value      float64 `desc:"Set this to return fixed value. Has priority over min/max"`
	Min        float64 `desc:"Set this and max to return random values within interval"`
	Max        float64 `desc:"Set this and min to return random values within interval"`
	NoLocation bool    `desc:"Generate gauges without locations" json:"noLocation"`
	NoAltitude bool    `desc:"Generate gauges with 0 altitude" json:"noAltitude"`
}

type scriptAllAtOnce struct {
	name    string
	options optionsAllAtOnce
	core.LoggingScript
}

func (s *scriptAllAtOnce) ListGauges() (core.Gauges, error) {
	res := make([]core.Gauge, s.options.Gauges)
	for i := 0; i < s.options.Gauges; i++ {
		res[i] = core.GenerateRandGauge(s.name, i)
		if s.options.NoLocation {
			res[i].Location = nil
		}
		if s.options.NoAltitude {
			res[i].Location.Altitude = 0
		}
	}
	return res, nil
}

func (s *scriptAllAtOnce) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	for i := 0; i < s.options.Gauges; i++ {
		m := core.GenerateRandMeasurement(s.name, fmt.Sprintf("g%03d", i), s.options.Value, s.options.Min, s.options.Max)
		select {
		case recv <- &m:
		case <-ctx.Done():
			return
		}
	}
}

var AllAtOnce = &core.ScriptDescriptor{
	Name:        "all_at_once",
	Description: "Test script for all at once harvesting mode",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsAllAtOnce{
			Gauges:     10,
			Value:      0,
			Min:        10,
			Max:        20,
			NoLocation: false,
		}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsAllAtOnce); ok {
			return &scriptAllAtOnce{name: name, options: *opts}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsAllAtOnce{})
	},
}
