package usnws

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "usnws",
	Description: "U.S. National Oceanic and Atmospheric Administration's National Weather Service",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsUsnws{
			pageSize:   100, // defaults to 5000 if not mentioned at all, total around 10500
			numWorkers: 5,
		}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsUsnws); ok {
			return &scriptUsnws{
				name:       name,
				url:        "https://mapservices.weather.noaa.gov/eventdriven/rest/services/water/ahps_riv_gauges/MapServer/0/query",
				pageSize:   opts.pageSize,
				numWorkers: opts.numWorkers,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsUsnws{})
	},
}
