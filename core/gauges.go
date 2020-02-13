package core

// Location is EPSG4326 coordinate
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude,omitempty"`
}

// GaugeID identifies gauge using script and code pair
type GaugeID struct {
	// id of script from gorge's script registry
	Script string `json:"script"`
	// unique gauge code from upstream
	// if no code is provided (ex. Georgia) script must generate it's own stable unique codes
	Code string `json:"code"`
}

// Less is helper for sorting gauges
func (id *GaugeID) Less(other *GaugeID) bool {
	if id.Script == other.Script {
		return id.Code < other.Code
	}
	return id.Script < other.Script
}

// Gauge represents gauge/station from upstream source
type Gauge struct {
	GaugeID
	Name string `json:"name"`
	// This is webpage URL from original source
	URL string `json:"url,omitempty"`
	// Water level unit, e.g, "m"/"ft"/"cm"
	LevelUnit string `json:"levelUnit,omitempty"`
	// Water flow/discharge unit, e.g. "cfs"/"m3/s"
	FlowUnit string `json:"flowUnit,omitempty"`
	// Station location, if known
	Location *Location `json:"location,omitempty"`
}

// Gauges is slice of Gauge with helper methods for sorting
type Gauges []Gauge

func (g Gauges) Len() int {
	return len(g)
}

func (g Gauges) Less(i, j int) bool {
	return g[i].GaugeID.Less(&(g[j].GaugeID))
}

func (g Gauges) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}
