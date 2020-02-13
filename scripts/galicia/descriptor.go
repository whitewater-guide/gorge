package galicia

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "galicia",
	Mode: core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsGalicia{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsGalicia); ok {
			return &scriptGalicia{
				name: name,
				url:  "http://servizos.meteogalicia.es/rss/observacion/jsonAforos.action",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsGalicia{})
	},
}
