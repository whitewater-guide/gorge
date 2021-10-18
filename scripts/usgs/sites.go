package usgs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/tz"
)

func (s *scriptUSGS) listStations(flow bool, gauges map[string]core.Gauge) error {
	// Select parameter https://help.waterdata.usgs.gov/parameter_cd?group_cd=PHY
	parameterCd, levelUnit, flowUnit := paramLevel, "ft", "" // Gage height, feet
	if flow {
		parameterCd, levelUnit, flowUnit = paramFlow, "", "ft3/s" // Discharge, cubic feet per second
	}
	return core.Client.StreamCSV(
		fmt.Sprintf("%s/site/?format=rdb&stateCd=%s&siteType=ST&parameterCd=%s&siteStatus=all&hasDataTypeCd=iv", s.url, s.stateCd, parameterCd),
		func(row []string) error {
			g, ok := gauges[row[1]]

			if ok {
				if flow {
					g.FlowUnit = flowUnit
				} else {
					g.LevelUnit = levelUnit
				}
			} else {
				lat, err := strconv.ParseFloat(row[4], 64)
				if err != nil {
					return nil
				}
				lng, err := strconv.ParseFloat(row[5], 64)
				if err != nil {
					return nil
				}
				zone, err := tz.CoordinateToTimezone(lat, lng)
				if err != nil {
					zone = "UTC"
				}
				alt, _ := strconv.ParseFloat(strings.TrimSpace(row[8]), 64)
				g = core.Gauge{
					GaugeID: core.GaugeID{
						Script: s.name,
						Code:   row[1],
					},
					Name:      row[2],
					URL:       fmt.Sprintf("https://waterdata.usgs.gov/nwis/inventory?agency_code=%s&site_no=%s", row[0], row[1]),
					LevelUnit: levelUnit,
					FlowUnit:  flowUnit,
					Location: &core.Location{
						Latitude:  core.TruncCoord(lat),
						Longitude: core.TruncCoord(lng),
						Altitude:  alt,
					},
					Timezone: zone,
				}
			}
			gauges[row[1]] = g
			return nil
		},
		core.CSVStreamOptions{
			Comma:        '\t',
			NumColumns:   12,
			HeaderHeight: 29,
		},
	)
}
