package testscripts

import (
	"context"
	"errors"

	"github.com/whitewater-guide/gorge/core"
)

type optionsBroken struct{}

type scriptBroken struct {
	core.LoggingScript
}

func (s *scriptBroken) ListGauges() (core.Gauges, error) {
	return nil, errors.New("this script is always broken")
}

func (s *scriptBroken) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	select {
	case errs <- errors.New("this script is always broken"):
	case <-ctx.Done():
		return
	}
}

var Broken = &core.ScriptDescriptor{
	Name: "broken",
	Mode: core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsBroken{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		return &scriptBroken{}, nil
	},
}
