package sepa

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "sepa",
	Description: "Scottish Environment Protection Agency",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsSepa{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsSepa); ok {
			return &scriptSepa{
				name: name,
				// This is old datasource list, maybe it's not working anymore
				listURL: "https://www2.sepa.org.uk/waterlevels/CSVs",
				apiURL:  "https://timeseries.sepa.org.uk/KiWIS/KiWIS",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsSepa{})
	},
}
