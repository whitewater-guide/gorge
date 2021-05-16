package sepa

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "sepa",
	Description: "Scottish Environment Protection Agency",
	Mode:        core.OneByOne,
	DefaultOptions: func() interface{} {
		return &optionsSepa{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsSepa); ok {
			return &scriptSepa{
				name: name,
				// Currently list URL is not working
				listURL:      "https://www2.sepa.org.uk/waterlevels/CSVs",
				gaugeURLBase: "https://www2.sepa.org.uk/HydroData/api/Level15",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsSepa{})
	},
}
