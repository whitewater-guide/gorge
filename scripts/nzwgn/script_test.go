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
				Code:   "fa1e028d-743e-3210-99e8-e0757dce8959",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -41.83477,
				Longitude: 173.73081,
			},
			Name: "Awatere River at Awapiri",
			URL:  "http://hydro.marlborough.govt.nz/reports/riverreport.html",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzwgn",
				Code:   "a4976a69-9c0b-3965-9766-a1494bfbe9ec",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -41.46991,
				Longitude: 173.97576,
			},
			Name: "Grovetown Lagoon at Drain Y",
			URL:  "http://hydro.marlborough.govt.nz/reports/riverreport.html",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
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
				Code:   "fa1e028d-743e-3210-99e8-e0757dce8959",
			},
			Flow:  nulltype.NullFloat64Of(3.38),
			Level: nulltype.NullFloat64Of(1.408),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 16, 16, 0, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzwgn",
				Code:   "a4976a69-9c0b-3965-9766-a1494bfbe9ec",
			},
			Level: nulltype.NullFloat64Of(0.048),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 16, 16, 25, 0, 0, time.UTC),
			},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
