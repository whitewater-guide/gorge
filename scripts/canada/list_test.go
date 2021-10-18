package canada

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestCanada_GaugeFromRow(t *testing.T) {
	s := scriptCanada{name: "canada"}
	actual, err := s.gaugeFromRow([]string{
		"02YA002",
		`"BARTLETTS RIVER NEAR ST. ANTHONY"`,
		"51.449220",
		"-55.641250",
		"NL",
		"UTC-03:30",
	})
	expected := &core.Gauge{
		GaugeID: core.GaugeID{
			Script: "canada",
			Code:   "02YA002",
		},
		Name:      "[NL] BARTLETTS RIVER NEAR ST. ANTHONY",
		URL:       "https://wateroffice.ec.gc.ca/report/real_time_e.html?stn=02YA002",
		LevelUnit: "m",
		FlowUnit:  "m3/s",
		Location: &core.Location{
			Latitude:  51.44922,
			Longitude: -55.64125,
		},
		Timezone: "America/St_Johns",
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
