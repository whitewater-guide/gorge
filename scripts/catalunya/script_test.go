package catalunya

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

func TestCatalunya_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptCatalunya{
		name:            "catalunya",
		gaugesURL:       ts.URL + "/list.json",
		measurementsURL: ts.URL + "/observations.json",
	}
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "catalunya",
				Code:   "CALC000007",
			},
			Name:     "Riu Ridaura - Santa Cristina d'Aro (m³/s)",
			URL:      "http://aca-web.gencat.cat/sentilo-catalog-web/component/AFORAMENT-EST.171812-001/detail",
			FlowUnit: "m³/s",
			Location: &core.Location{
				Latitude:  41.81813,
				Longitude: 2.97912,
				Altitude:  0.0,
			},
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "catalunya",
				Code:   "CALC001110",
			},
			Name:      "Riu Muga - Castelló d'Empúries (cm)",
			URL:       "http://aca-web.gencat.cat/sentilo-catalog-web/component/AFORAMENT-EST.170470-003/detail",
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  42.25438,
				Longitude: 3.07199,
				Altitude:  0.0,
			},
		},
	}
	actual, err := s.ListGauges()
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestCatalunya_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptCatalunya{
		name:            "catalunya",
		gaugesURL:       ts.URL + "/list.json",
		measurementsURL: ts.URL + "/observations.json",
	}
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "catalunya",
				Code:   "CALC001110",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 21, 18, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(148.512),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "catalunya",
				Code:   "CALC000007",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 21, 18, 10, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(4.296),
		},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
