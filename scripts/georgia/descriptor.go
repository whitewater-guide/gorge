package georgia

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "georgia",
	Mode: core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsGeorgia{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsGeorgia); ok {
			return &scriptGeorgia{
				name: name,
				url:  "http://meteo.gov.ge/index.php?l=2&pg=hd",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsGeorgia{})
	},
}
