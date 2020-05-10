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
		// 7, 8 or 9 are discharge
		// 1,2,3,4,5,6,8,9,19,20,23,24,25,27,28,37,38,39,40, or 41 are level
		kind := st.LaitteistoID.Int64Value()
		if kind == 7 || kind == 8 || kind == 9 {
			lat, _ := strconv.ParseFloat(st.Lat, 64)
			lng, _ := strconv.ParseFloat(st.Lng, 64)
			url := ""
			if len(st.Nro) == 7 {
				url = fmt.Sprintf("https://wwwi2.ymparisto.fi/i2/%s/q%sy/wqfi.html", st.Nro[:2], st.Nro)
			}
			*gauges = append(*gauges, core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   fmt.Sprintf("%d", st.PaikkaID),
				},
				Name: fmt.Sprintf("%s - %s - %s", st.KuntaNimi, st.Nimi, st.Nro),
				URL:  url,
				// LevelUnit: "cm",
				FlowUnit: "m3/s",
				Location: &core.Location{
					Latitude:  core.TruncCoord(lat / 10000),
					Longitude: core.TruncCoord(lng / 10000),
				},
			})
		}
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
