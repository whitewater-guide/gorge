package ukraine

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
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
				Name:     name,
				URL:      userURL,
				Timezone: "Europe/Kiev",
			},
			Timestamp: date.UTC(),
			Level:     nulltype.NullFloat64Of(level),
		}
	})

	return rivers, nil
}

func (s *scriptUkraine) harvest(measurements chan<- *core.Measurement, errs chan<- error) {
	var wg sync.WaitGroup
	waitHourlyData := map[string]bool{}
	for _, code := range s.station2code {
		waitHourlyData[code] = true
	}

	var rivers map[string]riverData
	riversReady := make(chan struct{})

	//load daily-updated measurements for all rivers
	//save measurements if it have no hourly-updated source
	wg.Add(1)
	go func() {
		defer func() {
			close(riversReady)
			wg.Done()
		}()
		var err error
		rivers, err = s.getAllRivers()
		if err != nil {
			errs <- err
			return
		}
		for _, river := range rivers {
			if !waitHourlyData[river.Code] {
				s.saveDailyMeasurements(measurements, river)
			}
		}
	}()

	//load&save hourly-updated measurements
	for station, code := range s.station2code {
		wg.Add(1)
		go func(code, station string) {
			defer wg.Done()
			ts := s.harvestSingle(code, station, measurements, errs)
			//wait loading daily-updated measurements
			//save as daily-updated if hourly-updated source 6 hours behind of daily-updated
			<-riversReady
			if rivers == nil {
				return
			}
			river, ok := rivers[code]
			if !ok {
				return
			}
			if ts.Add(6 * time.Hour).Before(river.Timestamp) {
				s.saveDailyMeasurements(measurements, river)
			}
		}(code, station)

	}

	wg.Wait()
}

func (s *scriptUkraine) saveDailyMeasurements(measurements chan<- *core.Measurement, data riverData) {
	measurements <- &core.Measurement{
		GaugeID:   data.GaugeID,
		Timestamp: core.HTime{Time: data.Timestamp},
		Level:     data.Level,
	}
	measurements <- &core.Measurement{
		GaugeID:   data.GaugeID,
		Timestamp: core.HTime{Time: data.Timestamp.Add(time.Hour)},
		Level:     data.Level,
	}
}

func (s *scriptUkraine) harvestSingle(code string, station string, measurements chan<- *core.Measurement, errs chan<- error) (lastTs time.Time) {
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
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   code,
				},
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
