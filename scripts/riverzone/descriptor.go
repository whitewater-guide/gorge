package riverzone

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "riverzone",
	Description: "riverzone.eu",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsRiverzone{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsRiverzone); ok {
			return &scriptRiverzone{
				name:                name,
				options:             *opts,
				stationsEndpointURL: "https://api.riverzone.eu/v1/stations",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsRiverzone{})
	},
}
