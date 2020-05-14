package nzbop

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name: "nzbop",
	Mode: core.OneByOne,
	DefaultOptions: func() interface{} {
		return &optionsBop{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsBop); ok {
			return &scriptBop{
				name:       name,
				numWorkers: 5,
				listURL:    "http://monitoring.boprc.govt.nz/MonitoredSites/cgi-bin/hydwebserver.cgi/districts/details?district=3",
				pageURL:    "http://monitoring.boprc.govt.nz/MonitoredSites/cgi-bin/hydwebserver.cgi/sites/details?site=%s&treecatchment=22",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsBop{})
	},
}
