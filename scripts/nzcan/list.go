package nzcan

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

const (
	markers0 = "var markers = ["
	markers1 = "initMap(markers, options);"
)

var tz, _ = time.LoadLocation("Pacific/Auckland")

type geoloc struct {
	location *core.Location
	name     string
}

func getTime(str string) (*core.HTime, error) {
	year := time.Now().In(tz).Year()
	ext := fmt.Sprintf("%d %s:00", year, str)
	ts, err := time.ParseInLocation("2006 02-January 15:04:05", ext, tz)
	if err != nil {
		return nil, err
	}
	return &core.HTime{
		Time: ts.UTC(),
	}, nil
}

func (s *scriptNzcan) fetchGeo() (map[string]geoloc, error) {
	txt, err := core.Client.GetAsString(s.url+"/RiverflowGeo/ALL", &core.RequestOptions{
		Headers: map[string]string{
			"X-Requested-With": "XMLHttpRequest",
		},
	})
	if err != nil {
		return nil, err
	}
	m0 := strings.Index(txt, markers0)
	if m0 == -1 {
		return nil, errors.New("markers not found")
	}
	txt = txt[m0+len(markers0)-1:]
	m1 := strings.Index(txt, markers1)
	if m1 == -1 {
		return nil, errors.New("markers end not found")
	}
	txt = txt[:m1-4]
	var markers markersList
	err = json.Unmarshal([]byte(txt), &markers)
	if err != nil {
		return nil, err
	}
	result := make(map[string]geoloc, len(markers))
	for _, m := range markers {
		result[m.SiteNo] = geoloc{
			location: &core.Location{
				Latitude:  core.TruncCoord(m.Lat),
				Longitude: core.TruncCoord(m.Lng),
			},
			name: m.SiteName,
		}
	}
	return result, nil
}

func (s *scriptNzcan) fetchList(suffix string, recv chan<- *core.Measurement) error {
	doc, err := core.Client.GetAsDoc(s.url+"/RiverflowList/"+suffix, &core.RequestOptions{
		Headers: map[string]string{
			"X-Requested-With": "XMLHttpRequest",
		},
	})
	if err != nil {
		return err
	}
	defer doc.Close()

	doc.Find("tr.riverflow-" + suffix[:1]).Each(func(i int, elem *goquery.Selection) {
		tds := elem.Find("td")
		link := elem.Find("th").Find("a").First()
		// name := strings.TrimSpace(link.Text())
		href, _ := link.Attr("href")
		code := href[strings.LastIndex(href, "/")+1:]
		ts, err := getTime(tds.Eq(0).Text())
		if err != nil {
			return
		}
		var level, flow nulltype.NullFloat64
		levelStr, flowStr := tds.Eq(1).Text(), tds.Eq(2).Text()
		if levelStr != "" {
			level.Scan(levelStr)
		}
		if flowStr != "" {
			flow.Scan(flowStr)
		}
		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   code,
			},
			Timestamp: *ts,
			Level:     level,
			Flow:      flow,
		}
	})
	return nil
}
