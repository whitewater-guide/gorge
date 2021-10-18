package canada

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
	return testutils.SetupFileServer(nil, nil)
}

func TestCanada_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptCanada{
		name:      "canada",
		baseURL:   ts.URL,
		provinces: getProvinces(""),
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Code:   "01AD003",
				Script: "canada",
			},
			Name:      "[NB] ST. FRANCIS RIVER AT OUTLET OF GLASIER LAKE",
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  47.20661,
				Longitude: -68.95694,
			},
			URL:      "https://wateroffice.ec.gc.ca/report/real_time_e.html?stn=01AD003",
			Timezone: "America/Moncton",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Code:   "02YA002",
				Script: "canada",
			},
			Name:      "[NL] BARTLETTS RIVER NEAR ST. ANTHONY",
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  51.44922,
				Longitude: -55.64125,
			},
			URL:      "https://wateroffice.ec.gc.ca/report/real_time_e.html?stn=02YA002",
			Timezone: "America/St_Johns",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Code:   "02YC001",
				Script: "canada",
			},
			Name:      "[NL] TORRENT RIVER AT BRISTOL'S POOL",
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  50.60747,
				Longitude: -57.15161,
			},
			URL:      "https://wateroffice.ec.gc.ca/report/real_time_e.html?stn=02YC001",
			Timezone: "America/St_Johns",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Code:   "08GA043",
				Script: "canada",
			},
			Name:      "[BC] CHEAKAMUS RIVER NEAR BRACKENDALE",
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  49.81603,
				Longitude: -123.15008,
			},
			URL:      "https://wateroffice.ec.gc.ca/report/real_time_e.html?stn=08GA043",
			Timezone: "America/Vancouver",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestCanada_HarvestRemap(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptCanada{
		name:      "canada",
		baseURL:   ts.URL,
		provinces: map[string]bool{"BC": true},
	}

	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "canada",
				Code:   "08GA043",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 19, 12, 45, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.864),
			Flow:  nulltype.NullFloat64Of(18.8),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "canada",
				Code:   "08GA043",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 19, 12, 50, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.863),
			Flow:  nulltype.NullFloat64Of(18.7),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "canada",
				Code:   "10CD005",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 19, 12, 45, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.675),
		},
	}

	actual, err := core.HarvestSlice(&s, core.StringSet{"08GA043": {}}, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestCanada_HarvestProvinces(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptCanada{
		name:      "canada",
		baseURL:   ts.URL,
		numWokers: 2,
		provinces: getProvinces(""),
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	if assert.NoError(t, err) {
		assert.Len(t, actual, 11)
	}
}
