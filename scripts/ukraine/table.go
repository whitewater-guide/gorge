package ukraine

import (
	"crypto/md5"
	"fmt"
	"github.com/mattn/go-nulltype"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/whitewater-guide/gorge/core"
)

var nameRegex = regexp.MustCompile(`\W`)

func (s *scriptUkraine) parseTable(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
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
	doc.Find("Document Placemark").Each(func(i int, elem *goquery.Selection) {
		description, _ := elem.Find("description").First().Html()
		rName, _ := regexp.Compile(`Річка <b>(.+)</b><br/>`)
		rPost, _ := regexp.Compile(`Пост <b-->(.+)<br/>`)

		matchName := rName.FindStringSubmatch(description)
		matchPost := rPost.FindStringSubmatch(description)

		name := strings.TrimSpace(matchName[1]) + " " + strings.TrimSpace(matchPost[1])
		code := nameRegex.ReplaceAllString(name, "")
		code = strings.ToLower(code)
		code = fmt.Sprintf("%x", md5.Sum([]byte(code)))
		gauges <- &core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   code,
			},
			LevelUnit: "cm",
			Name:      name,
			URL:       s.url,
		}
		if measurements != nil {
			var level nulltype.NullFloat64
			rLevel, _ := regexp.Compile(`\d+\.\d+\.\d+: (\d+)см\.`)
			matchLevel := rLevel.FindStringSubmatch(description)
			levelStr := strings.TrimSpace(matchLevel[1])
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
