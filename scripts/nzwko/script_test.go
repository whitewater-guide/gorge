package nzwko

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
	return testutils.SetupFileServer(nil, nil)
}

func TestWaikato_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptWaikato{
		name:       "nzwko",
		listURL:    ts.URL + "/list.html",
		pageURL:    ts.URL + "/%s.html",
		numWorkers: 2,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzwko",
				Code:   "1065",
			},
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  -38.60393,
				Longitude: 174.76472,
			},
			Name: "Rauroa Farm Bridge on Awakino",
			URL:  ts.URL + "/1065.html",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzwko",
				Code:   "68",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -37.29369,
				Longitude: 175.063,
			},
			Name: "Whangamarino Control Structure on the Waikato River",
			URL:  ts.URL + "/68.html",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzwko",
				Code:   "158",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -37.3445,
				Longitude: 175.1853,
			},
			Name: "Upstream of Falls Road on the Whangamarino Wetland",
			URL:  ts.URL + "/158.html",
		},
	}

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestWaikato_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptWaikato{
		name:       "nzwko",
		listURL:    ts.URL + "/list.html",
		pageURL:    ts.URL + "/%s.html",
		numWorkers: 2,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"894": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzwko",
				Code:   "1065",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 12, 16, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(-0.762),
			Flow:  nulltype.NullFloat64Of(4.2),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzwko",
				Code:   "68",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 12, 16, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.43),
			Flow:  nulltype.NullFloat64{},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzwko",
				Code:   "158",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 13, 4, 0, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(2.962),
			Flow:  nulltype.NullFloat64{},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
