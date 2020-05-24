package nzwgn

import (
	"encoding/xml"
	"time"
)

var tz, _ = time.LoadLocation("Pacific/Auckland")

// var tz, _ = time.LoadLocation("NZST")

type nzTime struct {
	time.Time
}

func (r *nzTime) UnmarshalJSON(b []byte) (err error) {
	t, err := time.ParseInLocation(`"02 January 2006 15:04:05"`, string(b), tz)
	r.Time = t.UTC()
	return
}

type siteList struct {
	XMLName xml.Name `xml:"HilltopServer"`
	// Text    string   `xml:",chardata"`
	// Agency  string   `xml:"Agency"`
	Site []struct {
		// Text      string `xml:",chardata"`
		Name      string `xml:"Name,attr"`
		Latitude  string `xml:"Latitude"`
		Longitude string `xml:"Longitude"`
	} `xml:"Site"`
}

type measurements struct {
	XMLName     xml.Name `xml:"Hilltop"`
	Error       string   `xml:"Error"`
	Measurement []struct {
		SiteName   string `xml:"SiteName,attr"`
		DataSource struct {
			ItemInfo struct {
				ItemName string `xml:"ItemName"`
				Units    string `xml:"Units"`
			} `xml:"ItemInfo"`
		} `xml:"DataSource"`
		Data struct {
			E struct {
				T  nzTime `xml:"T"`
				I1 string `xml:"I1"`
			} `xml:"E"`
		} `xml:"Data"`
	} `xml:"Measurement"`
}
