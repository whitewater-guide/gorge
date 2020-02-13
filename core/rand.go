package core

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mattn/go-nulltype"
)

// GenerateRandGauge generates random gauge for testing purposes
func GenerateRandGauge(script string, index int) Gauge {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return Gauge{
		GaugeID: GaugeID{
			Script: script,
			Code:   fmt.Sprintf("g%03d", index),
		},
		Name:      fmt.Sprintf("Test gauge #%d", index),
		URL:       fmt.Sprintf("http://whitewater.guide/gauges/%d", index),
		LevelUnit: "m",
		FlowUnit:  "m3/s",
		Location: &Location{
			Longitude: r.Float64()*360.0 - 180.0,
			Latitude:  r.Float64()*180.0 - 90.0,
			Altitude:  r.Float64() * 3000.0,
		},
	}
}

// GenerateRandMeasurement generates random measurement for testing purposes
// if value is not 0, it value will be returned for both level and flow
// otherwise level and flow will be random numbers in [min, max] range
func GenerateRandMeasurement(script string, code string, value float64, min float64, max float64) Measurement {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	level, flow := value, value
	if value == 0.0 {
		delta := max - min
		if delta == 0 {
			delta = 100
		} else if delta < 0 {
			delta = -delta
		}
		level = min + r.Float64()*delta
		flow = min + r.Float64()*delta
	}

	return Measurement{
		GaugeID: GaugeID{
			Script: script,
			Code:   code,
		},
		Timestamp: HTime{time.Now().UTC()},
		Level:     nulltype.NullFloat64Of(level),
		Flow:      nulltype.NullFloat64Of(flow),
	}
}
