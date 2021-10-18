package canada

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/tz"
)

// Sometimes coordinate to timezone returns error (lakes, boundaries, dunno).
// In this case, fall back to these default timezones.
// Manually compiled mapping between provinces, UTC offsets and IANA timezome names
// https://en.wikipedia.org/wiki/Time_in_Canada
// https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
// https://www.countries-ofthe-world.com/time-zones-canada.html
var defTz = map[string]map[string]string{
	"AB": {
		"UTC-06:00": "America/Edmonton",
		"UTC-07:00": "America/Edmonton",
	},
	"BC": {
		"UTC-07:00": "America/Edmonton",
		"UTC-08:00": "America/Vancouver",
	},
	"MB": {
		"UTC-06:00": "America/Winnipeg",
	},
	"NB": {
		"UTC-04:00": "America/Moncton",
		"UTC-05:00": "America/New_York",
	},
	"NL": {
		"UTC-03:30": "America/St_Johns",
		"UTC-04:00": "America/Goose_Bay",
	},
	"NS": {
		"UTC-04:00": "America/Halifax",
	},
	"NT": {
		"UTC-07:00": "America/Inuvik",
	},
	"NU": {
		"UTC-07:00": "America/Cambridge_Bay",
	},
	"ON": {
		"UTC-05:00": "America/Toronto",
	},
	"PE": {
		"UTC-04:00": "America/Halifax",
	},
	"QC": {
		"UTC-05:00": "America/Toronto",
	},
	"SK": {
		"UTC-06:00": "America/Yellowknife",
		"UTC-07:00": "America/Swift_Current",
	},
	"YT": {
		"UTC-07:00": "America/Dawson",
		"UTC-08:00": "America/Whitehorse",
	},
}

func (s *scriptCanada) gaugeFromRow(line []string) (*core.Gauge, error) {
	lat, err := strconv.ParseFloat(line[2], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latitude '%s'", line[2])
	}
	lon, err := strconv.ParseFloat(line[3], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse longtitude '%s'", line[3])
	}
	zone, err := tz.CoordinateToTimezone(lat, lon)
	if err != nil {
		// fall back to manually picked timezone
		// in practice, this happens to few gauges in Ontario
		zone = defTz[line[4]][line[5]]
	}

	return &core.Gauge{
		GaugeID: core.GaugeID{
			Code:   line[0],
			Script: s.name,
		},
		Location: &core.Location{
			Longitude: lon,
			Latitude:  lat,
		},
		LevelUnit: "m",
		FlowUnit:  "m3/s",
		Name:      fmt.Sprintf("[%s] %s", line[4], strings.ReplaceAll(line[1], `"`, "")),
		URL:       "https://wateroffice.ec.gc.ca/report/real_time_e.html?stn=" + line[0],
		Timezone:  zone,
	}, nil

}
