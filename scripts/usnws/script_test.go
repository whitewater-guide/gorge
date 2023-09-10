package usnws

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
		"": "ahps_national_obs.kmz",
	}, nil)
}

func TestUsnws_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUsnws{
		name:   "usnws",
		kmzUrl: ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "aplw3",
			},
			LevelUnit: "ft",
			Location: &core.Location{
				Latitude:  44.248056,
				Longitude: -88.423056,
			},
			Name:     "Fox River (North) at Appleton",
			URL:      "https://water.weather.gov/ahps2/hydrograph.php?wfo=GRB&gage=aplw3",
			Timezone: "America/Chicago",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "aubw1",
			},
			FlowUnit: "kcfs",
			Location: &core.Location{
				Latitude:  47.312500,
				Longitude: -122.202778,
			},
			Name:     "Green River (WA) near Auburn",
			URL:      "https://water.weather.gov/ahps2/hydrograph.php?wfo=SEW&gage=aubw1",
			Timezone: "America/Los_Angeles",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestUsnws_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUsnws{
		name:   "usnws",
		kmzUrl: ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "aplw3",
			},
			Level:     nulltype.NullFloat64Of(5.53),
			Timestamp: core.HTime{Time: time.Date(2023, time.September, 3, 14, 0, 0, 0, time.UTC)},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "aubw1",
			},
			Flow:      nulltype.NullFloat64Of(0.302),
			Timestamp: core.HTime{Time: time.Date(2023, time.September, 3, 13, 45, 0, 0, time.UTC)},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
