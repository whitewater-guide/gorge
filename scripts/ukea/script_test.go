package ukea

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

func TestUkea_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUkea{
		name: "ukea",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ukea",
				Code:   "4173TH",
			},
			LevelUnit: "mASD",
			Location: &core.Location{
				Latitude:  51.39833,
				Longitude: -0.18362,
			},
			Name: "River Wandle at Ravensbury Mill",
			URL:  "https://environment.data.gov.uk/flood-monitoring/id/stations/4173TH.html",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ukea",
				Code:   "1029TH",
			},
			LevelUnit: "mASD",
			Location: &core.Location{
				Latitude:  51.87476,
				Longitude: -1.74008,
			},
			Name: "7041 - River Dikler - Bourton Dickler",
			URL:  "https://environment.data.gov.uk/flood-monitoring/id/stations/1029TH.html",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ukea",
				Code:   "F1906",
			},
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  54.08070,
				Longitude: -2.02477,
			},
			Name: "8276 - River Wharfe - Netherside Hall",
			URL:  "https://environment.data.gov.uk/flood-monitoring/id/stations/F1906.html",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "ukea",
				Code:   "1301TH",
			},
			LevelUnit: "mASD",
			Location: &core.Location{
				Latitude:  51.78930,
				Longitude: -1.30693,
			},
			Name: "7055 - River Thames - Kings Lock",
			URL:  "https://environment.data.gov.uk/flood-monitoring/id/stations/1301TH.html",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
	s.rloi = rloiWith
	actual, err = s.ListGauges()
	if assert.NoError(t, err) {
		assert.Equal(t, expected[1:], actual)
	}
	s.rloi = rloiWithout
	actual, err = s.ListGauges()
	if assert.NoError(t, err) {
		assert.Equal(t, expected[0:1], actual)
	}
}

func TestUkea_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUkea{
		name: "ukea",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ukea",
				Code:   "2432TH",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 2, 18, 15, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.921),
		},
		// level-groundwater-i-1_h-mBDAT is harvested
		// it'll be filtered out on next stages
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ukea",
				Code:   "5003",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 3, 7, 00, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(41.03),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ukea",
				Code:   "E1660",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 30, 4, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(0.166),
			Flow:  nulltype.NullFloat64Of(0.261),
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
