package smhi

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
		"/api/version/latest/parameter/2/station-set/all/period/latest-hour/data.json": "data.json",
	}, nil)
}

func TestSmhi_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptSmhi{
		name: "smhi",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "smhi",
				Code:   "1583",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Latitude:  63.4334,
				Longitude: 18.3002,
			},
			Name:     "VÄSTERSEL (1583)",
			URL:      "https://www.smhi.se/en/weather/observations/observations#ws=wpt-a,proxy=wpt-a,tab=vatten,param=waterflow,stationid=1583,type=water",
			Timezone: "Europe/Stockholm",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "smhi",
				Code:   "1456",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Latitude:  67.7664,
				Longitude: 19.975,
			},
			Name:     "KAALASJÄRVI (1456)",
			URL:      "https://www.smhi.se/en/weather/observations/observations#ws=wpt-a,proxy=wpt-a,tab=vatten,param=waterflow,stationid=1456,type=water",
			Timezone: "Europe/Stockholm",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestSmhi_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptSmhi{
		name: "smhi",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "smhi",
				Code:   "1583",
			},
			Timestamp: core.HTime{
				Time: time.Date(2023, time.October, 8, 15, 15, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(13.5757),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "smhi",
				Code:   "1583",
			},
			Timestamp: core.HTime{
				Time: time.Date(2023, time.October, 8, 15, 30, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(13.6137),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "smhi",
				Code:   "1456",
			},
			Timestamp: core.HTime{
				Time: time.Date(2023, time.October, 8, 15, 15, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(44.9892),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "smhi",
				Code:   "1456",
			},
			Timestamp: core.HTime{
				Time: time.Date(2023, time.October, 8, 15, 30, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(44.9892),
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
