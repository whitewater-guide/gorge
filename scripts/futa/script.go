package futa

import (
	"bufio"
	"context"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

type optionsFuta struct{}

type scriptFuta struct {
	name    string
	dataURL string
	core.LoggingScript
}

func (s *scriptFuta) ListGauges() (result core.Gauges, err error) {
	result = append(result, core.Gauge{
		GaugeID:  core.GaugeID{Script: "futa", Code: "futa00"},
		Name:     "Futaleufu Hidroelectrica",
		URL:      "http://www.chfutaleufu.com.ar/default.asp",
		FlowUnit: "m3/s",
		Timezone: "America/Santiago",
	})
	return
}

func (s *scriptFuta) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)

	tz, err := time.LoadLocation("America/Santiago")
	if err != nil {
		return
	}

	resp, err := core.Client.Get(s.dataURL, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	lNum := 0
	for scanner.Scan() {
		var d time.Time
		var v nulltype.NullFloat64
		line := scanner.Text()
		if strings.Contains(line, "Estacion Meteorologica Futaleufu") {
			break
		}
		lNum += 1
		if lNum <= 5 || len(line) == 0 {
			continue
		}
		parts := strings.Split(line, "#")
		dS := parts[0][1 : len(parts[0])-1]
		vS := strings.TrimSpace(parts[12])
		//"02 Jan 06 15:04 MST"
		if d, err = time.ParseInLocation("15:04:05 01/02/06", dS, tz); err != nil {
			break
		}
		v.UnmarshalJSON([]byte(vS))
		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   "futa00",
			},
			Timestamp: core.HTime{
				Time: d.UTC(),
			},
			Flow: v,
		}
	}

	if err != nil {
		errs <- err
	}
}
