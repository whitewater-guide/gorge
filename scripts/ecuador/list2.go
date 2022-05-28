package ecuador

import (
	"encoding/json"
	"math"

	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (s *scriptEcuador) parseList2() (map[string]core.Gauge, error) {
	raw, err := core.Client.GetAsString(s.listURL2, nil)
	if err != nil {
		return nil, err
	}
	var items []inamhiEmasItem

	err = json.Unmarshal([]byte(raw[2:len(raw)-1]), &items) // cut JSONP brackets
	if err != nil {
		return nil, err
	}

	result := make(map[string]core.Gauge)
	for _, item := range items {
		if item.Category != "HIDROLOGICA" {
			continue
		}
		result[item.Code] = core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   item.Code,
			},
			Name: cases.Title(language.Spanish).String(item.Name),
			Location: &core.Location{
				Latitude:  core.TruncCoord(item.Lat),
				Longitude: core.TruncCoord(item.Lng),
				Altitude:  math.Trunc(item.Alt),
			},
			LevelUnit: "m",
			Timezone:  "America/Guayaquil",
		}
	}

	return result, nil

}
