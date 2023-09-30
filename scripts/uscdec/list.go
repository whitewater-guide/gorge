package uscdec

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/net/html"
)

var tz, _ = time.LoadLocation("US/Pacific")

func (s *scriptUSCDEC) parseList() (core.Measurements, error) {
	client := core.NewClient(core.ClientOptions{Timeout: 120}, s.GetLogger())
	doc, err := client.GetAsDoc(fmt.Sprintf("%s/getAll?sens_num=20", s.url), nil)
	if err != nil {
		return nil, err
	}
	result := core.Measurements{}
	doc.Find("div#tableContainer table.data tbody tr:has(td:nth-child(5))").Each(func(i int, q *goquery.Selection) {
		if q.HasClass("head") {
			return
		}
		if m, err := s.rowToMeasurement(q.Find("td").Nodes); err != nil {
			s.GetLogger().Warn(err)
		} else if m != nil {
			result = append(result, m)
		}
	})
	return result, nil
}

func (s *scriptUSCDEC) rowToMeasurement(row []*html.Node) (*core.Measurement, error) {
	if len(row) != 5 {
		return nil, nil
	}
	//  KLAMATH RIVER BELOW IRON GATE DAM KIG 2162' 09/30/2023 04:15 1,014 CFS
	//  PILARCITOS CREEK BL STONE DAM PIL 500' 09/30/2023 04:45 BRT CFS
	code, tStr, msmStr := strings.TrimSpace(getText(row[1])), strings.TrimSpace(getText(row[3])), strings.TrimSpace(getText(row[4]))

	if !strings.HasSuffix(msmStr, " CFS") {
		return nil, fmt.Errorf("measurement '%s' for gauge %s in not in CFS", msmStr, code)
	}
	msmStr = strings.TrimSuffix(msmStr, " CFS")
	msmStr = strings.ReplaceAll(msmStr, ",", "")
	cfs, err := strconv.Atoi(msmStr)
	if err != nil {
		return nil, fmt.Errorf("cannot parse measurement '%s' for gauge %s: %w", msmStr, code, err)
	}
	t, err := time.ParseInLocation("01/02/2006 15:04", tStr, tz)
	if err != nil {
		return nil, fmt.Errorf("cannot parse time '%s' for gauge %s: %w", tStr, code, err)
	}

	return &core.Measurement{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   code,
		},
		Timestamp: core.HTime{
			Time: t.UTC(),
		},
		Flow: nulltype.NullFloat64Of(float64(cfs)),
	}, nil
}

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getText(c)
	}
	return text
}
