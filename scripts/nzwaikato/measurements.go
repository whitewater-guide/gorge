package nzwaikato

import (
	"net/url"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

var tz, _ = time.LoadLocation("Pacific/Auckland")

func (s *scriptWaikato) parseMeasurements(measurements chan<- *core.Measurement) error {
	resp, err := core.Client.Get(s.listURL, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Sometimes this return document with empty body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	doc.Find("#RainfallTable tr").Each(func(i int, elem *goquery.Selection) {

		if i == 0 {
			return
		}
		href, _ := elem.Find("a").First().Attr("href")
		link, err := url.Parse(href)
		if err != nil {
			return
		}
		tds := elem.Find("td")
		tstr := tds.Eq(1).Text()
		t, err := time.ParseInLocation("02/01/06 15:04", tstr, tz)
		if err != nil {
			return
		}
		var level, flow nulltype.NullFloat64
		var levelStr, flowStr = tds.Eq(2).Text(), tds.Eq(3).Text()
		if levelStr != "" {
			levelF, err := strconv.ParseFloat(levelStr, 64)
			if err != nil {
				return
			}
			level = nulltype.NullFloat64Of(levelF)
		}
		if flowStr != "" {
			flowF, err := strconv.ParseFloat(flowStr, 64)
			if err != nil {
				return
			}
			flow = nulltype.NullFloat64Of(flowF)
		}

		measurements <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   link.Query().Get("point"),
			},
			Timestamp: core.HTime{
				Time: t.UTC(),
			},
			Level: level,
			Flow:  flow,
		}

	})
	return nil
}
