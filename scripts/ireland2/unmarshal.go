package ireland2

import (
	"github.com/mattn/go-nulltype"
)

type riverspyResp struct {
	Rivers []river `json:"rivers"`
}

type river struct {
	Code      string               `json:"code"`
	Rivername string               `json:"rivername"`
	Sitename  string               `json:"sitename"`
	Latitude  float64              `json:"latitude"`
	Longitude float64              `json:"longitude"`
	Updated   int64                `json:"updated"`
	Lastlevel nulltype.NullFloat64 `json:"lastlevel"`
	Yunit     string               `json:"yunit"`
	// Trend       string               `json:"trend"`
	// Updatefreq  string               `json:"updatefreq"`
	// Owner       string               `json:"owner"`
	// Phoneno     string               `json:"phoneno"`
	// Cmperbeep   int                  `json:"cmperbeep"`
	Graphlink string `json:"graphlink"`
	// Infolink    string               `json:"infolink"`
	// Gaugetype   string               `json:"gaugetype"`
	// Ontime      string               `json:"ontime"`
	// Quiettime   string               `json:"quiettime"`
	// Alertlevel  string               `json:"alertlevel"`
	// Voltage     string               `json:"voltage"`
	// Scrapelevel int                  `json:"scrapelevel"`
	// Greenlevel  int                  `json:"greenlevel"`
	// Yellowlevel int                  `json:"yellowlevel"`
	// Orangelevel int                  `json:"orangelevel"`
	// Pinklevel   int                  `json:"pinklevel"`
	// Redlevel    int                  `json:"redlevel"`
	// Temp        int                  `json:"temp"`
}
