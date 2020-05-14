package tirol

import (
	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "tirol",
	Description: "Tyrol Hydro Online",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsTirol{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		return &scriptTirol{name: name, csvURL: "https://apps.tirol.gv.at/hydro/ogd/OGD_W.csv"}, nil
	},
}
