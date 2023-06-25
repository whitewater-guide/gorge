package ireland2

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "ireland2",
	Description: "https://www.riverspy.net",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsIreland2{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsIreland2); ok {
			return &scriptIreland2{
				name: name,
				url:  "https://www.riverspy.net/indexdata.cgi",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsIreland2{})
	},
}
