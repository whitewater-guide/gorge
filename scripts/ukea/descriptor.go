package ukea

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "ukea",
	Description: "UK Environment Agency",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsUkea{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsUkea); ok {
			return &scriptUkea{
				rloi: opts.RLOI,
				name: name,
				url:  "http://environment.data.gov.uk/flood-monitoring",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsUkea{})
	},
}
