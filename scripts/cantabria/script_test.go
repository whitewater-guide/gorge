package cantabria

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"

	"github.com/stretchr/testify/assert"
)

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filename := "./test_data/list.html"
		if r.URL.Path != "/list" {
			filename = "./test_data/gauge.html"
		}
		file, _ := os.Open(filename)
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestCantabria_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptCantabria{
		name:         "cantabria",
		listURL:      ts.URL + "/list",
		gaugeURLBase: ts.URL + "/gauge",
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
		listURL:      ts.URL + "/list",
		gaugeURLBase: ts.URL + "/gauge",
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
