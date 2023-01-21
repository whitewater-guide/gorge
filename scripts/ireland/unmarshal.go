package ireland

import (
	"time"
)

type featureProps struct {
	StationRef  string    `json:"station_ref"`
	StationName string    `json:"station_name"`
	SensorRef   string    `json:"sensor_ref"`
	RegionID    int       `json:"region_id"`
	Datetime    time.Time `json:"datetime"`
	Value       string    `json:"value"`
	ErrCode     int       `json:"err_code"`
	URL         string    `json:"url"`
	CsvFile     string    `json:"csv_file"`
}

type geometry struct {
	Coordinates []float64 `json:"coordinates"`
}

type feature struct {
	Properties featureProps `json:"properties"`
	Geometry   geometry     `json:"geometry"`
}

type geojson struct {
	Features []feature `json:"features"`
}
