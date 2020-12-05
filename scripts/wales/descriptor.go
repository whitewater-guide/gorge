package wales

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "wales",
	Description: "Natural Resources Wales",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsWales{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsWales); ok {
			return &scriptWales{
				name:    name,
				url:     "https://api.naturalresources.wales/rivers-and-seas/v1/api/StationData",
				options: *opts,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsWales{})
	},
}
