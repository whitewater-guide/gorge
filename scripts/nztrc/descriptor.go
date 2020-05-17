package nztrc

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "nztrc",
	Description: "New Zealand: Taranaki Regional Council",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsNztrc{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsNztrc); ok {
			return &scriptNztrc{
				name: name,
				url:  "https://www.trc.govt.nz/environment/maps-and-data/regional-overview/MapMarkers/?measureID=",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNztrc{})
	},
}
