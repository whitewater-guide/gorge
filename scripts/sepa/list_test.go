package sepa

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestSepa_GetGauge(t *testing.T) {
	assert := assert.New(t)
	raw := csvRaw{
		sepaHydrologyOffice:   "Perth",
		stationName:           "Perth",
		locationCode:          "10048",
		nationalGridReference: "NO1160525332",
		catchmentName:         "---",
		riverName:             "Tay",
		gaugeDatum:            "2.08",
		catchmentArea:         "4991.0",
		startDate:             "Aug-91",
		endDate:               "07/04/2019 06:45",
		systemID:              "58156010",
		lowestValue:           "0.0",
		low:                   "0.161",
		maxValue:              "4.928",
		high:                  "3.493",
		maxDisplay:            "4.928m @ 17/01/1993 19:30:00",
		mean:                  "0.884",
		units:                 "m",
		webMessage:            "",
		nrfaLink:              "https://nrfa.ceh.ac.uk/data/station/info/15042",
	}

	s := &scriptSepa{name: "sepa"}
	result, err := s.getGauge(raw)

	if assert.NoError(err) {
		assert.Equal(core.Gauge{
			GaugeID: core.GaugeID{
				Code:   "10048",
				Script: "sepa",
			},
			Name:      "Tay - Perth",
			LevelUnit: "m",
			FlowUnit:  "",
			Location: &core.Location{
				Latitude:  56.41191,
				Longitude: -3.4342,
				Altitude:  2,
			},
			URL:      "http://apps.sepa.org.uk/waterlevels/default.aspx?sd=t&lc=10048",
			Timezone: "Europe/London",
		},
			result,
		)
	}
}
