package quebec

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/testutils"
)

// 030247 m
// 023402 m m3/s
// 050409 m3/s

func setupTestServer() *httptest.Server {
	return testutils.SetupFileServer(nil, nil)
}

func TestQuebec_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptQuebec{
		name:              "quebec",
		codesURL:          ts.URL + "/codes.html",
		referenceListURL:  ts.URL + "/references.csv",
		stationURLFormat:  ts.URL + "/stations/%s.html",
		readingsURLFormat: ts.URL + "/readings/%s.html",
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "023402",
			},
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Name:      "Chaudière (023402)",
			Location: &core.Location{
				Latitude:  46.58777,
				Longitude: -71.21638,
			},
			URL: "https://www.cehq.gouv.qc.ca/suivihydro/graphique.asp?NoStation=023402",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "030247",
			},
			LevelUnit: "m",
			Name:      "Barrage Bombardier (030247)",
			Location: &core.Location{
				Latitude:  45.4625,
				Longitude: -72.13333,
			},
			URL: "https://www.cehq.gouv.qc.ca/suivihydro/graphique.asp?NoStation=030247",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "050409",
			},
			FlowUnit: "m3/s",
			Name:     "Bras du Nord (050409)",
			Location: &core.Location{
				Latitude:  46.97166,
				Longitude: -71.85444,
			},
			URL: "https://www.cehq.gouv.qc.ca/suivihydro/graphique.asp?NoStation=050409",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestQuebec_Harvest_HTML(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptQuebec{
		name:              "quebec",
		codesURL:          ts.URL + "/codes.html",
		referenceListURL:  ts.URL + "/references.csv",
		stationURLFormat:  ts.URL + "/stations/%s.html",
		readingsURLFormat: ts.URL + "/readings/%s.csv",
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{"023402": {}}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "023402",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 24, 11, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(112.86),
			Flow:  nulltype.NullFloat64Of(62.32),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "023402",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.January, 24, 11, 15, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64Of(112.86),
			Flow:  nulltype.NullFloat64Of(62.22),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}

	// File encoding windows-1252
	actual, err = core.HarvestSlice(&s, core.StringSet{"062701": {}}, 0)
	expected = core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "062701",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.March, 16, 13, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64{},
			Flow:  nulltype.NullFloat64Of(9.716),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "062701",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.March, 16, 13, 15, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64{},
			Flow:  nulltype.NullFloat64Of(9.792),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
	// File encoding utf8
	actual, err = core.HarvestSlice(&s, core.StringSet{"062702": {}}, 0)
	expected = core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "062702",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.March, 16, 13, 30, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64{},
			Flow:  nulltype.NullFloat64Of(9.716),
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "quebec",
				Code:   "062702",
			},
			Timestamp: core.HTime{
				Time: time.Date(2020, time.March, 16, 13, 15, 0, 0, time.UTC),
			},
			Level: nulltype.NullFloat64{},
			Flow:  nulltype.NullFloat64Of(9.792),
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
