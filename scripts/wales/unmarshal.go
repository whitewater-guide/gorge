package wales

import (
	"time"
)

type wTime struct {
	time.Time
}

func (wt *wTime) UnmarshalJSON(b []byte) (err error) {
	t, err := time.Parse(`"02/01/2006 15:04"`, string(b))
	wt.Time = t.UTC()
	return
}

type walesData struct {
	Features []struct {
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
		Properties struct {
			Location    string `json:"Location"`
			// LatestValue string `json:"LatestValue"`
			LatestValue string `json:"LatestValue"`
			LatestTime  wTime `json:"LatestTime"`
			NameEN      string `json:"NameEN"`
			ParamNameEN string `json:"ParamNameEN"`
			TitleEN     string `json:"TitleEN"`
			Units       string `json:"Units"`
			URL         string `json:"url"`
		} `json:"properties"`
	} `json:"features"`
}