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

type readings struct {
	Cm  []timestampedValue `json:"cm"`
	M3s []timestampedValue `json:"m3s"`
}

// https://api.riverzone.eu/?http#station-objects
type station struct {
	ID            string            `json:"id"`
	Revision      int               `json:"revision"`
	LastUpdatedTs int               `json:"lastUpdatedTs"`
	Enabled       bool              `json:"enabled"`
	RiverName     string            `json:"riverName"`
	StationName   string            `json:"stationName"`
	CountryCode   string            `json:"countryCode"`
	State         string            `json:"state"`
	Latlng        latLng            `json:"latlng"`
	Source        string            `json:"source"`
	SourceLink    string            `json:"sourceLink"`
	RefreshMins   int               `json:"refreshMins"`
	Notes         map[string]string `json:"notes"`
	Readings      readings          `json:"readings"`
	ParserConfigs string            `json:"parserConfigs"`
}

// https://api.riverzone.eu/?http#source-object-properties
type source struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LicensingTerms string `json:"licensingTerms"`
	Website        string `json:"website"`
}

// https://api.riverzone.eu/?http#stations-api
type stations struct {
	ElapsedMs int       `json:"elapsedMs"`
	Stations  []station `json:"stations"`
	Sources   []source  `json:"sources"`
}
