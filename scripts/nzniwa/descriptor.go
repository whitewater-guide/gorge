package nzniwa

import (
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

var Descriptor = &core.ScriptDescriptor{
	Name:        "nzniwa",
	Description: "New Zealand: National Institute of Water and Atmospheric Research",
	Mode:        core.AllAtOnce,
	DefaultOptions: func() interface{} {
		return &optionsNzniwa{}
	},
	Factory: func(name string, options interface{}) (core.Script, error) {
		if _, ok := options.(*optionsNzniwa); ok {
			return &scriptNzniwa{
				name:        name,
				numWorkers:  5,
				flowURL:     "https://hydrowebportal.niwa.co.nz/Map/LocationIndicators?dataFilters=&parameters=327&interval=Latest&value=2&subValue=&subValueType=&refPeriod=1&calendar=1&type=Statistic&legend=6&legendFilter=-1&date=2020-05-16&endDate=&showValue=true&datasetBlacklist=true&datasetTimeSeries=true&datasetPrimary=true&locationState=&datasetState=&datasetStatistics=true&utcOffset=0&folder=",
				locationURL: "https://hydrowebportal.niwa.co.nz/Data/Data_Location?location=",
				// levelURL:    "https://hydrowebportal.niwa.co.nz/Map/LocationIndicators?dataFilters=&parameters=255&interval=Latest&value=1&subValue=&subValueType=&refPeriod=1&calendar=1&type=Statistic&legend=&legendFilter=-1&date=2020-05-16&endDate=&showValue=true&datasetBlacklist=true&datasetTimeSeries=true&datasetPrimary=true&locationState=&datasetState=&datasetStatistics=true&utcOffset=0&folder=",
			}, nil
		}
		return nil, fmt.Errorf("failed to cast %T", optionsNzniwa{})
	},
}
