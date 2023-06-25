package ireland2

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
		"/": "indexdata.cgi",
	}, nil)
}

func TestIreland_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptIreland2{
		name: "ireland2",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ireland2",
				Code:   "00001",
			},
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  51.9847,
				Longitude: -9.30276,
			},
			Name:     "FLESK - Clydagh Bridge",
			URL:      "https://riverspy.net/spygauge.html?code=00001",
			Timezone: "Europe/Dublin",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ireland2",
				Code:   "00008",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Latitude:  51.9,
				Longitude: -8.66167,
			},
			Name:     "Lee - Inniscarra Dam",
			URL:      "https://riverspy.net/esbgauge.html?code=00008",
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
	s := scriptIreland2{
		name: "ireland2",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ireland2",
				Code:   "00001",
			},
			Timestamp: core.HTime{
				// Mon Apr 11 2022 21:00:00 GMT+0000
				Time: time.Date(2022, time.April, 11, 21, 00, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(-19),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ireland2",
				Code:   "00008",
			},
			Timestamp: core.HTime{
				// Sun Jun 25 2023 08:00:00 GMT+0000
				Time: time.Date(2023, time.June, 25, 8, 0, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(2),
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
