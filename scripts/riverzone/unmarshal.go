package riverzone

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/mattn/go-nulltype"
)

type timestamp struct {
	time.Time
}

func (ts *timestamp) UnmarshalJSON(b []byte) error {
	epoch, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	t := time.Unix(epoch, 0)
	ts.Time = t.UTC()
	return nil
}

type timestampedValue struct {
	Timestamp timestamp            `json:"ts"`
	Value     nulltype.NullFloat64 `json:"v"`
}

type latLng struct {
	Lat float64
	Lng float64
}

func (ll *latLng) UnmarshalJSON(b []byte) error {
	var tuple []int
	err := json.Unmarshal(b, &tuple)
	if err != nil {
		return err
	}
	ll.Lat = float64(tuple[0]) / 1000000
	ll.Lng = float64(tuple[1]) / 1000000
	return nil
}

type translations struct {
	value string
}

type sourceLinks struct {
	value string
}

type sourceLink struct {
	Level string `json:"level"`
	Flow  string `json:"flow"`
}

func (s *translations) UnmarshalJSON(b []byte) error {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	// prefer english
	for k, v := range raw {
		s.value = string(v[1 : len(v)-1])
		if k == "en" {
			break
		}
	}
	return nil
}

func (s *sourceLinks) UnmarshalJSON(b []byte) error {
	var raw map[string]sourceLink
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	// prefer flow to level and english to other languages
	for k, v := range raw {
		s.value = v.Level
		if v.Flow != "" {
			s.value = v.Flow
		}
		if k == "en" {
			break
		}
	}
	return nil
}

type readings struct {
	Cm  []timestampedValue `json:"cm"`
	M3s []timestampedValue `json:"m3s"`
}

// https://api.rivermap.ch/doc/v2/#list-stations
type station struct {
	ID                  string       `json:"id"`
	Type                string       `json:"type"`
	Revision            int          `json:"revision"`
	RevisionTs          int          `json:"revisionTs"`
	Sensors             []string     `json:"sensors"`
	IsActive            bool         `json:"isActive"`
	Name                string       `json:"name"`
	River               translations `json:"river,omitempty"`
	CountryCode         string       `json:"countryCode"`
	State               string       `json:"state"`
	Latlng              latLng       `json:"latlng"`
	DataSourceID        string       `json:"dataSourceId"`
	SourceLinks         sourceLinks  `json:"sourceLinks,omitempty"`
	ParserConfigs       string       `json:"parserConfigs"`
	ObservationFreqMins int          `json:"observationFreqMins"`
	PublishingFreqMins  int          `json:"publishingFreqMins"`
	Note                translations `json:"note"`
}

// https://api.rivermap.ch/doc/v2/#list-stations
type source struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LicensingTerms string `json:"licensingTerms"`
	Website        string `json:"website"`
	CountryCode    string `json:"countryCode"`
}

// https://api.rivermap.ch/doc/v2/#list-stations
type stationsResp struct {
	ElapsedMs int       `json:"elapsedMs"`
	Stations  []station `json:"stations"`
	Sources   []source  `json:"sources"`
	License   string    `json:"license"`
}

// https://api.rivermap.ch/doc/v2/#get-all-readings
type readingsResp struct {
	ElapsedMs int                 `json:"elapsedMs"`
	Readings  map[string]readings `json:"readings"`
	License   string              `json:"license"`
}
