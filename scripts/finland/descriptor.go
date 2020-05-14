package finland

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "finland",
	Description: "The Finnish Environment Institute (SYKE)",
	Mode:        core.OneByOne,
	DefaultOptions: func() interface{} {
		return &optionsFinland{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsFinland); ok {
			return &scriptFinland{
				name: name,
				url:  "http://rajapinnat.ymparisto.fi/api/Hydrologiarajapinta/1.1/odata",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsFinland{})
	},
}
