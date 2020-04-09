package ukraine

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "ukraine",
	Mode: core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsUkraine{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsUkraine); ok {
			return &scriptUkraine{
				name: name,
				url:  "http://meteo.gov.ua/kml/kml_hydro_warn.kml",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsUkraine{})
	},
}
