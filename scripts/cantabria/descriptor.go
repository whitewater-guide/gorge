package cantabria

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "cantabria",
	Mode: core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsCantabria{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsCantabria); ok {
			return &scriptCantabria{
				name:         name,
				listURL:      "https://www.chcantabrico.es/web/chcmovil/tabla-resumen-niveles-",
				gaugeURLBase: "https://www.chcantabrico.es/sistema-automatico-de-informacion-detalle-estacion?cod_estacion=",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsCantabria{})
	},
}
