package usgs

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
		"/site": "/sites/{{ .stateCd }}/{{ .parameterCd }}.html",
		"":      "/iv/data.json",
	}, nil)
}

func TestUSGS_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUSGS{
		name:    "usgs",
		url:     ts.URL,
		stateCd: "wa",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usgs",
				Code:   "12010000",
			},
			FlowUnit:  "ft3/s",
			LevelUnit: "ft",
			Name:      "NASELLE RIVER NEAR NASELLE, WA",
			URL:       "https://waterdata.usgs.gov/nwis/inventory?agency_code=USGS&site_no=12010000",
			Location: &core.Location{
				Latitude:  46.37399,
				Longitude: -123.74348,
				Altitude:  24,
			},
			Timezone: "America/Los_Angeles",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usgs",
				Code:   "12025100",
			},
			LevelUnit: "ft",
			Name:      "CHEHALIS RIVER AT WWTP AT CHEHALIS, WA",
			URL:       "https://waterdata.usgs.gov/nwis/inventory?agency_code=USGS&site_no=12025100",
			Location: &core.Location{
				Latitude:  46.66093,
				Longitude: -122.98401,
			},
			Timezone: "America/Los_Angeles",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usgs",
				Code:   "12017000",
			},
			FlowUnit: "ft3/s",
			Name:     "NORTH RIVER NEAR RAYMOND, WA",
			URL:      "https://waterdata.usgs.gov/nwis/inventory?agency_code=USGS&site_no=12017000",
			Location: &core.Location{
				Latitude:  46.80731,
				Longitude: -123.85072,
				Altitude:  7.39,
			},
			Timezone: "America/Los_Angeles",
		},
	}

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestUSGS_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUSGS{
		name:    "usgs",
		url:     ts.URL,
		stateCd: "wa",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"12010000": {}, "12025100": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "usgs",
				Code:   "12010000",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 14, 14, 30, 0, 0, time.UTC),
			},
			Flow:  nulltype.NullFloat64Of(316),
			Level: nulltype.NullFloat64Of(5.26),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
