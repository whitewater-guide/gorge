package nztrc

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

var tz, _ = time.LoadLocation("Pacific/Auckland")
var timeRegExp = regexp.MustCompile(`River \w+ (\d\d:\d\d[ap]m)`)

type station struct {
	level *site
	flow  *site
}

func (s *scriptNztrc) fetchList(measureID string, out *map[int]station) error {
	var list []site
	err := core.Client.GetAsJSON(s.url+measureID, &list, nil)
	if err != nil {
		return err
	}

	for _, i := range list {
		ii := i
		st, ok := (*out)[i.SiteID]
		if !ok {
			st = station{}
		}
		if measureID == "9" {
			st.flow = &ii
		} else {
			st.level = &ii
		}
		(*out)[i.SiteID] = st
	}
	return nil
}

func (s *scriptNztrc) parseList(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	stations := map[int]station{}
	err := s.fetchList("9", &stations)
	if err != nil {
		errs <- err
		return
	}
	err = s.fetchList("7", &stations)
	if err != nil {
		errs <- err
		return
	}
	for code, st := range stations {
		var common site
		var levelUnit, flowUnit string
		var level, flow nulltype.NullFloat64
		if st.level != nil {
			levelUnit = st.level.Unit
			common = (*st.level)
			level.Scan((*st.level).Measure) //nolint:errcheck
		}
		if st.flow != nil {
			flowUnit = st.flow.Unit
			common = (*st.flow)
			flow.Scan((*st.flow).Measure) //nolint:errcheck
		}
		if flowUnit == "m3/sec" {
			flowUnit = "m3/s"
		}
		if gauges != nil {
			lat, err := strconv.ParseFloat(common.Lat, 64)
			if err != nil {
				continue
			}
			lng, err := strconv.ParseFloat(common.Lng, 64)
			if err != nil {
				continue
			}
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   fmt.Sprint(code),
				},
				Name:      common.Title,
				URL:       common.Link,
				LevelUnit: levelUnit,
				FlowUnit:  flowUnit,
				Location: &core.Location{
					Latitude:  core.TruncCoord(lat),
					Longitude: core.TruncCoord(lng),
				},
				Timezone: "Pacific/Auckland",
			}
		}
		if measurements != nil {
			matches := timeRegExp.FindAllStringSubmatch(common.Description, -1)
			if len(matches) != 1 || len(matches[0]) != 2 {
				continue
			}
			now := time.Now().In(tz)
			t, err := time.ParseInLocation("15:04pm", matches[0][1], tz)
			if err != nil {
				fmt.Println(err)
				continue
			}
			measurements <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   fmt.Sprint(code),
				},
				Timestamp: core.HTime{
					Time: time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, tz).UTC(),
				},
				Level: level,
				Flow:  flow,
			}
		}
	}
}
