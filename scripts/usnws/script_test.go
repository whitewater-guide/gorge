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
		"": "data_{{ if ne .returnCountOnly nil }}count{{ end }}{{ .resultOffset }}.json",
	}, nil)
}

func TestUsnws_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUsnws{
		name:       "usnws",
		url:        ts.URL,
		pageSize:   1,
		numWorkers: 2,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "AAIT2",
			},
			LevelUnit: "ft",
			FlowUnit:  "kcfs",
			Location: &core.Location{
				Latitude:  30.22111,
				Longitude: -97.79333,
			},
			Name:     "Williamson Creek / Manchaca Road at Austin / TX",
			URL:      "https://water.weather.gov/ahps2/hydrograph.php?wfo=ewx&gage=aait2",
			Timezone: "America/Chicago",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "LCTU1",
			},
			FlowUnit:  "kcfs",
			LevelUnit: "ft",
			Location: &core.Location{
				Latitude:  40.57777,
				Longitude: -111.79722,
			},
			Name:     "Little Cottonwood Creek / Salt Lake City / UT",
			URL:      "https://water.weather.gov/ahps2/hydrograph.php?wfo=slc&gage=lctu1",
			Timezone: "America/Denver",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "ANDI1",
			},
			FlowUnit:  "kcfs",
			LevelUnit: "ft",
			Location: &core.Location{
				Latitude:  43.34361,
				Longitude: -115.4775,
			},
			Name:     "South Fork Boise River / Anderson Ranch Dam / ID",
			URL:      "https://water.weather.gov/ahps2/hydrograph.php?wfo=boi&gage=andi1",
			Timezone: "America/Boise",
		},
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 3)
		assert.Contains(t, actual, expected[0])
		assert.Contains(t, actual, expected[1])
		assert.Contains(t, actual, expected[2])
	}
}

func TestUsnws_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUsnws{
		name:       "usnws",
		url:        ts.URL,
		pageSize:   1,
		numWorkers: 2,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "AAIT2",
			},
			Level:     nulltype.NullFloat64Of(1.99),
			Timestamp: core.HTime{Time: time.Date(2023, time.October, 1, 18, 30, 0, 0, time.UTC)},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "LCTU1",
			},
			Flow:      nulltype.NullFloat64Of(0.03),
			Level:     nulltype.NullFloat64Of(0.5),
			Timestamp: core.HTime{Time: time.Date(2023, time.October, 1, 12, 0, 0, 0, time.UTC)},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "usnws",
				Code:   "ANDI1",
			},
			Flow:      nulltype.NullFloat64Of(0.3),
			Level:     nulltype.NullFloat64Of(3.0),
			Timestamp: core.HTime{Time: time.Date(2023, time.October, 1, 20, 15, 0, 0, time.UTC)},
		},
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 3)
		assert.Contains(t, actual, expected[0])
		assert.Contains(t, actual, expected[1])
		assert.Contains(t, actual, expected[2])
	}
}
