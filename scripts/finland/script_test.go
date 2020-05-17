package finland

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
	return testutils.SetupFileServer(map[string]string{
		"/Paikka": "paikka_{{ .skip }}.json",
		"":        "virtaama.json",
	}, nil)
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
				Latitude:  62.28418,
				Longitude: 27.89071,
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
				Latitude:  61.2717,
				Longitude: 27.7022,
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
