package finland

import (
	"fmt"
	"strconv"
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
		lat, _ := strconv.ParseFloat(st.Lat, 64)
		lng, _ := strconv.ParseFloat(st.Lng, 64)
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
				Latitude:  core.TruncCoord(lat / 10000),
				Longitude: core.TruncCoord(lng / 10000),
			},
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
