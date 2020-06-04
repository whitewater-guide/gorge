package nzhkb

type hkbFeature struct {
	Attributes struct {
		ObjectID               int     `json:"ObjectID"`
		HilltopSite            string  `json:"Hilltop_Site"`
		HydrotelUnits          string  `json:"Hydrotel_Units"`
		URL                    string  `json:"URL"`
		HydrotelLastSampleTime int64   `json:"Hydrotel_LastSampleTime"`
		HydrotelCurrentValue   float64 `json:"Hydrotel_CurrentValue"`
	} `json:"attributes"`
	Geometry struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"geometry"`
}

type hkbList struct {
	Features []hkbFeature `json:"features"`
}
