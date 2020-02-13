package chile

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "chile",
	Mode: core.OneByOne,
	DefaultOptions: func() interface{} {
		return &optionsChile{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsChile); ok {
			return &scriptChile{
				name:            name,
				selectFormURL:   "http://dgasatel.mop.cl/filtro_paramxestac_new2.asp",
				webmapIDPageURL: "https://www.arcgis.com/sharing/rest/content/items/d508beb3a88f43d28c17a8ec9fac5ef0/data?f=json",
				webmapURLFormat: "https://www.arcgis.com/sharing/rest/content/items/%s/data?f=json",
				xlsURL:          "http://dgasatel.mop.cl/cons_det_instan_xls.asp",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsChile{})
	},
}
