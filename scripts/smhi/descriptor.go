package smhi

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "smhi",
	Description: "https://www.smhi.se",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsSmhi{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsSmhi); ok {
			return &scriptSmhi{
				name: name,
				url:  "https://opendata-download-hydroobs.smhi.se",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsSmhi{})
	},
}
