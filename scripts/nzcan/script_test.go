package nzcan

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
		"/RiverflowGeo/ALL":    "all.js",
		"/RiverflowList/NORTH": "north.html",
		"/RiverflowList/SOUTH": "south.html",
	}, nil)
}

func TestNzcan_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzcan{
		year: 2020,
		name: "nzcan",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzcan",
				Code:   "62105",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -42.45731,
				Longitude: 172.90635,
			},
			Name: "Waiau Toa~Clarence River at Jollies (NIWA)",
			URL:  "https://ecan.govt.nz/data/riverflow/sitedetails/62105",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzcan",
				Code:   "62107",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -42.11062,
				Longitude: 173.84193,
			},
			Name: "Waiau Toa~Clarence River at Clarence Valley Road Bridge",
			URL:  "https://ecan.govt.nz/data/riverflow/sitedetails/62107",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzcan",
				Code:   "68801",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -42.36892,
				Longitude: 173.67984,
			},
			Name: "Ashburton SH1",
			URL:  "https://ecan.govt.nz/data/riverflow/sitedetails/68801",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestNzcan_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzcan{
		year: 2020,
		name: "nzcan",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzcan",
				Code:   "62105",
			},
			Level: nulltype.NullFloat64Of(0.214),
			Flow:  nulltype.NullFloat64Of(7.556),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 16, 5, 25, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzcan",
				Code:   "62107",
			},
			Level: nulltype.NullFloat64Of(0.904),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 16, 5, 10, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzcan",
				Code:   "68801",
			},
			Level: nulltype.NullFloat64Of(1.589),
			Flow:  nulltype.NullFloat64Of(9.685),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 12, 23, 45, 0, 0, time.UTC),
			},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
