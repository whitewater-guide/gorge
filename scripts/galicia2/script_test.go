package galicia2

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := "." + r.URL.Path
		if !strings.HasSuffix(path, "table.xls") {
			path = "./test_data/A015.html"
		}
		file, _ := os.Open(path)
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestGalicia2_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGalicia2{
		name:           "galicia2",
		listURL:        ts.URL + "/test_data/table.xls",
		gaugeURLFormat: ts.URL + "/test_data/%s.html",
	}
	actual, err := s.ListGauges()
	expected := core.Gauge{
		GaugeID: core.GaugeID{
			Script: "galicia2",
			Code:   "A015",
		},
		LevelUnit: "m",
		Location: &core.Location{
			Latitude:  42.87182,
			Longitude: -7.52761,
			Altitude:  348.65,
		},
		Name: "Río Neira en Páramo (o)",
		URL:  ts.URL + "/test_data/A015.html",
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 58)
		assert.Contains(t, actual, expected)
	}
}

func TestGalicia2_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGalicia2{
		name:           "galicia2",
		listURL:        ts.URL + "/test_data/table.xls",
		gaugeURLFormat: ts.URL + "/test_data/%s.html",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "galicia2",
			Code:   "A015",
		},
		Timestamp: core.HTime{
			Time: time.Date(2020, time.January, 23, 8, 0, 0, 0, time.UTC),
		},
		Level: nulltype.NullFloat64Of(0.44),
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 58)
		assert.Contains(t, actual, expected)
	}
}
