package ireland

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
		"/": "latest.json",
	}, nil)
}

func TestIreland_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptIreland{
		name: "ireland",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ireland",
				Code:   "01041",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  54.838318,
				Longitude: -7.575758,
			},
			Name:     "Sandy Mills",
			URL:      "https://waterlevel.ie/0000001041/0001/",
			Timezone: "Europe/Dublin",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ireland",
				Code:   "01043",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  54.799769,
				Longitude: -7.790749,
			},
			Name:     "Ballybofey",
			URL:      "https://waterlevel.ie/0000001043/0001/",
			Timezone: "Europe/Dublin",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestIreland_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptIreland{
		name: "ireland",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ireland",
				Code:   "01041",
			},
			Timestamp: core.HTime{
				Time: time.Date(2023, time.January, 21, 22, 45, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.048),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ireland",
				Code:   "01043",
			},
			Timestamp: core.HTime{
				Time: time.Date(2023, time.January, 21, 19, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.934),
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
