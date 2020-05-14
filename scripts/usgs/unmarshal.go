package usgs

import (
	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

const (
	paramFlow  = "00060" // Discharge, cubic feet per second
	paramLevel = "00065" // Gage height, feet
)

type ivRoot struct {
	Value ivRootValue `json:"value"`
}

type ivRootValue struct {
	TimeSeries []timeSeries `json:"timeSeries"`
}

type timeSeries struct {
	SourceInfo sourceInfo `json:"sourceInfo"`
	Variable   variable   `json:"variable"`
	Values     []values   `json:"values"`
	Name       string     `json:"name"`
}

type siteCode struct {
	Value      string `json:"value"`
	Network    string `json:"network"`
	AgencyCode string `json:"agencyCode"`
}
type sourceInfo struct {
	SiteCode []siteCode `json:"siteCode"`
}
type variableCode struct {
	Value string `json:"value"`
}
type variable struct {
	VariableCode []variableCode `json:"variableCode"`
	NoDataValue  nulltype.NullFloat64        `json:"noDataValue"`
}
type value struct {
	Value    string     `json:"value"`
	DateTime core.HTime `json:"dateTime"`
}
type values struct {
	Value []value `json:"value"`
}
