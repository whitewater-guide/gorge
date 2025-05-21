package catalunya

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "catalunya",
	Description: "Catalan Water Agency",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsCatalunya{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsCatalunya); ok {
			return &scriptCatalunya{
				name: name,
				// gaugesURL:       "http://aca-web.gencat.cat/sdim2/apirest/catalog?componentType=aforament",
				gaugesURL: "https://aplicacions.aca.gencat.cat/sdim2/apirest/catalog?componentType=aforament",
				// measurementsURL: "http://aca-web.gencat.cat/sdim2/apirest/data/AFORAMENT-EST",
				measurementsURL: "https://aplicacions.aca.gencat.cat/sdim2/apirest/data/AFORAMENT-EST",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsCatalunya{})
	},
}
