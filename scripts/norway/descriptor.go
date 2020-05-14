package norway

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "norway",
	Description: "Norwegian Water Resources and Energy Directorate",
	Mode:        core.OneByOne,
	DefaultOptions: func() interface{} {
		return &optionsNorway{
			HTML:    false,
			Version: 1,
		}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsNorway); ok {
			return &scriptNorway{
				name:    name,
				urlBase: "http://www2.nve.no/h/hd/plotreal/Q",
				options: *opts,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNorway{})
	},
}
