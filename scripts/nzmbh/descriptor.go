package nzmbh

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "nzmbh",
	Description: "New Zealand: Marlborough District Council",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsNzmbh{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsNzmbh); ok {
			return &scriptNzmbh{
				name:        name,
				reportURL:   "http://hydro.marlborough.govt.nz/reports/riverreport.json",
				siteListURL: "http://hydro.marlborough.govt.nz/data.hts?service=WFS&request=GetFeature&typename=SiteList",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNzmbh{})
	},
}
