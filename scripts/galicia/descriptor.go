package galicia

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "galicia",
	Description: "Spain: MeteoGalicia",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsGalicia{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if opts, ok := options.(*optionsGalicia); ok {
			return &scriptGalicia{
				name: name,
				url:  fmt.Sprintf("https://servizos.meteogalicia.gal/mgafos/json/estadoActual.action?cod=%s", opts.Code),
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsGalicia{})
	},
}
