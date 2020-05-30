package ukea

import (
	"strings"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

type reading struct {
	time     core.HTime
	levelMid string
	level    nulltype.NullFloat64
	flow     nulltype.NullFloat64
}

func (s *scriptUkea) getReadings(recv chan<- *core.Measurement, errs chan<- error) {
	readings := map[string]reading{}
	err := core.Client.StreamCSV(s.url+"/data/readings.csv?latest&_limit=10000", func(row []string) error {
		t, err := time.ParseInLocation("2006-01-02T15:04:05Z", row[0], time.UTC)
		if err != nil {
			return nil
		}
		mid, code := getMeasureID(row[1])
		var value nulltype.NullFloat64
		err = value.Scan(row[2])
		if err != nil {
			return nil
		}
		r, ok := readings[code]
		if !ok {
			r = reading{
				time: core.HTime{Time: t},
			}
		}
		if strings.HasPrefix(mid, "flow") {
			r.flow = value
		} else if strings.HasPrefix(mid, "level") && !strings.HasPrefix(r.levelMid, "level-stage-i-15_min") {
			r.levelMid = mid
			r.level = value
		}
		readings[code] = r
		return nil
	}, core.CSVStreamOptions{
		HeaderHeight: 1,
		NumColumns:   3,
	})
	if err != nil {
		errs <- err
		return
	}

	for k, v := range readings {
		if v.level.Valid() || v.flow.Valid() {
			recv <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   k,
				},
				Timestamp: v.time,
				Level:     v.level,
				Flow:      v.flow,
			}
		}
	}
}
