package quebec2

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

func TestQuebec2_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptQuebec2{
		name: "quebec2",
		url:  ts.URL + "/data.json",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-15-filtre-j",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Longitude: -68.0264,
				Latitude:  48.1778,
			},
			Name:     "Mistigougèche - Bassin Lac Mistigougèche",
			URL:      "https://www.hydroquebec.com/generation/flows-water-level.html",
			Timezone: "America/Toronto",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-15-deverse-h",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Longitude: -68.0264,
				Latitude:  48.1778,
			},
			Name:     "Mistigougèche - Évacuateur Mistigougèche",
			URL:      "https://www.hydroquebec.com/generation/flows-water-level.html",
			Timezone: "America/Toronto",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-99-turbine-h",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Longitude: -69.2785,
				Latitude:  49.2358,
			},
			Name:     "Bersimis-2 - Centrale Bersimis-2",
			URL:      "https://www.hydroquebec.com/generation/flows-water-level.html",
			Timezone: "America/Toronto",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-99-total-h",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Longitude: -69.2785,
				Latitude:  49.2358,
			},
			Name:     "Bersimis-2 - Site Bersimis-2",
			URL:      "https://www.hydroquebec.com/generation/flows-water-level.html",
			Timezone: "America/Toronto",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestNztrc_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptQuebec2{
		name: "quebec2",
		url:  ts.URL + "/data.json",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)

	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-15-filtre-j",
			},
			Flow: nulltype.NullFloat64Of(6.48),
			Timestamp: core.HTime{
				Time: time.Date(2022, time.April, 22, 0, 0, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-15-filtre-j",
			},
			Flow: nulltype.NullFloat64Of(6.93),
			Timestamp: core.HTime{
				Time: time.Date(2022, time.April, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-15-deverse-h",
			},
			Flow: nulltype.NullFloat64Of(1.57),
			Timestamp: core.HTime{
				Time: time.Date(2022, time.April, 22, 11, 0, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-99-turbine-h",
			},
			Flow: nulltype.NullFloat64Of(297.78),
			Timestamp: core.HTime{
				Time: time.Date(2022, time.April, 22, 11, 0, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-99-turbine-h",
			},
			Flow: nulltype.NullFloat64Of(298.72),
			Timestamp: core.HTime{
				Time: time.Date(2022, time.April, 22, 12, 0, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec2",
				Code:   "3-99-total-h",
			},
			Flow: nulltype.NullFloat64Of(297.78),
			Timestamp: core.HTime{
				Time: time.Date(2022, time.April, 22, 11, 0, 0, 0, time.UTC),
			},
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
