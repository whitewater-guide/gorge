package uscdec

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptUSCDEC) parseDetails(ctx context.Context, code string) (*core.Gauge, error) {
	doc, err := core.Client.GetAsDoc(fmt.Sprintf("%s/staMeta?station_id=%s", s.url, code), nil)
	if err != nil {
		return nil, err
	}
	table := doc.Find("table").First()
	if table == nil {
		return nil, errors.New("table not found")
	}
	header := doc.Find("h2").First()
	if header == nil {
		return nil, errors.New("header not found")
	}
	g := core.Gauge{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   code,
		},
		Name:     trimTxt(header),
		URL:      fmt.Sprintf("%s/staMeta?station_id=%s", s.url, code),
		Timezone: tz.String(),
		FlowUnit: "cfs",
		Location: &core.Location{},
	}

	var loopErr error
	table.Find("td").EachWithBreak(func(i int, td *goquery.Selection) bool {
		txt := trimTxt(td)
		switch txt {
		case "Station ID":
			{
				if nextTd := td.Next(); nextTd != nil {
					actualCode := trimTxt(nextTd)
					if code != actualCode {
						loopErr = fmt.Errorf("found code %s, expected %s", actualCode, code)
						return false
					}
				}
			}
		case "Elevation":
			{
				if nextTd := td.Next(); nextTd != nil {
					elev := trimTxt(nextTd, " ft")
					ft, _ := strconv.ParseFloat(elev, 64)
					g.Location.Altitude = ft * 0.3048
				}
			}
		case "Latitude":
			{
				if nextTd := td.Next(); nextTd != nil {
					latStr := trimTxt(nextTd, "°")
					if g.Location.Latitude, loopErr = strconv.ParseFloat(latStr, 64); loopErr != nil {
						loopErr = fmt.Errorf("failed to parse latitude '%s': %w", latStr, loopErr)
						return false
					}
				}
			}
		case "Longitude":
			{
				if nextTd := td.Next(); nextTd != nil {
					longStr := trimTxt(nextTd, "°")
					if g.Location.Longitude, loopErr = strconv.ParseFloat(longStr, 64); loopErr != nil {
						loopErr = fmt.Errorf("failed to parse latitude '%s': %w", longStr, loopErr)
						return false
					}
				}
			}
		}

		return true
	})
	if loopErr != nil {
		return nil, loopErr
	}
	g.Location.Altitude = core.TruncCoord(g.Location.Altitude)
	g.Location.Latitude = core.TruncCoord(g.Location.Latitude)
	g.Location.Longitude = core.TruncCoord(g.Location.Longitude)

	return &g, nil
}

func trimTxt(selection *goquery.Selection, suffix ...string) string {
	out := strings.TrimSpace(selection.Text())
	for _, sfx := range suffix {
		out = strings.TrimSuffix(out, sfx)
	}
	return out
}
