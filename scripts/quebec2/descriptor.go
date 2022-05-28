package quebec2

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "quebec2",
	Description: "Hydro Québec",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsQuebec2{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsQuebec2); ok {
			return &scriptQuebec2{
				name:    name,
				urlBase: "https://www.hydroquebec.com/data/documents-donnees/donnees-ouvertes/json/",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsQuebec2{})
	},
}
