package canada

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "canada",
	Description: "Environment and Climate Change Canada",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsCanada{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsCanada); ok {
			return &scriptCanada{
				name:       name,
				baseURL:    "https://dd.weather.gc.ca/hydrometric",
				provinces:  getProvinces(opts.Provinces),
				timeoutSec: opts.Timeout,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsCanada{})
	},
}
