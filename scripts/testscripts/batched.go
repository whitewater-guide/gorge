package testscripts

import (
	"context"
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

// BatchedOptions is public for global test purposes
type BatchedOptions struct {
	Gauges    int     `desc:"Number of gauges"`
	Value     float64 `desc:"Set this to return fixed value. Has priority over min/max"`
	Min       float64 `desc:"Set this and max to return random values within interval"`
	Max       float64 `desc:"Set this and min to return random values within interval"`
	BatchSize int     `desc:"Number of gauges in batch"`
}

// GetBatchSize Implements core.BatchableOptions interface
func (o BatchedOptions) GetBatchSize() int {
	return o.BatchSize
}

type scriptBatched struct {
	name    string
	options BatchedOptions
	core.LoggingScript
}

func (s *scriptBatched) ListGauges() (core.Gauges, error) {
	res := make([]core.Gauge, s.options.Gauges)
	for i := 0; i < s.options.Gauges; i++ {
		res[i] = core.GenerateRandGauge(s.name, i)
	}
	return res, nil
}

func (s *scriptBatched) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	for code := range codes {
		m := core.GenerateRandMeasurement(s.name, code, s.options.Value, s.options.Min, s.options.Max)
		select {
		case recv <- &m:
		case <-ctx.Done():
			return
		}
	}
}

var Batched = &core.ScriptDescriptor{
	Name:        "batched",
	Description: "Test script for batched harvesting mode",
	Mode:        core.Batched,
	DefaultOptions: func() interface{} {
		return &BatchedOptions{
			Gauges:    10,
			Value:     0,
			Min:       10,
			Max:       20,
			BatchSize: 3,
		}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*BatchedOptions); ok {
			return &scriptBatched{name: name, options: *opts}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", BatchedOptions{})
	},
}
