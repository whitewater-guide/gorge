package wales

import (
	"fmt"
	"os"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptWales) fetchList(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	key := s.options.Key
	if key == "" {
		key = os.Getenv("WALES_KEY")
	}
	if key == "" {
		errs <- fmt.Errorf("wales api key not found")
		return
	}
	var data []stationData
	err := core.Client.GetAsJSON(
		s.url,
		&data,
		&core.RequestOptions{
			Headers: map[string]string{"Ocp-Apim-Subscription-Key": key},
		},
	)

	if err != nil {
		errs <- err
		return
	}

	for _, feat := range data {
		var param *stationParam
		for _, p := range feat.Parameters {
			if p.ParamNameEN == "River Level" {
				param = &p
			}
		}
		if param == nil {
			continue
		}

		if gauges != nil {
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   fmt.Sprint(feat.Location),
				},
				Name: feat.TitleEn,
				Location: &core.Location{
					Latitude:  core.TruncCoord(feat.Coordinates.Latitude),
					Longitude: core.TruncCoord(feat.Coordinates.Longitude),
				},
				LevelUnit: param.Units,
				URL:       "https://" + feat.URL,
			}
		}
		if measurements != nil {
			var level = nulltype.NullFloat64Of(param.LatestValue)
			if err != nil {
				continue
			}
			measurements <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   fmt.Sprint(feat.Location),
				},
				Level:     level,
				Timestamp: core.HTime{Time: param.LatestTime.Time},
			}
		}
	}
}
