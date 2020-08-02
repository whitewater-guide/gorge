package kuban

import (
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"
	"golang.org/x/text/encoding/charmap"

	"github.com/PuerkitoBio/goquery"
	"github.com/whitewater-guide/gorge/core"
)

var tz, _ = time.LoadLocation("Europe/Moscow")
var decoder = charmap.Windows1251.NewDecoder()

func (s *scriptKuban) parseTable(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	doc, err := core.Client.GetAsDoc(s.url, nil)
	if err != nil {
		errs <- err
		return
	}
	defer doc.Close()

	ts := core.HTime{}

	doc.Find("tr").EachWithBreak(func(i int, elem *goquery.Selection) bool {
		if i == 0 { // header row, parse date and time ("02.08.2020 08 ч.мск")
			tstr := elem.Children().Eq(2).Text()
			tstr = tstr[0 : len(tstr)-6]
			t, err := time.ParseInLocation("02.01.2006 15", tstr, tz)
			if err != nil {
				errs <- err
				return false
			}
			ts.Time = t.UTC()
		}
		if i < 3 { // couple of emty rows
			return true
		}
		nodes := elem.Find("td")
		if nodes.Length() < 9 {
			return true
		}
		if measurements != nil {
			var level nulltype.NullFloat64
			err := level.Scan(strings.ReplaceAll(nodes.Eq(2).Text(), ",", "."))
			if err == nil {
				measurements <- &core.Measurement{
					GaugeID: core.GaugeID{
						Script: s.name,
						Code:   nodes.Eq(0).Text(),
					},
					Timestamp: ts,
					Level:     level,
				}
			}
		}
		if gauges != nil {
			td0 := strings.TrimSpace(strings.ReplaceAll(nodes.Eq(1).Text(), "\n", ""))
			nb, err := decoder.Bytes([]byte(td0))
			if err != nil {
				return true
			}
			name := strings.ReplaceAll(string(nb), string(rune(160)), "")
			name = strings.ReplaceAll(name, "  ", " ")
			lat, err := strconv.ParseFloat(strings.Replace(nodes.Eq(6).Text(), ",", ".", -1), 64)
			if err != nil {
				return true
			}
			lng, err := strconv.ParseFloat(strings.Replace(nodes.Eq(7).Text(), ",", ".", -1), 64)
			if err != nil {
				return true
			}
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   nodes.Eq(0).Text(),
				},
				Name:      name,
				URL:       "",
				LevelUnit: "cm",
				Location: &core.Location{
					Latitude:  lat,
					Longitude: lng,
				},
			}
		}
		return true
	})

}
