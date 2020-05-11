package finland

import (
	"fmt"
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
		isList := strings.HasSuffix(r.URL.Path, "Paikka")
		var filename string
		if isList {
			skip := r.URL.Query().Get("$skip")
			filename = fmt.Sprintf("./test_data/paikka_%s.json", skip)
		} else {
			filename = "./test_data/virtaama.json"
		}
		file, _ := os.Open(filename)
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestFinland_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptFinland{
		name: "finland",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "finland",
				Code:   "1003",
			},
			LevelUnit: "cm",
			Location: &core.Location{
				Latitude:  63.0041,
				Longitude: 26.4053,
			},
			Name: "Tervo - Nilakka, Äyskoski - 1402710 (level)",
			URL:  "https://wwwi2.ymparisto.fi/i2/14/q1402710y/wqfi.html",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "finland",
				Code:   "514",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Latitude:  67.0814,
				Longitude: 25.4450,
			},
			Name: "Sodankylä - Unari, Sodankylä - 65501 (discharge)",
			URL:  "",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestFinland_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptFinland{
		name: "finland",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"894": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "finland",
				Code:   "894",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 7, 21, 0, 0, 0, time.UTC),
			},
			Flow: nulltype.NullFloat64Of(37.15),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
