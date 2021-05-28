package ukraine

import (
	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/testutils"
	"testing"
	"time"
)

func setupScript() (script core.Script, cls func()) {
	ts := testutils.SetupFileServer(map[string]string{
		"/kml_hydro_warn.kml": "kml_hydro_warn.kml",
		"/chornogolova":       "chornogolova.html",
		"/tatariv":            "tatariv.html",
	}, nil)
	cls = ts.Close
	s := &scriptUkraine{
		name:           "ukraine",
		urlDaily:       ts.URL,
		urlHourly:      ts.URL,
		timezone:       getTimezone(),
		addStation2url: true,
		station2code: map[string]string{
			"chornogolova": "44120",
			"tatariv":      "42136",
		},
	}
	script = s
	return
}

func TestUkraine_ListGauges(t *testing.T) {
	s, cls := setupScript()
	defer cls()

	actual, err := s.ListGauges()
	expected := core.Gauge{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "42136",
		},
		LevelUnit: "cm",
		Location: &core.Location{
			Latitude:  48.36876,
			Longitude: 24.55166,
		},
		Name: "Прут Татарів",
		URL:  "https://meteo.gov.ua/ua/33345/hydrostorm",
	}
	if assert.NoError(t, err) {
		assert.Len(t, actual, 192)
		assert.Contains(t, actual, expected)
	}
}

func TestUkraine_Harvest(t *testing.T) {
	s, cls := setupScript()
	defer cls()

	actual, err := core.HarvestSlice(s, core.StringSet{}, 0)
	expected1 := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "42136",
		},
		Timestamp: core.HTime{
			Time: time.Date(2021, 5, 21, 5, 0, 0, 0, time.UTC),
		},
		Level: nulltype.NullFloat64Of(145),
	}
	expected2 := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "44120",
		},
		Timestamp: core.HTime{
			Time: time.Date(2021, 5, 21, 10, 0, 0, 0, time.UTC),
		},
		Level: nulltype.NullFloat64Of(20),
	}
	expected3 := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "44120",
		},
		Timestamp: core.HTime{
			Time: time.Date(2021, 5, 21, 11, 0, 0, 0, time.UTC),
		},
		Level: nulltype.NullFloat64Of(19),
	}
	expected4 := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "ukraine",
			Code:   "44087",
		},
		Timestamp: core.HTime{
			Time: time.Date(2021, 5, 21, 5, 0, 0, 0, time.UTC),
		},
		Level: nulltype.NullFloat64Of(-32),
	}
	if assert.NoError(t, err) {
		assert.Contains(t, actual, expected1)
		assert.Contains(t, actual, expected2)
		assert.Contains(t, actual, expected3)
		assert.Contains(t, actual, expected4)
	}
}
