package nzstl

import (
	"encoding/json"
	"strings"
	"time"
)

var tz, _ = time.LoadLocation("Pacific/Auckland")

type nzTime struct {
	time.Time
}

func (r *nzTime) UnmarshalJSON(b []byte) (err error) {
	str := strings.Replace(string(b), ".", "", -1)
	t, err := time.ParseInLocation(`"02/01/2006 3:04:05 pm"`, str, tz)
	r.Time = t.UTC()
	return
}

type val struct {
	Value       string
	Measurement string
}

func (v *val) UnmarshalJSON(b []byte) error {
	if string(b) == `""` {
		return nil
	}
	var vv struct {
		Value       string `json:"Value"`
		Measurement string `json:"Measurement"`
	}
	err := json.Unmarshal(b, &vv)
	if err != nil {
		return err
	}
	v.Value = vv.Value
	v.Measurement = vv.Measurement
	return nil
}

type list struct {
	Sites []struct {
		Site       string `json:"Site"`
		DataTo     nzTime `json:"DataTo"`
		WaterLevel val    `json:"WaterLevel"`
		Flow       val    `json:"Flow"`
		Northing   string `json:"Northing"`
		Easting    string `json:"Easting"`
	} `json:"sites"`
}
