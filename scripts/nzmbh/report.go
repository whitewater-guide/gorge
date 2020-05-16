package nzmbh

import (
	"strconv"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptNzmbh) fetchReport(recv chan<- *core.Measurement, errs chan<- error) {
	var report riverReport
	err := core.Client.GetAsJSON(s.reportURL, &report, nil)
	if err != nil {
		errs <- err
		return
	}
	for _, e := range report {
		var flow, level nulltype.NullFloat64
		if f, err := strconv.ParseFloat(e.Flow, 64); err == nil {
			flow = nulltype.NullFloat64Of(f)
		}
		if l, err := strconv.ParseFloat(e.Stage, 64); err == nil {
			level = nulltype.NullFloat64Of(l)
		}
		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   codeFromName(e.SiteName),
			},
			Timestamp: core.HTime{
				Time: e.LastUpdate.Time,
			},
			Level: level,
			Flow:  flow,
		}
	}
}
