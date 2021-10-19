package finland

import (
	"fmt"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptFinland) fetchList(url string, gauges *core.Gauges) error {
	var data stationsList
	err := core.Client.GetAsJSON(url, &data, nil)
	if err != nil {
		return err
	}
	for _, st := range data.Value {
		lon, lat, err := core.ToEPSG4326(st.X, st.Y, "EPSG:3067")
		if err != nil {
			continue
		}
		var levelUnit, flowUnit, unit string
		if st.SuureID == 2 {
			flowUnit = "m3/s"
			unit = "discharge"
		} else if st.SuureID == 1 {
			levelUnit = "cm"
			unit = "level"
		} else {
			continue
		}
		url := ""
		if len(st.Nro) == 7 {
			url = fmt.Sprintf("https://wwwi2.ymparisto.fi/i2/%s/q%sy/wqfi.html", st.Nro[:2], st.Nro)
		}
		*gauges = append(*gauges, core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   fmt.Sprintf("%d", st.PaikkaID),
			},
			Name:      fmt.Sprintf("%s - %s - %s (%s)", st.KuntaNimi, st.Nimi, st.Nro, unit),
			URL:       url,
			LevelUnit: levelUnit,
			FlowUnit:  flowUnit,
			Location: &core.Location{
				Latitude:  core.TruncCoord(lat),
				Longitude: core.TruncCoord(lon),
			},
			Timezone: "Europe/Helsinki",
		})
	}
	if data.Next != "" {
		next := data.Next
		if !strings.HasPrefix(next, "http") {
			next = s.url + next
		}
		return s.fetchList(next, gauges)
	}
	return nil
}
