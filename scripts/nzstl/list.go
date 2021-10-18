package nzstl

import (
	"strconv"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

// EPSG:2193 NZGD2000, http://epsg.io/2193
const epsg2193 = "+proj=tmerc +lat_0=0 +lon_0=173 +k=0.9996 +x_0=1600000 +y_0=10000000 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs"

func (s *scriptNzstl) fetchList(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	var data list
	err := core.Client.GetAsJSON(s.url, &data, nil)
	if err != nil {
		errs <- err
		return
	}
	for _, site := range data.Sites {
		if gauges != nil {
			x, _ := strconv.ParseFloat(site.Easting, 64)
			y, _ := strconv.ParseFloat(site.Northing, 64)
			lng, lat, _ := core.ToEPSG4326(x, y, epsg2193)
			var flowUnit, levelUnit string
			if site.Flow.Measurement != "" {
				flowUnit = "m3/s"
			}
			if site.WaterLevel.Measurement != "" {
				levelUnit = "m"
			}
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   core.CodeFromName(site.Site),
				},
				Name:      site.Site,
				URL:       "http://envdata.es.govt.nz/",
				LevelUnit: levelUnit,
				FlowUnit:  flowUnit,
				Location: &core.Location{
					Latitude:  lat,
					Longitude: lng,
				},
				Timezone: "Pacific/Auckland",
			}
		}
		if measurements != nil {
			var flow, level nulltype.NullFloat64
			if site.Flow.Value != "" {
				flow.Scan(site.Flow.Value) //nolint:errcheck
			}
			if site.WaterLevel.Value != "" {
				level.Scan(site.WaterLevel.Value) //nolint:errcheck
			}
			measurements <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   core.CodeFromName(site.Site),
				},
				Timestamp: core.HTime{
					Time: site.DataTo.UTC(),
				},
				Level: level,
				Flow:  flow,
			}
		}
	}
}
