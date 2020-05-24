package nzwgn

import (
	"fmt"
	"strconv"

	"github.com/mattn/go-nulltype"

	"github.com/whitewater-guide/gorge/core"
)

type site struct {
	loc  core.Location
	name string
}

// Another possible option is "WQ%20/%20Rivers%20and%20Streams" but I think it's useless
const collection = "River%20and%20Stream%20Levels"

func (s *scriptNzwgn) fetchLocations(out *map[string]core.Location) error {
	url := fmt.Sprintf("%s&Request=SiteList&Collection=%s&Location=LatLong", s.url, collection)
	var list siteList
	err := core.Client.GetAsXML(url, &list, nil)
	if err != nil {
		return err
	}
	for _, v := range list.Site {
		// because lat/long is bad site doesn't stop to exist
		lat, _ := strconv.ParseFloat(v.Latitude, 64)  //nolint:errcheck
		lng, _ := strconv.ParseFloat(v.Longitude, 64) //nolint:errcheck
		(*out)[core.CodeFromName(v.Name)] = core.Location{
			Latitude:  lat,
			Longitude: lng,
		}
	}
	return nil
}

func (s *scriptNzwgn) fetchValues() (map[string]core.Measurement, error) {
	url := fmt.Sprintf("%s&Request=GetData&Collection=%s&Location=LatLong", s.url, collection)
	var data measurements
	err := core.Client.GetAsXML(url, &data, nil)
	if err != nil {
		return nil, err
	}
	result := map[string]core.Measurement{}
	for _, i := range data.Measurement {
		var flow, level nulltype.NullFloat64
		switch i.DataSource.ItemInfo.ItemName {
		case "Flow":
			flow.Scan(i.Data.E.I1) //nolint:errcheck
		case "Stage":
			level.Scan(i.Data.E.I1) //nolint:errcheck
		default:
			continue
		}
		m, ok := result[i.SiteName]
		if !ok {
			m = core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   core.CodeFromName(i.SiteName),
				},
				Timestamp: core.HTime{
					Time: i.Data.E.T.Time,
				},
			}
		}
		if !m.Level.Valid() && level.Valid() {
			m.Level = level
		}
		if !m.Flow.Valid() && flow.Valid() {
			m.Flow = flow
		}
		result[i.SiteName] = m
	}
	return result, nil
}
