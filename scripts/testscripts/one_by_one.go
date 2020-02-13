package testscripts

import (
	"context"
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

// OneByOneOptions is public for global test purposes
type OneByOneOptions struct {
	Gauges int     `desc:"Number of gauges"`
	Value  float64 `desc:"Set this to return fixed value. Has priority over min/max"`
	Min    float64 `desc:"Set this and max to return random values within interval"`
	Max    float64 `desc:"Set this and min to return random values within interval"`
}

type scriptOneByOne struct {
	name    string
	options OneByOneOptions
	core.LoggingScript
}

func (s *scriptOneByOne) ListGauges() (core.Gauges, error) {
	res := make([]core.Gauge, s.options.Gauges)
	for i := 0; i < s.options.Gauges; i++ {
		res[i] = core.GenerateRandGauge(s.name, i)
	}
	return res, nil
}

func (s *scriptOneByOne) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	code, err := codes.Only()
	if err != nil {
		errs <- err
		return
	}

	m := core.GenerateRandMeasurement(s.name, code, s.options.Value, s.options.Min, s.options.Max)
	select {
	case recv <- &m:
	case <-ctx.Done():
		return
	}
}

var OneByOne = &core.ScriptDescriptor{
	Name: "one_by_one",
	Mode: core.OneByOne,
	DefaultOptions: func() interface{} {
		return &OneByOneOptions{
			Gauges: 10,
			Value:  0,
			Min:    10,
			Max:    20,
		}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*OneByOneOptions); ok {
			return &scriptOneByOne{name: name, options: *opts}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", OneByOneOptions{})
	},
}
