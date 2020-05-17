package nztrc

type site struct {
	SiteID      int    `json:"siteID"`
	Title       string `json:"title"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
	Measure     string `json:"measure"`
	Unit        string `json:"unit"`
	Link        string `json:"link"`
	Description string `json:"description"`
	// MeasureType string      `json:"measureType"`
	// FillColour  string      `json:"fillColour"`
	// Icon        interface{} `json:"icon"`
	// CanSwim     string      `json:"canSwim"`
}
