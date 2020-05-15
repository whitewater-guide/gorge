package nzniwa

import (
	"time"

	"github.com/mattn/go-nulltype"
)

var tz, _ = time.LoadLocation("Pacific/Auckland")

type nzTime struct {
	time.Time
}

func (r *nzTime) UnmarshalJSON(b []byte) (err error) {
	// 2020-05-16T01:02:17.0000000
	t, err := time.ParseInLocation(`"2006-01-02T15:04:05"`, string(b), tz)
	r.Time = t.UTC()
	return
}

type niwaList struct {
	Indicators []indicators `json:"indicators"`
}

type indicators struct {
	Indicators []indicator `json:"indicators"`
}

type indicator struct {
	Location           string               `json:"Location"`
	LocationID         int                  `json:"LocationId"`
	LocationIdentifier string               `json:"LocationIdentifier"`
	LastUpdated        nzTime               `json:"LastUpdated"`
	LocX               float64              `json:"LocX"`
	LocY               float64              `json:"LocY"`
	Value              nulltype.NullFloat64 `json:"ValueNumber"`
	// ParameterID        int                  `json:"ParameterId"`
	// Active             bool                 `json:"Active"`
	// DatasetIdentifier  string  `json:"DatasetIdentifier"`
	// Unit               string  `json:"Unit"`
}

type niwaLocation struct {
	Data []struct {
		LocationID    int    `json:"LocationId"`
		ParameterID   int    `json:"ParameterId"`
		Unit          string `json:"Unit"`
		LocIdentifier string `json:"LocIdentifier"`
	} `json:"Data"`
}
