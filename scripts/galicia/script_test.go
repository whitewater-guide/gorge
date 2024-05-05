package galicia

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

func TestGalicia_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGalicia{
		name: "galicia",
		url:  ts.URL + "/galicia.json",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "141125",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  43.3302,
				Longitude: -8.42652,
			},
			Name:     "[CO] Pastoriza @ A Coruña",
			URL:      "https://servizos.meteogalicia.gal/mgafos/estacionshistorico/historico.action?idEst=141125",
			Timezone: "Europe/Madrid",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30546",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  42.7779,
				Longitude: -8.102,
			},
			Name:     "[PO] Arnego Ulla @ Agolada",
			URL:      "https://servizos.meteogalicia.gal/mgafos/estacionshistorico/historico.action?idEst=140515",
			Timezone: "Europe/Madrid",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestGalicia_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGalicia{
		name: "galicia",
		url:  ts.URL + "/galicia.json",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "141125",
			},
			Timestamp: core.HTime{
				Time: time.Date(2024, time.May, 3, 0, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.248),
			Flow:  nulltype.NullFloat64{},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30546",
			},
			Timestamp: core.HTime{
				Time: time.Date(2024, time.May, 3, 0, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.465),
			Flow:  nulltype.NullFloat64Of(9.372),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
