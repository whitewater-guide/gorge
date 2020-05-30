package ukea

type stationsList struct {
	Stations []station `json:"items"`
}
type measure struct {
	ID string `json:"@id"`
	// Known parameters "flow", "rainfall", "wind", "temperature", "level"
	Parameter     string `json:"parameter"`
	ParameterName string `json:"parameterName"`
	Period        int    `json:"period"`
	// Known flow Qualifiers and their units:
	// - "Logged": "m3/s", "---"
	// - "": "Ml/d", "m3/s", "---"
	// - "Stage": "Ml/d", "l/s", "m3/s"
	// - "Speed": "Ml/d"
	// - "1": "m3/s"
	// - "2": "m3/s"
	// - "Water": "m3/s"
	// Known level Qualifiers and their units:
	// - "Stage": "mASD", "m", "mAOD", "---", "m3/s", "mm",
	// - "Height": "mASD", "m", "mAOD",
	// - "3": "m", "mm", "mAOD",
	// - "Reservoir Level": "m", "mBDAT",
	// - "Crest Tapping": "mASD",
	// - "Tidal Level": "mAOD", "m", "---", "mASD", "mm",
	// - "Groundwater": "---", "mAOD", "m", "mBDAT",
	// - "Radar": "mASD",
	// - "Sump Level": "m", "mASD",
	// - "4": "mAOD", "m",
	// - "Downstream Stage": "mASD", "---", "mAOD", "m",
	// - "2": "m", "mAOD", "mm", "mASD",
	// - "1": "---", "mASD", "mm", "m", "mAOD",
	// - "Logged": "mASD", "m",
	// - "": "m", "---", "mASD", "mAOD",
	// - "Water": "m",
	Qualifier string `json:"qualifier"`
	// Known flow units: "Ml/d", "l/s", "m3/s", "---"
	UnitName string `json:"unitName"`
}
type station struct {
	// ID               string    `json:"@id"`
	RLOIid string `json:"RLOIid,omitempty"`
	// CatchmentName    string    `json:"catchmentName,omitempty"`
	// DateOpened       string    `json:"dateOpened,omitempty"`
	// Easting          int       `json:"easting"`
	Label    string    `json:"label"`
	Lat      float64   `json:"lat"`
	Long     float64   `json:"long"`
	Measures []measure `json:"measures"`
	// Northing         int       `json:"northing"`
	// Notation  string `json:"notation"`
	RiverName string `json:"riverName,omitempty"`
	// StageScale       string `json:"stageScale,omitempty"`
	StationReference string `json:"stationReference"`
	// Status           string `json:"status,omitempty"`
	// Town             string `json:"town,omitempty"`
	// WiskiID          string `json:"wiskiID,omitempty"`
	// DatumOffset      int    `json:"datumOffset,omitempty"`
	// GridReference    string `json:"gridReference,omitempty"`
	// DownstageScale   string `json:"downstageScale,omitempty"`
}
