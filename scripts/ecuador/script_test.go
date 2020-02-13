package ecuador

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
		path := "./test_data/H0064.json"
		switch r.URL.Path {
		case "/list1":
			path = "./test_data/RTMCProject.js.jgz"
		case "/list2":
			path = "./test_data/list2.jsonp"
		}
		file, _ := os.Open(path)
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestEcuador_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptEcuador{
		name:           "ecuador",
		listURL1:       ts.URL + "/list1",
		listURL2:       ts.URL + "/list2",
		gaugeURLFormat: ts.URL + "/%s/%d",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ecuador",
				Code:   "H1156",
			},
			LevelUnit: "m",
			Name:      "Napo En Ahuano",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ecuador",
				Code:   "H0719",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -0.30277,
				Longitude: -77.775,
				Altitude:  1490,
			},
			Name: "Quijos Dj Oyacachi",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ecuador",
				Code:   "H5011",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -0.44,
				Longitude: -77.008,
				Altitude:  265,
			},
			Name: "Payamino Aj Napo",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestEcuador_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptEcuador{
		name:           "ecuador",
		listURL1:       ts.URL + "/list1",
		listURL2:       ts.URL + "/list2",
		gaugeURLFormat: ts.URL + "/%s/%d",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"H0064": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ecuador",
				Code:   "H0064",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 23, 11, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.13),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
