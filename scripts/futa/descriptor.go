package futa

import (
	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "futa",
	Description: "Central Hidroeléctrica Futaleufú",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsFuta{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		return &scriptFuta{name: name, dataURL: "https://www.chfutaleufu.com.ar/serv/bdatos/hoyweb.txt"}, nil
	},
}
