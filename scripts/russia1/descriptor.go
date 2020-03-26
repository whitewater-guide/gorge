package russia1

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "russia1",
	Mode: core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsRussia1{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsRussia1); ok {
			return &scriptRussia1{
				name:      name,
				gaugesURL: "http://www.emercit.com/map/overall.php",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsRussia1{})
	},
}
