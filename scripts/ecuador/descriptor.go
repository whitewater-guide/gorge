package ecuador

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "ecuador",
	Mode: core.OneByOne,
	DefaultOptions: func() interface{} {
		return &optionsEcuador{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsEcuador); ok {
			return &scriptEcuador{
				name:           name,
				listURL1:       "http://186.42.174.243:9090/RTMCProject.js.jgz",
				listURL2:       "http://186.42.174.236/InamhiEmas/json.php?callback=",
				gaugeURLFormat: "http://186.42.174.243:9090/?command=DataQuery&uri=%s%%3Ahora1&format=json&mode=most-recent&p1=1&p2=&headsig=0&order=real-time&_=%d",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsEcuador{})
	},
}
