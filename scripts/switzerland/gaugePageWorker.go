package switzerland

import (
	"strconv"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

const altOpen = "<th>Station elevation</th>"
const altClose = " m a.s.l."
const td = "<td class=\"text-right\">"

func parseAltitude(baseURL string, gauge *core.Gauge) {
	raw, err := core.Client.GetAsString(baseURL+gauge.GaugeID.Code+".html", nil)
	if err != nil || raw == "" {
		return
	}
	altStart := strings.Index(raw, altOpen)
	raw = raw[altStart+len(altOpen):]
	tdStart := strings.Index(raw, td)
	raw = raw[tdStart+len(td):]
	altEnd := strings.Index(raw, altClose)
	var alt float64
	if altEnd != -1 {
		altStr := raw[:altEnd]
		alt, _ = strconv.ParseFloat(altStr, 64)
		gauge.Location.Altitude = alt
	}
}

func gaugePageWorker(baseURL string, gauges <-chan *core.Gauge, results chan<- struct{}) {
	for gauge := range gauges {
		parseAltitude(baseURL, gauge)
		results <- struct{}{}
	}
}
