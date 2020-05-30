package nzwgn

import (
	"fmt"
	"strconv"

	"github.com/mattn/go-nulltype"

	"github.com/whitewater-guide/gorge/core"
)

type dataItem struct {
	core.Measurement
	levelUnit string
	flowUnit  string
	name      string
}

// Another possible option is "WQ%20/%20Rivers%20and%20Streams" but I think it's useless
const collection = "River%20and%20Stream%20Levels"

func (s *scriptNzwgn) fetchLocations() (map[string]core.Location, error) {
	url := fmt.Sprintf("%s?Service=Hilltop&Request=SiteList&Collection=%s&Location=LatLong", s.url, collection)
	var list siteList
	err := core.Client.GetAsXML(url, &list, nil)
	if err != nil {
		return nil, err
	}
	result := make(map[string]core.Location, len(list.Site))
	for _, v := range list.Site {
		// because lat/long is bad site doesn't stop to exist
		lat, _ := strconv.ParseFloat(v.Latitude, 64)  //nolint:errcheck
		lng, _ := strconv.ParseFloat(v.Longitude, 64) //nolint:errcheck
		result[core.CodeFromName(v.Name)] = core.Location{
			Latitude:  core.TruncCoord(lat),
			Longitude: core.TruncCoord(lng),
		}
	}
	return result, nil
}

func (s *scriptNzwgn) fetchValues() (map[string]dataItem, error) {
	url := fmt.Sprintf("%s?Service=Hilltop&Request=GetData&Collection=%s&Location=LatLong", s.url, collection)
	var data measurements
	err := core.Client.GetAsXML(url, &data, nil)
	if err != nil {
		return nil, err
	}
	result := map[string]dataItem{}
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
		unit := i.DataSource.ItemInfo.Units
		if unit == "m³/sec" {
			unit = "m3/s"
		}
		m, ok := result[i.SiteName]
		if !ok {
			m = dataItem{
				Measurement: core.Measurement{
					GaugeID: core.GaugeID{
						Script: s.name,
						Code:   core.CodeFromName(i.SiteName),
					},
					Timestamp: core.HTime{
						Time: i.Data.E.T.Time,
					},
				},
			}

		}
		if !m.Level.Valid() && level.Valid() {
			m.levelUnit = unit
			m.Level = level
		}
		if !m.Flow.Valid() && flow.Valid() {
			m.flowUnit = unit
			m.Flow = flow
		}
		m.name = i.SiteName
		result[i.SiteName] = m
	}

	return result, nil
}
