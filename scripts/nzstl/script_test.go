package nzstl

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

func TestNzstl_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzstl{
		name: "nzstl",
		url:  ts.URL + "/list.json",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzstl",
				Code:   "f2714dc2-2594-3381-914a-9169695cabcd",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -45.73744,
				Longitude: 168.11601,
			},
			Name: "Aparima River at Dunrobin",
			URL:  "http://envdata.es.govt.nz/",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzstl",
				Code:   "0c54dbaa-6615-3b83-8bcb-401ed6e884ee",
			},
			LevelUnit: "m",
			Location: &core.Location{
				Latitude:  -45.8533,
				Longitude: 168.12916,
			},
			Name: "Aparima River at Etalvale",
			URL:  "http://envdata.es.govt.nz/",
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestNzstl_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzstl{
		name: "nzstl",
		url:  ts.URL + "/list.json",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzstl",
				Code:   "f2714dc2-2594-3381-914a-9169695cabcd",
			},
			Level: nulltype.NullFloat64Of(0.511),
			Flow:  nulltype.NullFloat64Of(2.45),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 18, 17, 0, 0, 0, time.UTC),
			},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzstl",
				Code:   "0c54dbaa-6615-3b83-8bcb-401ed6e884ee",
			},
			Level: nulltype.NullFloat64Of(0.691),
			Timestamp: core.HTime{
				Time: time.Date(2020, time.May, 18, 16, 30, 0, 0, time.UTC),
			},
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
