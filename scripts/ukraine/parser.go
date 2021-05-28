package ukraine

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const userURL = "https://meteo.gov.ua/ua/33345/hydrostorm"

var client = core.NewClient(core.ClientOptions{
	UserAgent:  "whitewater.guide robot",
	Timeout:    60,
	WithoutTLS: true,
}, nil)

type riverData struct {
	core.Gauge
	Timestamp time.Time
	Level     nulltype.NullFloat64
}

func getTimezone() *time.Location {
	name := "Europe/Kiev"
	tz, err := time.LoadLocation(name)
	if err != nil {
		time.FixedZone(name, 2*3600)
	}
	return tz
}

var rDateLevel = regexp.MustCompile(`на (\d+ год. \d+\.\d+\.\d+): (-?\d+)см\.`)
var rName = regexp.MustCompile(`Річка <b>(.+)</b><br/>`)
var rPost = regexp.MustCompile(`Пост <b-->(.+)<br/>`)

func (s *scriptUkraine) getAllRivers() (map[string]riverData, error) {
	doc, err := client.GetAsDoc(s.urlDaily+"/kml_hydro_warn.kml", nil)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	rivers := map[string]riverData{}

	doc.Find("Document Placemark").Each(func(i int, elem *goquery.Selection) {
		code := elem.Find("name").First().Text()
		description, _ := elem.Find("description").First().Html()
		location := elem.Find("Point coordinates").First().Text()
		var locStr = strings.Split(location, ",")
		lng, _ := strconv.ParseFloat(locStr[0], 64)
		lat, _ := strconv.ParseFloat(locStr[1], 64)

		matchName := rName.FindStringSubmatch(description)
		matchPost := rPost.FindStringSubmatch(description)

		name := strings.TrimSpace(matchName[1]) + " " + strings.TrimSpace(matchPost[1])

		matchLevel := rDateLevel.FindStringSubmatch(description)
		levelStr := strings.TrimSpace(matchLevel[2])
		dateStr := strings.TrimSpace(matchLevel[1])
		level, err := strconv.ParseFloat(levelStr, 64)
		if err != nil {
			s.GetLogger().Errorf("failed to parse level string %s", levelStr)
			return
		}
		date, err := time.ParseInLocation("15 год. 02.01.2006", dateStr, s.timezone)
		if err != nil {
			s.GetLogger().Errorf("failed to parse date/time string %s", levelStr)
			return
		}

		rivers[code] = riverData{
			Gauge: core.Gauge{
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
			},
			Timestamp: date.UTC(),
			Level:     nulltype.NullFloat64Of(level),
		}
	})

	return rivers, nil
}

func (s *scriptUkraine) harvest(measurements chan<- *core.Measurement, errs chan<- error) {
	rivers, err := s.getAllRivers()
	if err != nil {
		errs <- err
		return
	}

	var wg sync.WaitGroup
	lastMeasurements := map[core.GaugeID]time.Time{}
	var lastMeasurementsLock sync.Mutex

	for station, code := range s.station2code {
		river, ok := rivers[code]
		if !ok {
			return
		}
		wg.Add(1)
		go func(station string) {
			ts := s.harvestSingle(river, station, measurements, errs)

			lastMeasurementsLock.Lock()
			lastMeasurements[river.GaugeID] = ts
			lastMeasurementsLock.Unlock()

			wg.Done()
		}(station)

	}
	wg.Wait()

	for _, river := range rivers {
		if river.Timestamp.After(lastMeasurements[river.GaugeID]) {
			measurements <- &core.Measurement{
				GaugeID:   river.GaugeID,
				Timestamp: core.HTime{Time: river.Timestamp},
				Level:     river.Level,
			}
		}
	}

}

func (s *scriptUkraine) harvestSingle(riverInfo riverData, station string, measurements chan<- *core.Measurement, errs chan<- error) (lastTs time.Time) {
	doc, err := s.getMeasurementsAsDoc(station)
	if err != nil {
		errs <- err
		return
	}

	doc.Find("tr.table_result2, tr.table_result3").Each(func(i int, elem *goquery.Selection) {
		cel := elem.Find("td")
		dateStr := strings.TrimSpace(cel.First().Text())
		date, err := time.ParseInLocation("02.01.2006", dateStr, s.timezone)
		if err != nil {
			errs <- fmt.Errorf("failed to parse date string %s (station: %s)", dateStr, station)
			return
		}
		date = date.UTC()
		for hour := 0; hour < 24; hour++ {
			t := date.Add(time.Duration(hour) * time.Hour)
			levelStr := strings.TrimSpace(cel.Eq(hour + 1).Text())
			if levelStr == "-" {
				continue
			}
			level, err := strconv.ParseFloat(levelStr, 64)
			if err != nil {
				errs <- fmt.Errorf("failed to parse level string %s (station: %s)", levelStr, station)
				return
			}
			measurements <- &core.Measurement{
				GaugeID:   riverInfo.GaugeID,
				Timestamp: core.HTime{Time: t},
				Level:     nulltype.NullFloat64Of(level),
			}
			if t.After(lastTs) {
				lastTs = t
			}
		}
	})

	return
}

func (s *scriptUkraine) getMeasurementsAsDoc(station string) (*goquery.Document, error) {
	now := time.Now().In(s.timezone)
	params := url.Values{}
	params.Set("station", station)
	params.Add("date_start", now.Add(-24*time.Hour).Format("02-01-2006"))
	params.Add("date_end", now.Format("02-01-2006"))
	dataUrl := s.urlHourly
	if s.addStation2url {
		dataUrl += "/" + station
	}
	resp, _, err := client.PostForm(dataUrl, params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
