package canada

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptCanada) gaugeFromRow(line []string) (*core.Gauge, error) {
	lat, err := strconv.ParseFloat(line[2], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latitude '%s'", line[2])
	}
	lon, err := strconv.ParseFloat(line[3], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse longtitude '%s'", line[3])
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
	}, nil

}
