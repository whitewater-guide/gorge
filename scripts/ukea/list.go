package ukea

import (
	"fmt"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

const (
	rloiWith    = 1
	rloiWithout = 2
)

func getName(st station) string {
	parts := []string{}
	if st.RLOIid != "" {
		parts = append(parts, st.RLOIid)
	}
	if st.RiverName != "" {
		parts = append(parts, strings.TrimSpace(st.RiverName))
	}
	if st.Label != "" {
		parts = append(parts, strings.TrimSpace(st.Label))
	}
	return strings.Join(parts, " - ")
}

func getMeasureID(url string) (measure string, station string) {
	parts := strings.Split(url, "/")
	parts = strings.Split(parts[len(parts)-1], "-")
	station = parts[0]
	measure = strings.Join(parts[1:], "-")
	return
}

// For flow units there's never more than 1 measure
// For level units, prefer level-stage-i-15_min, otherwise take the first available
func selectMeasures(measures []measure) (levelUnit string, flowUnit string) {
	levelOk := false
	for _, m := range measures {
		id, _ := getMeasureID(m.ID)
		if m.Parameter == "flow" {
			flowUnit = m.UnitName
		} else if m.Parameter == "level" && !levelOk {
			if strings.Contains(id, "groundwater") {
				continue
			}
			levelUnit = m.UnitName
			if strings.HasPrefix(id, "level-stage-i-15_min") && !strings.HasPrefix(id, "level-stage-i-15_min----") {
				levelOk = true
			}
		}
	}
	return
}

func (s *scriptUkea) fetchList() (core.Gauges, error) {
	gauges := core.Gauges{}
	var data stationsList
	err := core.Client.GetAsJSON(s.url+"/id/stations.json?_limit=10000", &data, nil)
	if err != nil {
		return nil, err
	}
	for _, st := range data.Stations {
		if st.RLOIid == "" && s.rloi == rloiWith || st.RLOIid != "" && s.rloi == rloiWithout {
			continue
		}
		g := core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   st.StationReference,
			},
			Name:      getName(st),
			URL:       fmt.Sprintf("https://environment.data.gov.uk/flood-monitoring/id/stations/%s.html", st.StationReference),
			LevelUnit: "m",
			FlowUnit:  "m3/s",
			Location: &core.Location{
				Latitude:  core.TruncCoord(st.Lat),
				Longitude: core.TruncCoord(st.Long),
			},
			Timezone: "Europe/London",
		}
		g.LevelUnit, g.FlowUnit = selectMeasures(st.Measures)
		if g.LevelUnit != "" || g.FlowUnit != "" {
			gauges = append(gauges, g)
		}
	}
	return gauges, nil
}
