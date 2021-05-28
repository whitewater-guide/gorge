package ukraine

import (
	"fmt"
	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "ukraine",
	Description: "Ukrainian Hydrometeorological Center",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsUkraine{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsUkraine); ok {
			return &scriptUkraine{
				name:         name,
				urlDaily:     "http://meteo.gov.ua/kml",
				urlHourly:    "http://hydro.meteo.gov.ua",
				timezone:     getTimezone(),
				station2code: station2code,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsUkraine{})
	},
}
