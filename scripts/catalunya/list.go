package catalunya

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptCatalunya) fetchList() ([]sensor, error) {
	list := &catalunyaList{}
	err := core.Client.GetAsJSON(s.gaugesURL, list, nil)
	if err != nil {
		return nil, err
	}
	return list.Providers[0].Sensors, err
}

func (s *scriptCatalunya) convert(sensor *sensor) (*core.Gauge, error) {
	var locStr = strings.Split(sensor.Location, " ")
	if len(locStr) != 2 {
		return nil, errors.New("failed to parse location " + sensor.Location)
	}
	lat, err := strconv.ParseFloat(locStr[0], 64)
	if err != nil {
		return nil, errors.New("failed to parse latitude of " + sensor.Location)
	}
	lng, err := strconv.ParseFloat(locStr[1], 64)
	if err != nil {
		return nil, errors.New("failed to parse longitude of " + sensor.Location)
	}
	var levelUnit, flowUnit string
	switch sensor.Type {
	case "0014": // m3/s
		flowUnit = sensor.Unit
	case "0035": // l/s
		flowUnit = sensor.Unit
	case "0019": // cm
		levelUnit = sensor.Unit
	}
	return &core.Gauge{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   sensor.Sensor,
		},
		Name:      strings.Title(strings.ToLower(sensor.ComponentAdditionalInfo.Riu)) + " - " + sensor.ComponentDesc + " (" + levelUnit + flowUnit + ")",
		URL:       "http://aca-web.gencat.cat/sentilo-catalog-web/component/AFORAMENT-EST." + sensor.Component + "/detail",
		LevelUnit: levelUnit,
		FlowUnit:  flowUnit,
		Location: &core.Location{
			Latitude:  core.TruncCoord(lat),
			Longitude: core.TruncCoord(lng),
		},
	}, nil
}

func (s *scriptCatalunya) parseList() (core.Gauges, error) {
	sensors, err := s.fetchList()

	if err != nil {
		return nil, err
	}

	var result core.Gauges
	for _, sensor := range sensors {
		if strings.Contains(strings.ToLower(sensor.Description), "canal") {
			continue
		}
		g, err := s.convert(&sensor)
		if err != nil {
			return nil, err
		}
		result = append(result, *g)
	}

	return result, nil
}

var flowSensors map[string]bool

func (s *scriptCatalunya) isFlowSensor(sensor *dataSensor) (bool, error) {
	if flowSensors == nil {
		sensors, err := s.fetchList()
		if err != nil {
			return false, err
		}
		flowSensors = make(map[string]bool)
		for _, sr := range sensors {
			flowSensors[sr.Sensor] = sr.Type != "0019" // cm is the only level type
		}
	}
	result, ok := flowSensors[sensor.Sensor]
	if !ok {
		return false, fmt.Errorf("sensor with code %s not found", sensor.Sensor)
	}
	return result, nil
}
