package usgs

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
		filename := ""
		if strings.HasPrefix(r.URL.Path, "/site") {
			stateCd := r.URL.Query().Get("stateCd")
			parameterCd := r.URL.Query().Get("parameterCd")
			filename = fmt.Sprintf("./test_data/sites/%s/%s.html", stateCd, parameterCd)
		}
		file, _ := os.Open(filename)
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, file)
		if err != nil {
			panic("failed to send test file")
		}
	}))
}

func TestUSGS_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUSGS{
		name:    "usgs",
		url:     ts.URL,
		stateCd: "wa",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usgs",
				Code:   "12010000",
			},
			FlowUnit:  "ft3/s",
			LevelUnit: "ft",
			Name:      "NASELLE RIVER NEAR NASELLE, WA",
			URL:       "https://waterdata.usgs.gov/nwis/inventory?agency_code=USGS&site_no=12010000",
			Location: &core.Location{
				Latitude:  46.37399,
				Longitude: -123.74348,
				Altitude:  24,
			},
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usgs",
				Code:   "12025100",
			},
			LevelUnit: "ft",
			Name:      "CHEHALIS RIVER AT WWTP AT CHEHALIS, WA",
			URL:       "https://waterdata.usgs.gov/nwis/inventory?agency_code=USGS&site_no=12025100",
			Location: &core.Location{
				Latitude:  46.66093,
				Longitude: -122.98401,
			},
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "usgs",
				Code:   "12017000",
			},
			FlowUnit: "ft3/s",
			Name:     "NORTH RIVER NEAR RAYMOND, WA",
			URL:      "https://waterdata.usgs.gov/nwis/inventory?agency_code=USGS&site_no=12017000",
			Location: &core.Location{
				Latitude:  46.80731,
				Longitude: -123.85072,
				Altitude:  7.39,
			},
		},
	}

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestUSGS_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUSGS{
		name:    "usgs",
		url:     ts.URL,
		stateCd: "wa",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"202": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "usgs",
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
