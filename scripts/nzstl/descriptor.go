package nzstl

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "nzstl",
	Description: "New Zealand: Environment Southland",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsNzstl{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsNzstl); ok {
			return &scriptNzstl{
				name: name,
				url:  "http://envdata.es.govt.nz/services/data.ashx?f=water-level.xml",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNzstl{})
	},
}
