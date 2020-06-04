package nzhkb

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
		"/7": "{{if .returnGeometry}}/7/measurements.json{{else}}/7/gauges.json{{end}}",
		"/8": "{{if .returnGeometry}}/8/measurements.json{{else}}/8/gauges.json{{end}}",
	}, nil)
}

func TestNzhkb_ListGauges(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzhkb{
		name: "nzhkb",
		url:  ts.URL,
	}
	actual, err := s.ListGauges()
	expected := core.Gauges{
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzhkb",
				Code:   "24",
			},
			FlowUnit: "m3/s",
			Location: &core.Location{
				Latitude:  177.17031,
				Longitude: -38.74013,
			},
			Name: "Aniwaniwa Stream at Aniwaniwa",
			URL:  "http://data.hbrc.govt.nz/hydrotel/cgi-bin/hydwebserver.cgi/points/details?point=3612",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzhkb",
				Code:   "3649",
			},
			FlowUnit:  "m3/s",
			LevelUnit: "mm",
			Location: &core.Location{
				Latitude:  177.47985,
				Longitude: -38.80358,
			},
			Name: "Ruakituri River at Tauwharetoi Climate",
			URL:  "http://data.hbrc.govt.nz/hydrotel/cgi-bin/hydwebserver.cgi/points/details?point=3717",
		},
		core.Gauge{
			GaugeID: core.GaugeID{
				Script: "nzhkb",
				Code:   "4",
			},
			LevelUnit: "mm",
			Location: &core.Location{
				Latitude:  176.85914,
				Longitude: -39.48327,
			},
			Name: "Ahuriri Lagoon at Causeway",
			URL:  "http://data.hbrc.govt.nz/hydrotel/cgi-bin/hydwebserver.cgi/points/details?point=274",
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func TestNzhkb_Harvest(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	s := scriptNzhkb{
		name: "nzhkb",
		url:  ts.URL,
	}
	actual, err := core.HarvestSlice(&s, core.StringSet{}, 0)
	expected := core.Measurements{
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzhkb",
				Code:   "4",
			},
			Level:     nulltype.NullFloat64Of(10447),
			Timestamp: core.HTime{Time: time.Date(2020, time.June, 5, 0, 5, 0, 0, time.UTC)},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzhkb",
				Code:   "3649",
			},
			Flow:      nulltype.NullFloat64Of(28.718),
			Level:     nulltype.NullFloat64Of(1202),
			Timestamp: core.HTime{Time: time.Date(2020, time.June, 5, 0, 5, 0, 0, time.UTC)},
		},
		&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "nzhkb",
				Code:   "24",
			},
			Flow:      nulltype.NullFloat64Of(3.01),
			Timestamp: core.HTime{Time: time.Date(2020, time.June, 4, 18, 0, 0, 0, time.UTC)},
		},
	}
	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expected, actual)
	}
}
