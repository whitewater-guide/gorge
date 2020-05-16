package wales

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
		"": "data.json",
	}, &testutils.HeaderAuthorizer{Key: "Ocp-Apim-Subscription-Key"})
}

func TestWales_Auth(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptWales{
		name:    "wales",
		url:     ts.URL + "/data.json",
		options: optionsWales{Key: "__bad__"},
	}
	_, err := s.ListGauges()
	assert.Error(t, err)
}

func TestWales_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptWales{
		name:    "wales",
		url:     ts.URL,
		options: optionsWales{Key: testutils.TestAuthKey},
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "wales",
				Code:   "4064",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  51.70875,
				Longitude: -3.34592,
			},
			Name: "Taff at Troedyrhiw",
			URL:  "https://rloi.naturalresources.wales/ViewDetails?station=4064",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "wales",
				Code:   "4067",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  51.4973,
				Longitude: -3.20988,
			},
			Name: "Taff at Western Avenue",
			URL:  "https://rloi.naturalresources.wales/ViewDetails?station=4067",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestWales_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptWales{
		name:    "wales",
		url:     ts.URL,
		options: optionsWales{Key: testutils.TestAuthKey},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "wales",
				Code:   "4064",
			},
			Timestamp: core.HTime{
				Time: time.Date(2016, time.June, 13, 14, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.275),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "wales",
				Code:   "4067",
			},
			Timestamp: core.HTime{
				Time: time.Date(2016, time.June, 13, 14, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.633),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
