package nzcan

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "nzcan",
	Description: "New Zealand: Environment Canterbury",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsNzcan{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsNzcan); ok {
			return &scriptNzcan{
				name: name,
				url:  "https://ecan.govt.nz/data/riverflow",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNzcan{})
	},
}
