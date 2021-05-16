package switzerland

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "switzerland",
	Description: "Switzerland: Federal Office for the Environment",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsSwitzerland{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsSwitzerland); ok {
			return &scriptSwitzerland{
				name:             name,
				xmlURL:           "https://www.hydrodaten.admin.ch/lhg/az/xml/hydroweb.xml",
				gaugePageURLBase: "https://www.hydrodaten.admin.ch/en/",
				options:          *opts,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsSwitzerland{})
	},
}
