package uscdec

import (
	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "uscdec",
	Description: "U.S. California Data Exchange Center",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsUSCDEC{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		return &scriptUSCDEC{name: name, url: "https://cdec.water.ca.gov/dynamicapp"}, nil
	},
}
