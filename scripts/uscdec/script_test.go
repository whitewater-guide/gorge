package uscdec

import (
	"context"
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
		"/getAll":  "all_{{ .sens_num }}.html",
		"/staMeta": "meta_{{ .station_id }}.html",
	}, nil)
}

func TestUscdec_Details(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUSCDEC{
		name: "uscdec",
		url:  ts.URL,
	}
	actual, err := s.parseDetails(context.Background(), "JED")
	expected := &core.Gauge{
		GaugeID: core.GaugeID{
			Script: "uscdec",
			Code:   "JED",
		},
		FlowUnit: "cfs",
		Location: &core.Location{
			Altitude:  36.576,
			Latitude:  41.7915,
			Longitude: -124.07618,
		},
		Name:     "SMITH R NR CRESCENT CITY (JED SMITH SP)",
		URL:      ts.URL + "/staMeta?station_id=JED",
		Timezone: "US/Pacific",
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}

	actual, err = s.parseDetails(context.Background(), "FPT")
	expected = &core.Gauge{
		GaugeID: core.GaugeID{
			Script: "uscdec",
			Code:   "FPT",
		},
		FlowUnit: "cfs",
		Location: &core.Location{
			Altitude:  0.0,
			Latitude:  38.45611,
			Longitude: -121.5003,
		},
		Name:     "SACRAMENTO RIVER AT FREEPORT",
		URL:      ts.URL + "/staMeta?station_id=FPT",
		Timezone: "US/Pacific",
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestUscdec_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptUSCDEC{
		name: "uscdec",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected1 := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "uscdec",
			Code:   "FCC",
		},
		Timestamp: core.HTime{ // 09/30/2023 04:45
			Time: time.Date(2023, time.September, 30, 11, 45, 0, 0, time.UTC),
		},
		Level: nulltype.NullFloat64{},
		Flow:  nulltype.NullFloat64Of(4),
	}
	expected2 := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "uscdec",
			Code:   "SMS",
		},
		Timestamp: core.HTime{
			Time: time.Date(2023, time.September, 30, 11, 30, 0, 0, time.UTC),
		},
		Level: nulltype.NullFloat64{},
		Flow:  nulltype.NullFloat64Of(241),
	}
	if assert.NoError(t, err) {
		assert.Contains(t, actual, expected1)
		assert.Contains(t, actual, expected2)
	}
}
