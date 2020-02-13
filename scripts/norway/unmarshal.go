package norway

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/PuerkitoBio/goquery"
	"github.com/whitewater-guide/gorge/core"
)

type listItem struct {
	name     string
	href     string
	id       string
	altitude float64
}

type gaugePage struct {
	latitude  float64
	longitude float64
	timestamp nTime
	value     nulltype.NullFloat64
}

type rawJSON []struct {
	SeriesPoints []struct {
		Key   nTime
		Value nulltype.NullFloat64
	}
}

type nTime struct {
	time.Time
}

func (nt *nTime) UnmarshalJSON(b []byte) (err error) {
	// This look like "/Date(1519056000000)/" (with quotes)
	// This is in nanoseconds, so cut last 3 zeroes
	s := b[8 : len(b)-7]
	i, err := strconv.ParseInt(string(s), 10, 64)
	t := time.Unix(i, 0).UTC()
	nt.Time = t
	return
}

func (s *scriptNorway) parseList(doc goquery.Document) []listItem {
	var result []listItem
	table := doc.Find("table").First()
	table.Find("tr").Slice(1, goquery.ToEnd).Each(func(i int, tr *goquery.Selection) {
		cols := tr.Find("td")
		altStr := cols.Eq(2).Text()
		altStr = strings.TrimSpace(altStr)
		altStr = strings.Replace(altStr, "m", "", 1)
		altitude, _ := strconv.ParseFloat(altStr, 64)
		result = append(result, listItem{
			name:     cols.Eq(0).Text(),
			href:     s.urlBase + "/" + cols.Eq(0).Find("a").First().AttrOr("href", ""),
			id:       cols.Eq(1).Text(),
			altitude: altitude,
		})
	})
	return result
}

func (s *scriptNorway) parsePage(data string) gaugePage {
	var result gaugePage
	// Find lastTimestamp and value (flow)
	// <center>Siste m�ling, tid=20.02.2018 06:00, verdi=  2.619</center><br>
	tsInd := strings.Index(data, "tid=")
	if tsInd >= 0 {
		data = string(data[tsInd+len("tid="):])
		tsEnd := strings.Index(data, ",")
		tsStr := string(data[:tsEnd])
		ts, _ := time.Parse("02.01.2006 15:04", tsStr)
		result.timestamp = nTime{Time: ts.UTC()}

		valInd := strings.Index(data, "verdi=")
		data = string(data[valInd+len("verdi="):])
		valEnd := strings.Index(data, "<")
		valStr := string(data[:valEnd])

		result.value.UnmarshalJSON([]byte(strings.TrimSpace(valStr))) //nolint:errcheck
	}

	// Breddegrad = latitude Lengdegrad = longitude
	lngInd := strings.Index(data, "Lengdegrad: <B>")
	if lngInd >= 0 {
		data = string(data[lngInd+len("Lengdegrad: <B>"):])
		lngEnd := strings.Index(data, "<")
		lngStr := string(data[:lngEnd])
		result.longitude, _ = strconv.ParseFloat(strings.TrimSpace(lngStr), 64)
		result.longitude = core.TruncCoord(result.longitude)

		latInd := strings.Index(data, "Breddegrad: <B>")
		data = string(data[latInd+len("Breddegrad: <B>"):])
		latEnd := strings.Index(data, "<")
		latStr := string(data[:latEnd])
		result.latitude, _ = strconv.ParseFloat(strings.TrimSpace(latStr), 64)
		result.latitude = core.TruncCoord(result.latitude)
	}

	return result
}

func (s *scriptNorway) parseRawJSON(recv chan<- *core.Measurement, code string, rawStr string) error {
	var res rawJSON
	err := json.Unmarshal([]byte(rawStr), &res)
	if err != nil {
		return err
	}
	points := res[0].SeriesPoints
	for _, p := range points {
		recv <- &core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   code,
			},
			Timestamp: core.HTime{Time: p.Key.Time},
			Flow:      p.Value,
		}
	}
	return nil
}
