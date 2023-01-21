package ireland

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "ireland",
	Description: "https://waterlevel.ie/",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsIreland{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsIreland); ok {
			return &scriptIreland{
				name: name,
				url:  "http://waterlevel.ie/geojson/latest/",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsIreland{})
	},
}
