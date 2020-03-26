package russia1

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
		file, _ := os.Open("./test_data" + r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestRussia1_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptRussia1{
		name:      "russia1",
		gaugesURL: ts.URL + "/data.json",
	}
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "russia1",
				Code:   "ЭМЕРСИТ-0249",
			},
			Name:      "ЭМЕРСИТ-0249",
			URL:       "http://www.emercit.com/map/",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  44.473948,
				Longitude: 33.803694,
				Altitude:  0.0,
			},
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "russia1",
				Code:   "АГК-0034",
			},
			Name:      "АГК-0034",
			URL:       "http://www.emercit.com/map/",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  44.32223,
				Longitude: 38.70207,
				Altitude:  0.0,
			},
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "russia1",
				Code:   "ЭМЕРСИТ-0007Д",
			},
			Name:      "ЭМЕРСИТ-0007Д",
			URL:       "http://www.emercit.com/map/",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  45.0123,
				Longitude: 38.997167,
				Altitude:  0.0,
			},
		},
	}
	actual, err := s.ListGauges()
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestRussia1_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptRussia1{
		name:      "russia1",
		gaugesURL: ts.URL + "/data.json",
	}
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "russia1",
				Code:   "ЭМЕРСИТ-0249",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.March, 26, 13, 40, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(256.02199115753),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "russia1",
				Code:   "АГК-0034",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.March, 26, 13, 42, 32, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(-0.19361280441284),
		},
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
