package smhi

import (
	"encoding/json"
	"time"

	"github.com/mattn/go-nulltype"
)

type response struct {
	// Updated   int64 `json:"updated"`
	// Parameter struct {
	// 	Key  string `json:"key"`
	// 	Name string `json:"name"`
	// 	Unit string `json:"unit"`
	// } `json:"parameter"`
	// Period struct {
	// 	Key     string `json:"key"`
	// 	From    int64  `json:"from"`
	// 	To      int64  `json:"to"`
	// 	Summary string `json:"summary"`
	// } `json:"period"`
	// Link []struct {
	// 	Href string `json:"href"`
	// 	Rel  string `json:"rel"`
	// 	Type string `json:"type"`
	// } `json:"link"`
	Station []station `json:"station"`
}

type station struct {
	ID int `json:"id"`
	// Key               string  `json:"key"`
	Name string `json:"name"`
	// Owner             string  `json:"owner"`
	// MeasuringStations string  `json:"measuringStations"`
	// Region            int     `json:"region"`
	// From              int64   `json:"from"`
	// To                int64   `json:"to"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Value     []value `json:"value"`
}

type value struct {
	Date  time.Time            `json:"date"`
	Value nulltype.NullFloat64 `json:"value"`
	// Quality string  `json:"quality"`
}

func (b *value) UnmarshalJSON(data []byte) error {
	type alias value
	aux := &struct {
		Date int64 `json:"date"`
		*alias
	}{
		alias: (*alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	b.Date = time.UnixMilli(aux.Date).UTC()
	return nil
}
