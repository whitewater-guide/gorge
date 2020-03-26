package russia1

import (
	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptRussia1) fetchList(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	list := &russia1Features{}
	err := core.Client.GetAsJSON(s.gaugesURL, list, nil)
	if err != nil {
		errs <- err
		return
	}
	for _, f := range list.Features {
		if gauges != nil {
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   f.Properties.Name,
				},
				Name:      f.Properties.Name,
				URL:       "http://www.emercit.com/map/",
				LevelUnit: "m",
				Location: &core.Location{
					Latitude:  f.Geometry.Coordinates[1],
					Longitude: f.Geometry.Coordinates[0],
				},
			}
		}
		if measurements != nil && !f.Properties.Data.RiverLevel.Time.IsZero() {
			measurements <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   f.Properties.Name,
				},
				Timestamp: core.HTime{
					Time: f.Properties.Data.RiverLevel.Time.Time.UTC(),
				},
				Level: f.Properties.Data.RiverLevel.Level.Bs,
			}
		}
	}
}
