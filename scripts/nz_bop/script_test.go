package nz_bop

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
		filename := "./test_data" + r.URL.Path
		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		file, _ := os.Open(filename)
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestBop_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptBop{
		name:       "nz_bop",
		listURL:    ts.URL + "/list.html",
		pageURL:    ts.URL + "/%s.html",
		numWorkers: 2,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nz_bop",
				Code:   "9199",
			},
			LevelUnit: "m",
			Name:      "Mangaone at Braemar Rd",
			URL:       ts.URL + "/9199.html",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nz_bop",
				Code:   "220",
			},
			Name:      "Wairoa at above Ruahihi Power Station",
			URL:       ts.URL + "/220.html",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -37.77573,
				Longitude: 176.05292,
				Altitude:  15,
			},
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nz_bop",
				Code:   "202",
			},
			Name:      "Kopurererua at SH 29",
			URL:       ts.URL + "/202.html",
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  -37.73268,
				Longitude: 176.1101,
				Altitude:  10,
			},
		},
	}

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, actual, expected)
	}
}

func TestBop_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptBop{
		name:       "nz_bop",
		listURL:    ts.URL + "/list.html",
		pageURL:    ts.URL + "/%s.html",
		numWorkers: 2,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"202": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nz_bop",
				Code:   "202",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 13, 18, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.684),
			Flow:  nulltype.NullFloat64Of(1.654),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestBop_ParseList(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptBop{
		name:       "nz_bop",
		listURL:    ts.URL + "/list.html",
		pageURL:    ts.URL + "/%s.html",
		numWorkers: 2,
	}
	actual, err := s.parseList()
	expected := []string{
		"11386",
		"333",
		"193",
		"220",
		"276",
		"203",
		"202",
		"11050",
		"219",
		"297",
		"187",
		"11512",
		"186",
		"244",
		"184",
		"257",
		"302",
		"249",
		"9210",
		"9003",
		"4632",
		"179",
		"9214",
		"9199",
		"176",
		"307",
		"163",
		"11162",
		"165",
		"168",
		"166",
		"11277",
		"267",
		"344",
		"157",
		"158",
		"154",
		"159",
		"153",
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
