package nzwgn

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "nzwgn",
	Description: "New Zealand: Greater Wellington Regional Council",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsNzwgn{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsNzwgn); ok {
			return &scriptNzwgn{
				name: name,
				url:  "http://hilltop.gw.govt.nz/Data.hts?Service=Hilltop",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNzwgn{})
	},
}
