package core

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/mattn/go-nulltype"
)

// GenerateRandGauge generates random gauge for testing purposes
func GenerateRandGauge(script string, index int) Gauge {
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
			Longitude: rand.Float64()*360.0 - 180.0,
			Latitude:  rand.Float64()*180.0 - 90.0,
			Altitude:  rand.Float64() * 3000.0,
		},
		Timezone: "UTC",
	}
}

// GenerateRandMeasurement generates random measurement for testing purposes
// if value is not 0, it value will be returned for both level and flow
// otherwise level and flow will be random numbers in [min, max] range
func GenerateRandMeasurement(script string, code string, value float64, min float64, max float64) Measurement {
	level, flow := value, value
	if value == 0.0 {
		delta := max - min
		if delta == 0 {
			delta = 100
		} else if delta < 0 {
			delta = -delta
		}
		level = min + rand.Float64()*delta
		flow = min + rand.Float64()*delta
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
