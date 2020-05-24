package nzwgn

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
		"/": "{{ .Request }}.xml",
	}, nil)
}

func TestNzwgn_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzwgn{
		name: "nzwgn",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzwgn",
				Code:   "7318daa7-db55-37ce-80ed-e20ab489f2b5",
			},
			LevelUnit: "mm",
			Location: &core.Location{
				Latitude:  -41.01501,
				Longitude: 175.53607,
			},
			Name: "Booths Creek at Andersons Line",
			URL:  "https://graphs.gw.govt.nz/",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzwgn",
				Code:   "9d63b387-b201-3838-9b2b-807a3d0c4f08",
			},
			LevelUnit: "mm",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  -40.89357,
				Longitude: 175.69179,
			},
			Name: "Kopuaranga River at Stuarts",
			URL:  "https://graphs.gw.govt.nz/",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestNzwgn_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzwgn{
		name: "nzwgn",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzwgn",
				Code:   "7318daa7-db55-37ce-80ed-e20ab489f2b5",
			},
			Level: nulltype.NullFloat64Of(332),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 24, 5, 0, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzwgn",
				Code:   "9d63b387-b201-3838-9b2b-807a3d0c4f08",
			},
			Level: nulltype.NullFloat64Of(788),
			Flow:  nulltype.NullFloat64Of(0.676535),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 24, 5, 0, 0, 0, time.UTC),
			},
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
