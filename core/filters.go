package core

import (
	"context"
	"time"
)

// MeasurementsFilter is used to skip unwanted measurements based on code, timestamp, etc...
type MeasurementsFilter func(measurement Measurement) bool

// FilterMeasurements filters channel of measurements using any number of filters
// It also supports context cancelation
func FilterMeasurements(ctx context.Context, in <-chan *Measurement, filters ...MeasurementsFilter) <-chan *Measurement {
	out := make(chan *Measurement)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-in:
				if !ok {
					return
				}
				accept := true
				for _, f := range filters {
					if !f(*m) {
						accept = false
						break
					}
				}
				if accept {
					select {
					case <-ctx.Done():
						return
					case out <- m:
					}
				}
			}
		}
	}()
	return out
}

// NewLatestFilter returns measurements filter that accepts only measurements that are either
// - newer than latest measurement for same gauge in the map
// - or if no latest measurement is found, are no older than defaultHours
func NewLatestFilter(latest map[GaugeID]Measurement, defaultHours int) MeasurementsFilter {
	defaultStamp := time.Now().Add(time.Duration(-defaultHours) * time.Hour)
	return newLatestFilter(latest, defaultStamp)
}

func newLatestFilter(latest map[GaugeID]Measurement, defaultTime time.Time) MeasurementsFilter {
	return func(m Measurement) bool {
		l, ok := latest[m.GaugeID]
		if ok {
			return m.Timestamp.After(l.Timestamp.Time)
		}
		return m.Timestamp.After(defaultTime)
	}
}

// NewCodesFilter return measurements filter that accepts only measuerments with given gauge codes
func NewCodesFilter(codes StringSet) MeasurementsFilter {
	return func(m Measurement) bool {
		return codes.Contains(m.Code)
	}
}
