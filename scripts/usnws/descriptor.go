package usnws

import (
	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "usnws",
	Description: "U.S. National Oceanic and Atmospheric Administration's National Weather Service",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsUsnws{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		return &scriptUsnws{name: name, kmzUrl: "https://water.weather.gov/ahps/download.php?data=kmz_obs"}, nil
	},
}
