package nzniwa

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/testutils"
)

func setupTestServer() *httptest.Server {
	return testutils.SetupFileServer(map[string]string{
		"/locations": "{{ .id }}.json",
		"":           "flow.json",
	}, nil)
}

func TestNzniwa_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzniwa{
		name:        "nzniwa",
		locationURL: ts.URL + "/locations?id=",
		numWorkers:  2,
		flowURL:     ts.URL + "/flow.json",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzniwa",
				Code:   "40810",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Latitude:  174.733,
				Longitude: -38.62335,
			},
			Name: "Awakino at Gorge",
			URL:  "https://hydrowebportal.niwa.co.nz/Data/Location/Summary/Location/40810/Interval/Latest",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzniwa",
				Code:   "93202",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Latitude:  172.38639,
				Longitude: -41.76348,
			},
			Name: "Buller at Longford",
			URL:  "https://hydrowebportal.niwa.co.nz/Data/Location/Summary/Location/93202/Interval/Latest",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzniwa",
				Code:   "62103",
			},
			FlowUnit: "l/s",
			Location: &core.Location{
				Latitude:  172.96479,
				Longitude: -42.37541,
			},
			Name: "Acheron at Clarence",
			URL:  "https://hydrowebportal.niwa.co.nz/Data/Location/Summary/Location/62103/Interval/Latest",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestNzniwa_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzniwa{
		name:        "nzniwa",
		locationURL: ts.URL + "/locations/",
		numWorkers:  2,
		flowURL:     ts.URL + "/flow.json",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzniwa",
				Code:   "40810",
			},
			Flow: nulltype.NullFloat64Of(4.5),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 15, 13, 02, 17, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzniwa",
				Code:   "93202",
			},
			Flow: nulltype.NullFloat64Of(40.4),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 15, 13, 15, 8, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzniwa",
				Code:   "62103",
			},
			Flow: nulltype.NullFloat64Of(8.4),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 15, 13, 05, 52, 0, time.UTC),
			},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
