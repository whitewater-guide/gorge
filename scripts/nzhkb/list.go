package nzhkb

import (
	"fmt"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

const (
	endpointLevels     = "8"
	endpointFlow       = "7"
	paramsGauges       = "where=1%3D1&outFields=ObjectID,Hilltop_Site,Hydrotel_Units,URL&outSR=4326&f=json"
	paramsMeasurements = "where=1%3D1&outFields=ObjectID,Hydrotel_LastSampleTime,Hydrotel_CurrentValue&returnGeometry=false&outSR=4326&f=json"
)

func (s *scriptNzhkb) fetchList(endpoint string, params string) ([]hkbFeature, error) {
	var data hkbList
	url := fmt.Sprintf("%s/%s/query?%s", s.url, endpoint, params)
	err := core.Client.GetAsJSON(url, &data, nil)
	if err != nil {
		return nil, err
	}
	return data.Features, nil
}

func (s *scriptNzhkb) fetchGauges() (core.Gauges, error) {
	flows, err := s.fetchList(endpointFlow, paramsGauges)
	if err != nil {
		return nil, err
	}
	levels, err := s.fetchList(endpointLevels, paramsGauges)
	if err != nil {
		return nil, err
	}
	byCode := map[int]core.Gauge{}
	newGauge := func(item hkbFeature) core.Gauge {
		return core.Gauge{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   fmt.Sprint(item.Attributes.ObjectID),
			},
			Name: item.Attributes.HilltopSite,
			URL:  item.Attributes.URL,
			Location: &core.Location{
				Latitude:  core.TruncCoord(item.Geometry.X),
				Longitude: core.TruncCoord(item.Geometry.Y),
			},
			Timezone: "Pacific/Auckland",
		}
	}
	for _, i := range levels {
		g := newGauge(i)
		g.LevelUnit = i.Attributes.HydrotelUnits
		byCode[i.Attributes.ObjectID] = g
	}
	for _, i := range flows {
		g, ok := byCode[i.Attributes.ObjectID]
		if !ok {
			g = newGauge(i)
		}
		g.FlowUnit = i.Attributes.HydrotelUnits
		if g.FlowUnit == "m³/s" {
			g.FlowUnit = "m3/s"
		}
		byCode[i.Attributes.ObjectID] = g
	}
	result := make(core.Gauges, len(byCode))
	j := 0
	for _, v := range byCode {
		result[j] = v
		j++
	}
	return result, nil
}

func (s *scriptNzhkb) fetchMeasurements(recv chan<- *core.Measurement, errs chan<- error) {
	flows, err := s.fetchList(endpointFlow, paramsMeasurements)
	if err != nil {
		errs <- err
		return
	}
	levels, err := s.fetchList(endpointLevels, paramsMeasurements)
	if err != nil {
		errs <- err
		return
	}
	byCode := map[int]core.Measurement{}
	newMeasurement := func(item hkbFeature) core.Measurement {
		return core.Measurement{
			GaugeID: core.GaugeID{
				Script: s.name,
				Code:   fmt.Sprint(item.Attributes.ObjectID),
			},
			Timestamp: core.HTime{Time: time.Unix(item.Attributes.HydrotelLastSampleTime/1000, 0).UTC()},
		}
	}
	for _, i := range levels {
		m := newMeasurement(i)
		m.Level = nulltype.NullFloat64Of(i.Attributes.HydrotelCurrentValue)
		byCode[i.Attributes.ObjectID] = m
	}
	for _, i := range flows {
		m, ok := byCode[i.Attributes.ObjectID]
		if !ok {
			m = newMeasurement(i)
		}
		m.Flow = nulltype.NullFloat64Of(i.Attributes.HydrotelCurrentValue)
		byCode[i.Attributes.ObjectID] = m
	}

	for _, v := range byCode {
		vv := v
		recv <- &vv
	}
}
