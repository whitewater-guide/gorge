package russia1

import (
	"time"

	"github.com/mattn/go-nulltype"
)

type cTime struct {
	time.Time
}

var timezone, _ = time.LoadLocation("Europe/Moscow")

func (ct *cTime) UnmarshalJSON(b []byte) (err error) {
	if string(b) == "null" {
		return
	}
	t, err := time.ParseInLocation(`"2006-01-02 15:04:05"`, string(b), timezone)
	ct.Time = t.UTC()
	return
}

type russia1Feature struct {
	Geometry struct {
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Desc string `json:"desc"`
		Fix  bool   `json:"fix"`
		Type int    `json:"type"`
		Data struct {
			RiverLevel struct {
				Level struct {
					Bs nulltype.NullFloat64 `json:"bs"`
				} `json:"level"`
				Time cTime `json:"time"`
			} `json:"river_level"`
		} `json:"data"`
	} `json:"properties"`
}

type russia1Features struct {
	Features []russia1Feature `json:"features"`
}
