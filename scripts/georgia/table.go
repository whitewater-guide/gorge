package georgia

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/PuerkitoBio/goquery"
	"github.com/whitewater-guide/gorge/core"
)

var nameRegex = regexp.MustCompile(`\W`)

func (s *scriptGeorgia) parseTable(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	resp, err := core.Client.Get(s.url, nil)
	if err != nil {
		errs <- err
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		errs <- err
		return
	}
	doc.Find("td[background='images/hidro1.gif']").Each(func(i int, elem *goquery.Selection) {
		table := elem.ParentsFiltered("table").First()
		name := strings.TrimSpace(table.Find("b").First().Text())
		code := nameRegex.ReplaceAllString(name, "")
		code = strings.ToLower(code)
		code = fmt.Sprintf("%x", md5.Sum([]byte(code)))
		if gauges != nil {
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   code,
				},
				LevelUnit: "cm",
				Name:      name,
				URL:       s.url,
			}
		}
		if measurements != nil {
			var level nulltype.NullFloat64
			levelStr := strings.TrimSpace(table.Find(".date2").First().Text())
			err := level.UnmarshalJSON([]byte(levelStr))
			if err != nil {
				s.GetLogger().Errorf("failed to parse level string %s", levelStr)
				return
			}
			// I think gauges are updated at 6 a.m. UTC
			now := time.Now().UTC().Truncate(time.Hour)
			measurements <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   code,
				},
				Timestamp: core.HTime{Time: now},
				Level:     level,
			}
		}
	})
}
