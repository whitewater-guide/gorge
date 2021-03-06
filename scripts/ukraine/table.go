package ukraine

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/PuerkitoBio/goquery"
	"github.com/whitewater-guide/gorge/core"
)

var client = core.NewClient(core.ClientOptions{
	UserAgent:  "whitewater.guide robot",
	Timeout:    60,
	WithoutTLS: true,
})

var userURL = "https://meteo.gov.ua/ua/33345/hydrostorm"

func (s *scriptUkraine) parseTable(gauges chan<- *core.Gauge, measurements chan<- *core.Measurement, errs chan<- error) {
	doc, err := client.GetAsDoc(s.url+"/kml_hydro_warn.kml", nil)
	if err != nil {
		errs <- err
		return
	}
	defer doc.Close()

	doc.Find("Document Placemark").Each(func(i int, elem *goquery.Selection) {
		code := elem.Find("name").First().Text()
		description, _ := elem.Find("description").First().Html()
		rName, _ := regexp.Compile(`Річка <b>(.+)</b><br/>`)
		rPost, _ := regexp.Compile(`Пост <b-->(.+)<br/>`)
		location := elem.Find("Point coordinates").First().Text()
		var locStr = strings.Split(location, ",")
		lng, _ := strconv.ParseFloat(locStr[0], 64)
		lat, _ := strconv.ParseFloat(locStr[1], 64)

		matchName := rName.FindStringSubmatch(description)
		matchPost := rPost.FindStringSubmatch(description)

		name := strings.TrimSpace(matchName[1]) + " " + strings.TrimSpace(matchPost[1])
		if gauges != nil {
			gauges <- &core.Gauge{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   code,
				},
				LevelUnit: "cm",
				Location: &core.Location{
					Latitude:  core.TruncCoord(lat),
					Longitude: core.TruncCoord(lng),
				},
				Name: name,
				URL:  userURL,
			}
		}
		if measurements != nil {
			var level nulltype.NullFloat64
			rLevel, _ := regexp.Compile(`\d+\.\d+\.\d+: (-?\d+)см\.`)
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
