package norway

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
		"/Stations":     "/stations.json",
		"/Observations": "/observations.json",
	}, &testutils.HeaderAuthorizer{Key: "X-API-Key"})
}

func TestNorway_Auth(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNorway{
		name:    "norway",
		urlBase: ts.URL,
		apiKey:  "__bad__",
	}
	_, err := s.ListGauges()
	assert.Error(t, err)
}

func TestNorway_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNorway{
		name:    "norway",
		urlBase: ts.URL,
		apiKey:  testutils.TestAuthKey,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "2.284",
			},
			Name: "Finna - Sælatunga",
			Location: &core.Location{
				Latitude:  61.88439,
				Longitude: 9.06212,
				Altitude:  427,
			},
			FlowUnit:  "m³/s",
			LevelUnit: "m",
			URL:       "https://sildre.nve.no/station/2.284.0",
			Timezone:  "Europe/Oslo",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "2.13.0",
			},
			Name: "Sjoa - Nedre Sjodalsvatn",
			Location: &core.Location{
				Latitude:  61.56064,
				Longitude: 8.91856,
				Altitude:  942,
			},
			FlowUnit:  "m³/s",
			LevelUnit: "m",
			URL:       "https://sildre.nve.no/station/2.13.0",
			Timezone:  "Europe/Oslo",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestNorwayHarvest(t *testing.T) {
	// https://hydapi.nve.no/api/v1/Observations?StationId=2.284.0%2C2.13.0&Parameter=1000%2C1001&ResolutionTime=0&ReferenceTime=PT1H%2F
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNorway{
		name:    "norway",
		urlBase: ts.URL,
		apiKey:  testutils.TestAuthKey,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"2.284": {}, "2.13.0": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "2.284",
			},
			Timestamp: core.HTime{
				Time: time.Date(2024, time.June, 8, 18, 0, 0, 0, time.UTC),
			},
			Flow:  nulltype.NullFloat64Of(19.97738),
			Level: nulltype.NullFloat64Of(406.4),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "norway",
				Code:   "2.13.0",
			},
			Timestamp: core.HTime{
				Time: time.Date(2024, time.June, 8, 18, 0, 0, 0, time.UTC),
			},
			Flow:  nulltype.NullFloat64Of(56.41846),
			Level: nulltype.NullFloat64Of(940.731),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
