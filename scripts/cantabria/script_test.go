package cantabria

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/testutils"

	"github.com/stretchr/testify/assert"
)

func setupTestServer() *httptest.Server {
	return testutils.SetupFileServer(map[string]string{
		"/list": "list.html",
		"":      "gauge.html",
	}, nil)
}

func TestCantabria_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptCantabria{
		name:         "cantabria",
		listURL:      ts.URL + "/list.html",
		gaugeURLBase: ts.URL + "/",
	}
	actual, err := s.ListGauges()
	expected := core.Gauge{
		GaugeID: core.GaugeID{
			Script: "cantabria",
			// Code:   "A613",
			Code: "A047",
		},
		Name:      "Eo - Ribera de Piquín",
		URL:       "https://www.chcantabrico.es/sistema-automatico-de-informacion-detalle-estacion?cod_estacion=A047",
		LevelUnit: "m",
		Location: &core.Location{
			Latitude:  43.17956,
			Longitude: -7.19921,
		},
		Timezone: "Europe/Madrid",
	}
	if assert.NoError(t, err) && assert.Len(t, actual, 76) {
		assert.Equal(t, expected, actual[0])
	}
}

func TestCantabria_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptCantabria{
		name:         "cantabria",
		listURL:      ts.URL + "/list.html",
		gaugeURLBase: ts.URL + "/",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "cantabria",
			Code:   "A047",
		},
		Timestamp: core.HTime{
			Time: time.Date(2020, time.January, 20, 18, 30, 0, 0, time.UTC),
		},
		Level: nulltype.NullFloat64Of(0.51),
	}
	if assert.NoError(t, err) && assert.Len(t, actual, 76) {
		assert.Equal(t, expected, actual[0])
	}
}
