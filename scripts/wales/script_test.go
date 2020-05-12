package wales

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
		key := r.Header.Get("Ocp-Apim-Subscription-Key")
		if key != "__thekey__" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized.\n")) //nolint:errcheck
			return
		}
		file, _ := os.Open("./test_data/data.json")
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestWales_Auth(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptWales{
		name:    "wales",
		url:     ts.URL,
		options: optionsWales{Key: "__bad__"},
	}
	_, err := s.ListGauges()
	assert.Error(t, err)
}

func TestWales_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptWales{
		name:    "wales",
		url:     ts.URL,
		options: optionsWales{Key: "__thekey__"},
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "wales",
				Code:   "4064",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  51.70875,
				Longitude: -3.34592,
			},
			Name: "Taff at Troedyrhiw",
			URL:  "https://rloi.naturalresources.wales/ViewDetails?station=4064",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "wales",
				Code:   "4067",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  51.4973,
				Longitude: -3.20988,
			},
			Name: "Taff at Western Avenue",
			URL:  "https://rloi.naturalresources.wales/ViewDetails?station=4067",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestWales_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptWales{
		name:    "wales",
		url:     ts.URL,
		options: optionsWales{Key: "__thekey__"},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "wales",
				Code:   "4064",
			},
			Timestamp: core.HTime{
				Time: time.Date(2016, time.June, 13, 14, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.275),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "wales",
				Code:   "4067",
			},
			Timestamp: core.HTime{
				Time: time.Date(2016, time.June, 13, 14, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.633),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
