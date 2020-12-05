package wales

import (
	"time"
)

type wTime struct {
	time.Time
}

func (wt *wTime) UnmarshalJSON(b []byte) (err error) {
	t, err := time.Parse(`"2006-01-02T15:04:05Z07:00"`, string(b))
	wt.Time = t.UTC()
	return
}

type stationParam struct {
	Parameter         int     `json:"parameter"`
	ParamNameEN       string  `json:"paramNameEN"`
	ParamNameCY       string  `json:"paramNameCY"`
	ParameterStatusEN string  `json:"parameterStatusEN"`
	ParameterStatusCY string  `json:"parameterStatusCY"`
	Units             string  `json:"units"`
	LatestValue       float64 `json:"latestValue"`
	LatestTime        wTime   `json:"LatestTime"`
}

type stationData struct {
	Coordinates struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"coordinates"`
	Location   int            `json:"location"`
	NameEN     string         `json:"nameEN"`
	NameCY     string         `json:"nameCY"`
	StatusEN   string         `json:"statusEN"`
	StatusCY   string         `json:"statusCY"`
	URL        string         `json:"url"`
	Ngr        string         `json:"ngr"`
	TitleEn    string         `json:"titleEn"`
	TitleCy    string         `json:"titleCy"`
	Parameters []stationParam `json:"parameters"`
}
