package wales

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

type item struct {
	gauge       *core.Gauge
	measurement *core.Measurement
}

func (s *scriptWales) fetchList(path string, gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	key := s.options.Key
	if key == "" {
		key = os.Getenv("WALES_KEY")
	}
	if key == "" {
		errs <- fmt.Errorf("wales api key not found")
		return
	}
	var data walesData
	err := core.Client.GetAsJSON(
		s.url+path,
		&data,
		&core.RequestOptions{
			Headers: map[string]string{"Ocp-Apim-Subscription-Key": key},
		},
	)

	if err != nil {
		errs <- err
		return
	}

	for _, feat := range data.Features {
		if gauges != nil {
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   feat.Properties.Location,
				},
				Name: feat.Properties.TitleEN,
				Location: &core.Location{
					Latitude:  core.TruncCoord(feat.Geometry.Coordinates[1]),
					Longitude: core.TruncCoord(feat.Geometry.Coordinates[0]),
				},
				LevelUnit: feat.Properties.Units,
				URL:       strings.Replace(feat.Properties.URL, "http://", "https://", 1),
			}
		}
		if measurements != nil {
			var level nulltype.NullFloat64
			err := level.UnmarshalJSON([]byte(feat.Properties.LatestValue))
			if err != nil {
				continue
			}
			measurements <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   feat.Properties.Location,
				},
				Level:     level,
				Timestamp: core.HTime{Time: feat.Properties.LatestTime.Time},
			}
		}
	}
}
