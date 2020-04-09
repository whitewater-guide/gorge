package ukraine

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _ := os.Open("./test_data/kml_hydro_warn.kml")
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestUkraine_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUkraine{
		name: "ukraine",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauge{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "42136",
		},
		LevelUnit: "cm",
		Name:      "Прут Татарів",
		URL:       ts.URL,
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 192)
		assert.Contains(t, actual, expected)
	}
}

func TestUkraine_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUkraine{
		name: "ukraine",
		url:  ts.URL,
	}
	now := time.Now().UTC().Truncate(time.Hour)
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "42136",
		},
		Timestamp: core.HTime{
			Time: now,
		},
		Level: nulltype.NullFloat64Of(136),
	}
	if assert.NoError(t, err) {
		assert.Contains(t, actual, expected)
	}
}
