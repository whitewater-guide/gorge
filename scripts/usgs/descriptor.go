package usgs

import (
	"errors"
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "usgs",
	Description: "U.S. Geological Survey National Water Information System",
	Mode:        core.Batched,
	DefaultOptions: func() interface{} {
		return &optionsUSGS{
			BatchSize: 200, // has to be tuned in production not to hit URL limits
		}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsUSGS); ok {
			if opts.StateCD == "" {
				return nil, errors.New("state code must be provided")
			}
			return &scriptUSGS{
				name:    name,
				url:     "https://waterservices.usgs.gov/nwis",
				stateCd: opts.StateCD,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsUSGS{})
	},
}
