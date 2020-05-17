package nztrc

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
		"/9": "flow.json",
		"/7": "level.json",
	}, nil)
}

func TestNztrc_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNztrc{
		name: "nztrc",
		url:  ts.URL + "/",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nztrc",
				Code:   "8",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -39.2743,
				Longitude: 173.75802,
			},
			Name: "Kapoaiaia at Cape Egmont",
			URL:  "https://www.trc.govt.nz/environment/maps-and-data/site-details/?siteID=8&measureID=9",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nztrc",
				Code:   "12",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Latitude:  -39.3884,
				Longitude: 174.46663,
			},
			Name: "Mangaehu at Huinga",
			URL:  "https://www.trc.govt.nz/environment/maps-and-data/site-details/?siteID=12&measureID=9",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nztrc",
				Code:   "6",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -39.72498,
				Longitude: 174.43136,
			},
			Name: "Kaikura below 7346",
			URL:  "https://www.trc.govt.nz/environment/maps-and-data/site-details/?siteID=6&measureID=7",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestNztrc_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNztrc{
		name: "nztrc",
		url:  ts.URL + "/",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	now := time.Now().In(tz)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nztrc",
				Code:   "8",
			},
			Flow:  nulltype.NullFloat64Of(0.41),
			Level: nulltype.NullFloat64Of(0.52),
			Timestamp: core.HTime{
				Time: time.Date(now.Year(), now.Month(), now.Day(), 5, 30, 0, 0, tz).UTC(),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nztrc",
				Code:   "12",
			},
			Flow: nulltype.NullFloat64Of(2.27),
			Timestamp: core.HTime{
				Time: time.Date(now.Year(), now.Month(), now.Day(), 5, 45, 0, 0, tz).UTC(),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nztrc",
				Code:   "6",
			},
			Level: nulltype.NullFloat64Of(0.28),
			Timestamp: core.HTime{
				Time: time.Date(now.Year(), now.Month(), now.Day(), 5, 30, 0, 0, tz).UTC(),
			},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
