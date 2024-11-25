package galicia2

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "galicia2",
	Description: "Spain: Confederación Hidrográfica del Miño-Sil",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsGalicia2{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsGalicia2); ok {
			return &scriptGalicia2{
				name:           name,
				listURL:        "https://saih.chminosil.es/index.php?url=/datos/resumen_excel",
				gaugeURLFormat: "https://saih.chminosil.es/index.php?url=/datos/ficha/estacion:%s",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsGalicia2{})
	},
}
