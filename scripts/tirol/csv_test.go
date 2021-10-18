package tirol

import (
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestTirol_GetGauge(t *testing.T) {
	s := &scriptTirol{name: "tirol"}
	assert := assert.New(t)
	raw := csvRaw{
		name:      "Steeg",
		code:      "201012",
		river:     "Lech",
		parameter: "W.RADAR",
		timestamp: "2019-02-01T16:30:00+0100",
		value:     "215.5",
		unit:      "cm",
		elevation: "1109.3",
		easting:   "147008.0",
		northing:  "233674.0",
		epsg:      "EPSG:31257",
	}
	gauge, err := s.getGauge(raw)

	if assert.NoError(err) {
		assert.Equal(
			core.Gauge{
				GaugeID:   core.GaugeID{Script: "tirol", Code: "201012"},
				Name:      "Lech / Steeg",
				URL:       "https://apps.tirol.gv.at/hydro/#/Wasserstand/?station=201012",
				LevelUnit: "cm",
				FlowUnit:  "",
				Location:  &core.Location{Latitude: 47.24192, Longitude: 10.2935, Altitude: 1109},
				Timezone:  "Europe/Vienna",
			},
			gauge,
		)
	}
}

func TestTirol_GetMeasurement(t *testing.T) {
	assert := assert.New(t)
	s := &scriptTirol{name: "tirol"}
	raw := csvRaw{
		name:      "Steeg",
		code:      "201012",
		river:     "Lech",
		parameter: "W.RADAR",
		timestamp: "2019-02-01T16:30:00+0100",
		value:     "215.5",
		unit:      "cm",
		elevation: "1109.3",
		easting:   "147008.0",
		northing:  "233674.0",
		epsg:      "EPSG:31257",
	}
	loc, _ := time.LoadLocation("Europe/Vienna")
	m, err := s.getMeasurement(raw)

	if assert.NoError(err) {
		assert.Equal(m.GaugeID.Script, "tirol")
		assert.Equal(m.GaugeID.Code, "201012")
		assert.Equal(m.Level, nulltype.NullFloat64Of(215.5))
		assert.True(time.Date(2019, time.February, 1, 16, 30, 0, 0, loc).Equal(m.Timestamp.Time))
	}
}
