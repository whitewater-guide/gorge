package nzcan

type markersList []struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	SiteName string  `json:"SiteName"`
	SiteNo   string  `json:"SiteNo"`
	// Value       string  `json:"Value"`
	// Colour      int     `json:"Colour"`
	// Total       string  `json:"Total"`
	// TotalColour int     `json:"TotalColour"`
	// Type        string  `json:"Type"`
}
