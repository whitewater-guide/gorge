package nzwaikato

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
)

const epsg2193 = "+proj=tmerc +lat_0=0 +lon_0=173 +k=0.9996 +x_0=1600000 +y_0=10000000 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs"

var locRegExp = regexp.MustCompile(`NZTM:\s*(\d+)\s*-\s*(\d+)`)
var spaces = regexp.MustCompile(`\s+`)

func (s *scriptWaikato) parseGaugePage(code string, hasFlow, hasLevel bool) (*core.Gauge, error) {
	url := fmt.Sprintf(s.pageURL, code)
	resp, err := core.Client.Get(url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	name := doc.Find(`td:contains("Site name:")`).Next().Text()
	locStr := strings.TrimSpace(doc.Find(`td:contains("Map reference:")`).Next().Text())
	var lat, lon float64
	if locStr != "" {
		match := locRegExp.FindAllStringSubmatch(locStr, -1)
		xStr, yStr := match[0][1], match[0][2]
		x, err := strconv.ParseFloat(xStr, 64)
		if err != nil {
			return nil, err
		}
		y, err := strconv.ParseFloat(yStr, 64)
		if err != nil {
			return nil, err
		}
		lon, lat, err = core.ToEPSG4326(x, y, epsg2193)
		if err != nil {
			return nil, err
		}
	}
	var levelUnit, flowUnit string
	if hasLevel {
		levelUnit = "m"
	}
	if hasFlow {
		flowUnit = "m3/s"
	}
	return &core.Gauge{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   code,
		},
		Name:      strings.TrimSpace(spaces.ReplaceAllString(name, " ")),
		URL:       url,
		LevelUnit: levelUnit,
		FlowUnit:  flowUnit,
		Location: &core.Location{
			Latitude:  lat,
			Longitude: lon,
		},
	}, nil
}

func (s *scriptWaikato) gaugePageWorker(ms <-chan *core.Measurement, results chan<- *core.Gauge, wg *sync.WaitGroup) {
	for m := range ms {
		code := m.Code
		gauge, err := s.parseGaugePage(code, m.Flow.Valid(), m.Level.Valid())
		if err != nil {
			fmt.Println(err)
			s.GetLogger().WithFields(logrus.Fields{
				"script":  s.name,
				"command": "harvest",
				"code":    code,
			}).Error(err)
			continue
		}
		results <- gauge
	}
	wg.Done()
}
