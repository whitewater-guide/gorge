package nzmbh

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
	t, err := time.ParseInLocation(`"02 January 2006 15:04"`, string(b), tz)
	r.Time = t.UTC()
	return
}

type riverReport []struct {
	SiteName   string `json:"SiteName"`
	LastUpdate nzTime `json:"LastUpdate"`
	Flow       string `json:"Flow"`
	Stage      string `json:"Stage"`
}

type siteList struct {
	XMLName       xml.Name `xml:"FeatureCollection"`
	FeatureMember []struct {
		SiteList struct {
			Site     string `xml:"Site"`
			Location struct {
				Point struct {
					Pos string `xml:"pos"`
				} `xml:"Point"`
			} `xml:"Location"`
		} `xml:"SiteList"`
	} `xml:"featureMember"`
}
