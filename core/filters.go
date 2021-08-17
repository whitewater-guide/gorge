package core

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type filterStats struct {
	inCnt  int
	outCnt int
}

func (stats *filterStats) incoming() {
	stats.inCnt += 1
}

func (stats *filterStats) outgoing() {
	stats.outCnt += 1
}

func formatStats(m map[string]filterStats) string {
	result := ""
	for f, s := range m {
		result += fmt.Sprintf("[%s: %d/%d]", f, s.outCnt, s.inCnt)
	}
	return result
}

// MeasurementsFilter is used to skip unwanted measurements based on code, timestamp, etc...
type MeasurementsFilter interface {
	// Returns true if measurement matches filter's criteria
	filter(measurement Measurement) bool
	name() string
}

// FilterMeasurements filters channel of measurements using any number of filters
// It also supports context cancelation
func FilterMeasurements(ctx context.Context, in <-chan *Measurement, logger *logrus.Entry, filters ...MeasurementsFilter) <-chan *Measurement {
	out := make(chan *Measurement)

	stats := make(map[string]filterStats, len(filters))
	for _, f := range filters {
		stats[f.name()] = filterStats{}
	}

	go func() {
		defer close(out)
		defer func() {
			if logger != nil {
				logger.Debugf("filter stats %s", formatStats(stats))
			}
		}()
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
					fStats := stats[f.name()]
					fStats.incoming()
					stats[f.name()] = fStats
					if !f.filter(*m) {
						accept = false
						break
					} else {
						fStats.outgoing()
						stats[f.name()] = fStats
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

// LatestFilter returns measurements filter that accepts only measurements that are either
// - newer than latest measurement for same gauge in the map
// - or if no latest measurement is found, are no older than "After" date
type LatestFilter struct {
	Latest map[GaugeID]Measurement
	After  time.Time
}

func (f LatestFilter) filter(m Measurement) bool {
	l, ok := f.Latest[m.GaugeID]
	if ok {
		return m.Timestamp.After(l.Timestamp.Time)
	}
	return m.Timestamp.After(f.After)
}

func (f LatestFilter) name() string {
	return "latest"
}

// CodesFilter return measurements filter that accepts only measuerments with given gauge codes
type CodesFilter struct {
	Codes StringSet
}

func (f CodesFilter) filter(m Measurement) bool {
	return f.Codes.Contains(m.Code)
}

func (f CodesFilter) name() string {
	return "codes"
}
