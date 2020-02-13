package core

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/mattn/go-nulltype"
)

func unixHTime(sec int64) HTime {
	return HTime{time.Unix(sec, 0)}
}

// GaugesCodes is helper method used to get codes from job description
func GaugesCodes(m map[string]json.RawMessage) StringSet {
	res := StringSet{}
	for k := range m {
		res[k] = struct{}{}
	}
	return res
}

// Precision truncates float64 to given precision
func Precision(value float64, precision int) float64 {
	d := math.Pow(10, float64(precision))
	return math.Trunc(value*d) / d
}

// NPrecision is like Precision but for nullables
func NPrecision(value float64, precision int) nulltype.NullFloat64 {
	return nulltype.NullFloat64Of(Precision(value, precision))
}

// TruncCoord truncates locations to 5 digits precision [link](https://en.wikipedia.org/wiki/Decimal_degrees)
func TruncCoord(value float64) float64 {
	return math.Trunc(value*100000) / 100000
}

// NTruncCoord is same as TruncCoord but for nullable values
func NTruncCoord(value float64) nulltype.NullFloat64 {
	return nulltype.NullFloat64Of(TruncCoord(value))
}

// MeasurementsFactory is helper used in tests
type MeasurementsFactory struct {
	Time   time.Time
	Script string
	Code   string
	Value  float64
}

// GenOne makes new measurement by adding offset hours to base time, offset as number to base value
// variadic args can be used to supply overrides in this order:
// - value
// - code
// - script
func (f *MeasurementsFactory) GenOne(offset int, args ...interface{}) Measurement {
	script, code, val := f.Script, f.Code, f.Value
	if val == 0 {
		val = 1
	}
	value := nulltype.NullFloat64Of(val + float64(offset))
	if len(args) > 0 {
		if oVal, ok := args[0].(float64); ok {
			if math.IsNaN(oVal) {
				value = nulltype.NullFloat64{}
			} else {
				value = nulltype.NullFloat64Of(oVal)
			}
		}
	}
	if len(args) > 1 {
		if oCode, ok := args[1].(string); ok {
			code = oCode
		}
	}
	if len(args) > 2 {
		if oScript, ok := args[2].(string); ok {
			script = oScript
		}
	}
	return Measurement{
		GaugeID: GaugeID{
			Script: script,
			Code:   code,
		},
		Timestamp: HTime{
			Time: f.Time.Add(time.Duration(offset) * time.Hour),
		},
		Level: value,
		Flow:  value,
	}
}

// GenOnePtr is like GenOne, but returns pointer to Measurement
func (f *MeasurementsFactory) GenOnePtr(offset int, args ...interface{}) *Measurement {
	m := f.GenOne(offset, args...)
	return &m
}

// GenMany will generate many measurements using GenOne in steps of 1
func (f *MeasurementsFactory) GenMany(size int) []Measurement {
	result := make([]Measurement, size)
	for i := 0; i < size; i++ {
		result[i] = f.GenOne(i)
	}
	return result
}

// GenManyPtr is like GenMany, but returns slice of pointers
func (f *MeasurementsFactory) GenManyPtr(size int) []*Measurement {
	result := make([]*Measurement, size)
	for i := 0; i < size; i++ {
		result[i] = f.GenOnePtr(i)
	}
	return result
}

// HarvestSlice is test helper that runs script's harvest and returns result as slice
func HarvestSlice(script Script, codes StringSet, since int64) (Measurements, error) {
	ctx := context.Background()
	in := make(chan *Measurement)
	errCh := make(chan error, 1)
	out := SinkToSlice(ctx, in)
	script.Harvest(ctx, in, errCh, codes, since)
	return <-out, <-errCh
}
