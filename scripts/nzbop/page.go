package nzbop

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

var tz, _ = time.LoadLocation("Pacific/Auckland")

func parseRow(td *goquery.Selection) (v nulltype.NullFloat64, m string, t time.Time, err error) {
	valStr := td.Next().Text()
	ts := td.Next().Next().Text()
	valParts := strings.Split(strings.TrimSpace(valStr), " ")
	vf, err := strconv.ParseFloat(strings.TrimSpace(valParts[0]), 64)
	if err != nil {
		return
	}
	v = nulltype.NullFloat64Of(vf)
	m = valParts[1]
	if m == "metres" {
		m = "m"
	}
	// Mon Jan 2 15:04:05 -0700 MST 2006
	t, err = time.ParseInLocation("2-January-2006 15:04", strings.TrimSpace(ts), tz)
	if err != nil {
		return
	}
	t = t.UTC()
	return
}

func (s *scriptBop) parsePage(code string, gauges chan<- *core.Gauge, measurements chan<- *core.Measurement) {
	url := fmt.Sprintf(s.pageURL, code)
	resp, err := core.Client.Get(url, nil)
	if err != nil {
		s.GetLogger().WithField("code", code).Error("failed to fetch page")
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		s.GetLogger().WithField("code", code).Error("failed to parse page")
		return
	}
	var level, flow nulltype.NullFloat64
	var levelUnit, flowUnit string
	var t time.Time
	var name string
	nameNode := doc.Find(`h1`)
	if nameNode.Length() != 0 {
		name = nameNode.Text()
	}
	levelNode := doc.Find(`b:contains("River Level")`)
	if levelNode.Length() == 0 {
		levelNode = doc.Find(`b:contains("Water Level")`)
	}
	if levelNode.Length() == 0 {
		levelNode = doc.Find(`b:contains("Level")`)
	}
	if levelNode.Length() != 0 {
		td := levelNode.ParentsUntil("tr")
		level, levelUnit, t, _ = parseRow(td)
	}
	flowNode := doc.Find(`b:contains("Flow")`)
	if flowNode.Length() != 0 {
		td := flowNode.ParentsUntil("tr")
		flow, flowUnit, t, _ = parseRow(td)
	}
	if levelUnit == "" && flowUnit == "" {
		return
	}
	if gauges != nil {
		location := &core.Location{}
		locNode := doc.Find(`th:contains("Grid Reference:")`)
		if locNode.Length() != 0 {
			gridRef := strings.TrimSpace(locNode.Next().Text())
			location, _ = convertNZMS260(gridRef)
		} else {
			s.GetLogger().WithField("code", code).Warn("location not found")
		}
		altNode := doc.Find(`th:contains("Elevation")`)
		if altNode.Length() != 0 {
			altStr := strings.TrimSpace(altNode.Next().Text())
			alt, _ := strconv.ParseFloat(altStr, 64)
			if location != nil {
				location.Altitude = alt
			}
		}
		gauges <- &core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   code,
			},
			Name:      strings.TrimSpace(name),
			URL:       url,
			LevelUnit: levelUnit,
			FlowUnit:  flowUnit,
			Location:  location,
		}
	} else {
		measurements <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   code,
			},
			Timestamp: core.HTime{
				Time: t,
			},
			Level: level,
			Flow:  flow,
		}
	}
}

func (s *scriptBop) gaugePageWorker(codes <-chan string, results chan<- *core.Gauge, wg *sync.WaitGroup) {
	for code := range codes {
		s.parsePage(code, results, nil)
	}
	wg.Done()
}
