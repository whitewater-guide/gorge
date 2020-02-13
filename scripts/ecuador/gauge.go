package ecuador

import (
	"fmt"
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/whitewater-guide/gorge/core"
)

func findIndices(response *ecuadorRoot) (int, int, error) {
	dateIndex, valueIndex := -1, -1
	for i, v := range response.Head.Fields {
		if v.Name == "FechaHora" {
			dateIndex = i
			continue
		}
		if v.Name == "nivelSmp" {
			valueIndex = i
			continue
		}
	}
	if dateIndex == -1 {
		return dateIndex, valueIndex, fmt.Errorf("date field not found")
	}
	if valueIndex == -1 {
		return dateIndex, valueIndex, fmt.Errorf("level field not found")
	}
	return dateIndex, valueIndex, nil
}

func (s *scriptEcuador) parseMeasurement(raw []interface{}, code string, dateInd, valueInd int) (*core.Measurement, error) {
	if dateInd >= len(raw) {
		return nil, fmt.Errorf("date index is outside of vals range")
	}
	if valueInd >= len(raw) {
		return nil, fmt.Errorf("value index is outside of vals range")
	}

	dateRaw := raw[dateInd]
	valueRaw := raw[valueInd]

	dateStr, ok := dateRaw.(string)
	if !ok {
		return nil, fmt.Errorf("vals element at date index is not a string")
	}
	value, ok := valueRaw.(float64)
	if !ok {
		return nil, fmt.Errorf("vals element at value index is not a float")
	}

	t, err := time.ParseInLocation("20060102150405", dateStr, time.UTC)
	if err != nil {
		return nil, err
	}
	return &core.Measurement{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   code,
		},
		Level:     nulltype.NullFloat64Of(value),
		Timestamp: core.HTime{Time: t},
	}, nil
}

func (s *scriptEcuador) parseGauge(recv chan<- *core.Measurement, errs chan<- error, code string) {
	ts := time.Now().In(time.UTC).UnixNano() / int64(time.Millisecond)
	resp := &ecuadorRoot{}
	err := core.Client.GetAsJSON(fmt.Sprintf(s.gaugeURLFormat, code, ts), resp, nil)

	if err != nil {
		errs <- err
		return
	}

	dateInd, valueInd, err := findIndices(resp)
	if err != nil {
		errs <- err
		return
	}

	for _, rawData := range resp.Data {
		m, err := s.parseMeasurement(rawData.Vals, code, dateInd, valueInd)
		if err != nil {
			s.GetLogger().Error(err)
			continue
		}
		recv <- m
	}

}
