package norway

import (
	"fmt"
	"os"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "norway",
	Description: "Norwegian Water Resources and Energy Directorate",
	Mode:        core.Batched,
	DefaultOptions: func() interface{} {
		return &optionsNorway{
			ApiKey:    os.Getenv("NVE_API_KEY"),
			BatchSize: 5, // determined experiementally
		}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsNorway); ok {
			return &scriptNorway{
				name:         name,
				urlBase:      "https://hydapi.nve.no/api/v1",
				apiKey:       opts.ApiKey,
				ignoreLegacy: opts.IgnoreLegacy,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNorway{})
	},
}
