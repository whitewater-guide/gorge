package norway

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/PuerkitoBio/goquery"
	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/text/encoding/charmap"
)

type csvOptions struct {
	Header      int `desc:"CSV header height in rows" json:"header"`
	LevelColumn int `desc:"CSV level column index. Set to -1 if there's no such column" json:"levelColumn"`
	FlowColumn  int `desc:"CSV flow column index. Set to -1 if there's no such column" json:"flowColumn"`
	NumColumns  int `desc:"CSV columns number" json:"numColumns"`
}

type optionsNorway struct {
	Version int         `desc:"Gauge version number" json:"version"`
	HTML    bool        `desc:"Set to true to parse raw HTML instead of json. If CSV options are set, CSV parsing will be used instead" json:"html"`
	CSV     *csvOptions `desc:"Set to CSV instead of json." json:"csv"`
}
type scriptNorway struct {
	name          string
	urlBase       string
	jsonURLFormat string
	randomSeed    int64
	options       optionsNorway
	core.LoggingScript
}

func (s *scriptNorway) ListGauges() (core.Gauges, error) {
	resp, err := core.Client.Get(s.urlBase+"/list.html", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	utfReader := charmap.ISO8859_1.NewDecoder().Reader(resp.Body)

	doc, err := goquery.NewDocumentFromReader(utfReader)
	if err != nil {
		return nil, err
	}
	gauges := s.parseList(*doc)

	jobsCh := make(chan listItem, len(gauges))
	resultsCh := make(chan core.Gauge, len(gauges))
	results := make([]core.Gauge, len(gauges))

	for i := 1; i <= 5; i++ {
		go s.gaugePageWorker(jobsCh, resultsCh)
	}
	for _, g := range gauges {
		jobsCh <- g
	}
	close(jobsCh)
	for i := range gauges {
		results[i] = <-resultsCh
	}
	close(resultsCh)
	return results, nil
}

func (s *scriptNorway) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	code, err := codes.Only()
	if err != nil {
		errs <- err
		return
	}
	// Sometimes gauge JSON will contain message like "213.4.0 This station is not enabled for viewing - Ingen data"
	// In this case, we should set csv (preferred) or html flag to fallback
	if s.options.CSV != nil {
		errs <- s.harvestFromCSV(recv, code)
	} else if s.options.HTML {
		errs <- s.harvestFromPage(recv, code)
	} else {
		errs <- s.harvestFromJSON(recv, code, since)
	}
}

func (s *scriptNorway) harvestFromPage(recv chan<- *core.Measurement, code string) error {
	// http://www2.nve.no/h/hd/plotreal/Q/0213.00004.000/index.html
	parts := strings.Split(code, ".")
	url := fmt.Sprintf("%s/%04v.%05v.000/index.html", s.urlBase, parts[0], parts[1])
	raw, err := core.Client.GetAsString(url, nil)
	if err != nil {
		return err
	}
	page := s.parsePage(raw)
	m := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   code,
		},
		Timestamp: core.HTime{Time: page.timestamp.Time.UTC()},
		Flow:      page.value,
	}
	recv <- m
	return nil
}

func (s *scriptNorway) harvestFromCSV(recv chan<- *core.Measurement, code string) error {
	headerHeight := s.options.CSV.Header
	if headerHeight == 0 {
		headerHeight = 3
	}
	levelColumn := s.options.CSV.LevelColumn
	if levelColumn == 0 {
		levelColumn = 1
	}
	flowColumn := s.options.CSV.FlowColumn
	if flowColumn == 0 {
		flowColumn = 2
	}
	numColumns := s.options.CSV.NumColumns
	if numColumns == 0 {
		numColumns = 3
	}
	parts := strings.Split(code, ".")
	url := fmt.Sprintf("%s/%04v.%05v.000/knekkpunkt.csv", s.urlBase, parts[0], parts[1])
	return core.Client.StreamCSV(url, func(row []string) error {
		ts, err := time.Parse("2006-01-02 15:04", row[0])
		if err != nil {
			s.GetLogger().Errorf("failed to parse row's timestamp: %s", row[0])
			return nil
		}
		var level, flow nulltype.NullFloat64
		if levelColumn != -1 {
			err := level.UnmarshalJSON([]byte(strings.Replace(row[levelColumn], ",", ".", 1)))
			if err != nil {
				s.GetLogger().Errorf("failed to parse row's level: %s", row[levelColumn])
				return nil
			}
		}
		if flowColumn != -1 {
			err := flow.UnmarshalJSON([]byte(strings.Replace(row[flowColumn], ",", ".", 1)))
			if err != nil {
				s.GetLogger().Errorf("failed to parse row's flow: %s", row[flowColumn])
				return nil
			}
		}
		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   code,
			},
			Timestamp: core.HTime{
				Time: ts.UTC(),
			},
			Level: level,
			Flow:  flow,
		}
		return nil
	}, core.CSVStreamOptions{
		Comma:        ';',
		HeaderHeight: headerHeight,
		NumColumns:   numColumns,
	})
}

func (s *scriptNorway) getJSONUrl(code string, since int64) string {
	// this is for tests only
	if s.jsonURLFormat != "" {
		return fmt.Sprintf(s.jsonURLFormat, code, s.options.Version)
	}
	url := "http://h-web01.nve.no/chartserver/ShowData.aspx?req=getchart&ver=1.0&vfmt=json&time="
	// It seems that this endpoint cannot filter by hours, only by days
	// So "since" parameter has daily granularity.
	// E.g. if you want to filter values from 15:00 yesterday, you will still get all values from today and yesterday
	// If you want to filter values from 13:00 today, you will still get values from 06:00 today, but none from yesterday
	// I keep it here anyway
	var sinceStr string
	if since == 0 {
		sinceStr = "-1;0"
	} else {
		sinceT := time.Unix(since, 0)
		sinceStr = sinceT.UTC().Format("20060102T1504") + ";0"
	}
	paddedCode := fmt.Sprintf("%s.0.1001.%d", code, s.options.Version)
	url += sinceStr
	url += "&lang=no&chd=ds=htsr,da=29,id=" + paddedCode + ",rt=0"

	seed := s.randomSeed
	if seed == 0 {
		seed = time.Now().Unix()
	}
	r := rand.New(rand.NewSource(seed))
	url += fmt.Sprintf("&nocache=%d", r.Int31n(1000))

	return url
}

func (s *scriptNorway) harvestFromJSON(recv chan<- *core.Measurement, code string, since int64) error {
	url := s.getJSONUrl(code, since)
	resp, err := core.Client.GetAsString(url, nil)
	if err != nil {
		return err
	}
	return s.parseRawJSON(recv, code, resp)
}

func (s *scriptNorway) gaugePageWorker(gauges <-chan listItem, results chan<- core.Gauge) {
	for gauge := range gauges {
		resp, err := core.Client.Get(gauge.href, nil)
		result := core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   gauge.id,
			},
			Name:     gauge.name,
			FlowUnit: "m3/s",
			URL:      gauge.href,
			Location: &core.Location{
				Altitude: gauge.altitude,
			},
		}
		if err != nil {
			results <- core.Gauge{}
			continue
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			results <- core.Gauge{}
			resp.Body.Close()
			continue
		}
		page := s.parsePage(string(bytes))
		result.Location.Longitude = page.longitude
		result.Location.Latitude = page.latitude
		results <- result
		resp.Body.Close()
	}
}
