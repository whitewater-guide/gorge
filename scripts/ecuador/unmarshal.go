package ecuador

type ecuadorField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Process  string `json:"process"`
	Settable bool   `json:"settable"`
}

type ecuadorHead struct {
	Fields []ecuadorField `json:"fields"`
}

type ecuadorData struct {
	No   int           `json:"no"`
	Time string        `json:"time"`
	Vals []interface{} `json:"vals"`
}

type ecuadorRoot struct {
	Head ecuadorHead   `json:"head"`
	Data []ecuadorData `json:"data"`
	More bool          `json:"more"`
}

// This is second format, from http://186.42.174.236/InamhiEmas/#
type inamhiEmasItem struct {
	ID       int64   `json:"esta__id"`
	Name     string  `json:"puobnomb"`
	Code     string  `json:"puobcodi"`
	Lat      float64 `json:"coorlati,string"`
	Lng      float64 `json:"coorlong,string"`
	Alt      float64 `json:"cooraltu,string"`
	Status   string  `json:"estenomb"` //  'OPERATIVA', 'NO OPERATIVA'
	Category string  `json:"catenomb"` // 'HIDROLOGICA', 'METEOROLOGICA'
	Source   string  `json:"proesnomb"`
}
