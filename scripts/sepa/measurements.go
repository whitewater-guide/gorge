package sepa

import (
	"encoding/json"
	"fmt"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

type DataPoint struct {
	Timestamp core.HTime           `json:"timestamp"`
	Value     nulltype.NullFloat64 `json:"value"`
}

func (dp *DataPoint) UnmarshalJSON(data []byte) error {
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) != 2 {
		return fmt.Errorf("data point must have exactly 2 elements, got %d", len(raw))
	}

	// Parse timestamp (first element)
	timestampStr, ok := raw[0].(string)
	if !ok {
		return fmt.Errorf("timestamp must be a string, got %T", raw[0])
	}

	var timestamp core.HTime
	if err := timestamp.UnmarshalJSON([]byte(`"` + timestampStr + `"`)); err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}
	dp.Timestamp = timestamp

	// Parse value (second element)
	switch v := raw[1].(type) {
	case float64:
		dp.Value = nulltype.NullFloat64Of(v)
	case int:
		dp.Value = nulltype.NullFloat64Of(float64(v))
	case int64:
		dp.Value = nulltype.NullFloat64Of(float64(v))
	case nil:
		dp.Value = nulltype.NullFloat64{}
	default:
		return fmt.Errorf("value must be a number or null, got %T", raw[1])
	}

	return nil
}

type SEPAStationMeasurement struct {
	StationNo string `json:"station_no"`
	// TsUnitSymbol         string      `json:"ts_unitsymbol"`
	StationParameterName string      `json:"stationparameter_name"`
	Rows                 string      `json:"rows"`
	Columns              string      `json:"columns"`
	Data                 []DataPoint `json:"data"`
}

// UnmarshalJSON implements json.Unmarshaler interface for SEPAStationMeasurement
func (sm *SEPAStationMeasurement) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to unmarshal the raw data
	type tempSEPAStationMeasurement struct {
		StationNo string `json:"station_no"`
		// TsUnitSymbol         string          `json:"ts_unitsymbol"`
		StationParameterName string          `json:"stationparameter_name"`
		Rows                 string          `json:"rows"`
		Columns              string          `json:"columns"`
		Data                 [][]interface{} `json:"data"`
	}

	var temp tempSEPAStationMeasurement
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy the simple fields
	sm.StationNo = temp.StationNo
	// sm.TsUnitSymbol = temp.TsUnitSymbol
	sm.StationParameterName = temp.StationParameterName
	sm.Rows = temp.Rows
	sm.Columns = temp.Columns

	// Parse the data points
	sm.Data = make([]DataPoint, len(temp.Data))
	for i, rawPoint := range temp.Data {
		// Marshal the raw point back to JSON for DataPoint.UnmarshalJSON
		pointJSON, err := json.Marshal(rawPoint)
		if err != nil {
			return fmt.Errorf("failed to marshal data point %d: %w", i, err)
		}

		if err := sm.Data[i].UnmarshalJSON(pointJSON); err != nil {
			return fmt.Errorf("failed to parse data point %d: %w", i, err)
		}
	}

	return nil
}

// SEPAStationMeasurements represents the full response from SEPA API
type SEPAStationMeasurements []SEPAStationMeasurement
