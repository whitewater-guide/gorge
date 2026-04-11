package quebec2

import (
	"fmt"
	"os"

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
		if opts, ok := options.(*optionsQuebec2); ok {
			urlBase := "https://www.hydroquebec.com/data/documents-donnees/donnees-ouvertes/json/"
			if opts.urlBase != "" {
				urlBase = os.ExpandEnv(opts.urlBase)
			}
			return &scriptQuebec2{
				name:    name,
				urlBase: urlBase,
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsQuebec2{})
	},
}
