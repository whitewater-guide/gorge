package chile

import (
	"fmt"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

const wgs84 = "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext  +no_defs"

func (s *scriptChile) getWebGauges() (map[string]core.Gauge, error) {
	webmap, err := s.parseWebmap()
	if err != nil {
		return nil, err
	}

	listedGauges, err := s.getListedGauges()
	if err != nil {
		return nil, err
	}
	gaugeIds := make([]string, 0)
	for k := range listedGauges {
		gaugeIds = append(gaugeIds, k)
	}

	numGauges := len(gaugeIds)
	usefulness := make(map[string]bool)
	for i := 0; i < numGauges; i += 3 {
		err := s.areGaugesUseful(gaugeIds[i:i+3], usefulness)
		if err != nil {
			return nil, err
		}
	}

	result := make(map[string]core.Gauge)

	for id, ok := range usefulness {
		if !ok {
			continue
		}
		if feature, ok := webmap[id]; ok {
			x, y, err := core.ToEPSG4326(feature.Geometry.X, feature.Geometry.Y, wgs84)
			if err != nil {
				return nil, fmt.Errorf("failed to convert coordinates from mercator: %w", err)
			}
			result[id] = core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   id,
				},
				Name:      strings.Title(strings.ToLower(feature.Attributes.Name)),
				FlowUnit:  "m3/s",
				LevelUnit: "m",
				Location: &core.Location{
					Longitude: x,
					Latitude:  y,
				},
			}
		} else {
			// gauge is not present on webmap, so we cannot get it's Location and nice name
			result[id] = core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   id,
				},
				Name:      listedGauges[id],
				FlowUnit:  "m3/s",
				LevelUnit: "m",
			}
		}
	}
	return result, nil
}
