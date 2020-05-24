package nzwko

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "nzwko",
	Description: "New Zealand: Waikato Regional Council",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsWaikato{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsWaikato); ok {
			return &scriptWaikato{
				name:       name,
				numWorkers: 5,
				listURL:    "https://www.waikatoregion.govt.nz/services/regional-services/river-levels-and-rainfall/river-levels-and-flow-latest-reading/",
				pageURL:    "http://riverlevelsmap.waikatoregion.govt.nz/cgi-bin/hydwebserver.cgi/points/details?point=%s&catchment=17",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsWaikato{})
	},
}
