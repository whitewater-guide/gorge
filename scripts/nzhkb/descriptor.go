package nzhkb

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "nzhkb",
	Description: "New Zealand: Hawke's Bay Regional Council",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsNzhkb{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsNzhkb); ok {
			return &scriptNzhkb{
				name: name,
				url:  "https://hbmaps.hbrc.govt.nz/arcgis/rest/services/WebMaps/Environmental1/MapServer",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNzhkb{})
	},
}
