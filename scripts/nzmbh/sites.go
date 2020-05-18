package nzmbh

import (
	"encoding/xml"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/whitewater-guide/gorge/core"
)

var nameRegex = regexp.MustCompile(`\W`)
var space = uuid.MustParse("344d640b-2569-4b47-ab4e-1541b23b864f")

func codeFromName(name string) string {
	code := nameRegex.ReplaceAllString(strings.TrimSpace(name), "")
	code = strings.ToLower(code)
	return uuid.NewMD5(space, []byte(code)).String()
}

type site struct {
	loc  core.Location
	name string
}

func (s *scriptNzmbh) fetchSiteList() (map[string]site, error) {
	req, _ := http.NewRequest("GET", s.siteListURL, nil)
	resp, err := core.Client.Do(req, nil)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	list := &siteList{}
	err = xml.NewDecoder(resp.Body).Decode(list)
	if err != nil {
		return nil, err
	}
	result := make(map[string]site, len(list.FeatureMember))
	for _, m := range list.FeatureMember {
		llStr := strings.Split(m.SiteList.Location.Point.Pos, " ")
		if len(llStr) != 2 {
			continue
		}
		lat, err := strconv.ParseFloat(llStr[0], 64)
		if err != nil {
			continue
		}
		lng, err := strconv.ParseFloat(llStr[1], 64)
		if err != nil {
			continue
		}
		result[codeFromName(m.SiteList.Site)] = site{
			loc: core.Location{
				Latitude:  core.TruncCoord(lat),
				Longitude: core.TruncCoord(lng),
			},
			name: m.SiteList.Site,
		}
	}
	return result, nil
}

func (s *scriptNzmbh) genGauge(sites map[string]site, m *core.Measurement) *core.Gauge {
	site, siteOk := sites[m.Code]
	if !siteOk {
		return nil
	}
	var lu, fu string
	if m.Level.Valid() {
		lu = "m"
	}
	if m.Flow.Valid() {
		fu = "m3/s"
	}
	return &core.Gauge{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   m.Code,
		},
		Name:      site.name,
		URL:       "http://hydro.marlborough.govt.nz/reports/riverreport.html",
		LevelUnit: lu,
		FlowUnit:  fu,
		Location:  &site.loc,
	}
}
