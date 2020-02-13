package galicia

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
		file, _ := os.Open("./test_data/galicia.json")
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestGalicia_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptGalicia{
		name: "galicia",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30431",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  43.47774,
				Longitude: -7.33423,
			},
			Name: "Masma",
			URL:  "http://www2.meteogalicia.gal/servizos/AugasdeGalicia/estacionsinfo.asp?Nest=30431",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30433",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  43.55815,
				Longitude: -7.37617,
			},
			Name: "Ouro",
			URL:  "http://www2.meteogalicia.gal/servizos/AugasdeGalicia/estacionsinfo.asp?Nest=30433",
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
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30431",
			},
			Timestamp: core.HTime{
				Time: time.Date(2018, time.February, 19, 19, 40, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.07),
			Flow:  nulltype.NullFloat64Of(8.689578),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "galicia",
				Code:   "30433",
			},
			Timestamp: core.HTime{
				Time: time.Date(2018, time.February, 19, 19, 40, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(1.63),
			Flow:  nulltype.NullFloat64Of(7.1654706),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
