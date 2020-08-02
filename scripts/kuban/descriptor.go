package kuban

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "kuban",
	Description: "Russia: Kuban drainage",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsKuban{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsKuban); ok {
			return &scriptKuban{
				name: name,
				url:  "http://193.7.160.230/web/osio/hydro/TABL.files/sheet001.html",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsKuban{})
	},
}
