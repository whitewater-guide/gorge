package core

import (
	"context"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
)

func TestNewLatestFilter(t *testing.T) {
	f := LatestFilter{
		Latest: map[GaugeID]Measurement{
			GaugeID{"all_at_once", "a000"}: {
				GaugeID:   GaugeID{"all_at_once", "a000"},
				Timestamp: unixHTime(10000),
				Level:     nulltype.NullFloat64Of(100),
				Flow:      nulltype.NullFloat64Of(100),
			},
		},
		After: time.Unix(5000, 0),
	}

	tests := []struct {
		name     string
		input    Measurement
		expected bool
	}{
		{
			name: "latest good",
			input: Measurement{
				GaugeID:   GaugeID{"all_at_once", "a000"},
				Timestamp: unixHTime(12000),
			},
			expected: true,
		},
		{
			name: "latest bad",
			input: Measurement{
				GaugeID:   GaugeID{"all_at_once", "a000"},
				Timestamp: unixHTime(9000),
			},
			expected: false,
		},
		{
			name: "latest edge",
			input: Measurement{
				GaugeID:   GaugeID{"all_at_once", "a000"},
				Timestamp: unixHTime(10000),
			},
			expected: false,
		},
		{
			name: "default good",
			input: Measurement{
				GaugeID:   GaugeID{"all_at_once", "a001"},
				Timestamp: unixHTime(12000),
			},
			expected: true,
		},
		{
			name: "default bad",
			input: Measurement{
				GaugeID:   GaugeID{"all_at_once", "a001"},
				Timestamp: unixHTime(3000),
			},
			expected: false,
		},
		{
			name: "default edge",
			input: Measurement{
				GaugeID:   GaugeID{"all_at_once", "a001"},
				Timestamp: unixHTime(5000),
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			assert.Equal(t, tt.expected, f.filter(tt.input))
		})
	}
}

func TestNewCodesFilter(t *testing.T) {
	f := CodesFilter{Codes: StringSet{"a000": {}, "a001": {}}}

	tests := []struct {
		name     string
		input    Measurement
		expected bool
	}{
		{
			name: "good",
			input: Measurement{
				GaugeID: GaugeID{"all_at_once", "a000"},
			},
			expected: true,
		},
		{
			name: "bad",
			input: Measurement{
				GaugeID: GaugeID{"all_at_once", "a002"},
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			assert.Equal(t, tt.expected, f.filter(tt.input))
		})
	}
}

func TestPartitionRangeFilter(t *testing.T) {
	f := PartitionRangeFilter{Now: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)}

	tests := []struct {
		name     string
		input    Measurement
		expected bool
	}{
		{
			name: "too old",
			input: Measurement{
				Timestamp: HTime{Time: time.Date(1998, time.November, 1, 0, 0, 0, 0, time.UTC)},
				GaugeID:   GaugeID{"all_at_once", "a000"},
			},
			expected: false,
		},
		{
			name: "from the future",
			input: Measurement{
				Timestamp: HTime{Time: time.Date(2000, time.August, 1, 0, 0, 0, 0, time.UTC)},
				GaugeID:   GaugeID{"all_at_once", "a001"},
			},
			expected: false,
		},
		{
			name: "old but good",
			input: Measurement{
				Timestamp: HTime{Time: time.Date(1999, time.December, 1, 0, 0, 0, 0, time.UTC)},
				GaugeID:   GaugeID{"all_at_once", "a002"},
			},
			expected: true,
		},
		{
			name: "from the future but good",
			input: Measurement{
				Timestamp: HTime{Time: time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC)},
				GaugeID:   GaugeID{"all_at_once", "a002"},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			assert.Equal(t, tt.expected, f.filter(tt.input))
		})
	}
}

func TestFilterMeasurements(t *testing.T) {
	input := []Measurement{
		{
			// pass
			GaugeID:   GaugeID{"all_at_once", "a000"},
			Timestamp: unixHTime(12000),
		},
		{
			// fail by latest
			GaugeID:   GaugeID{"all_at_once", "a000"},
			Timestamp: unixHTime(7000),
		},
		{
			// fail by default
			GaugeID:   GaugeID{"all_at_once", "a001"},
			Timestamp: unixHTime(3000),
		},
		{
			// fail by codes
			GaugeID:   GaugeID{"all_at_once", "a002"},
			Timestamp: unixHTime(12000),
		},
	}
	expected := []*Measurement{
		{
			// pass
			GaugeID:   GaugeID{"all_at_once", "a000"},
			Timestamp: unixHTime(12000),
		},
	}
	fCodes := CodesFilter{Codes: StringSet{"a000": {}, "a001": {}}}
	fLatest := LatestFilter{
		Latest: map[GaugeID]Measurement{
			GaugeID{"all_at_once", "a000"}: {
				GaugeID:   GaugeID{"all_at_once", "a000"},
				Timestamp: unixHTime(10000),
				Level:     nulltype.NullFloat64Of(100),
				Flow:      nulltype.NullFloat64Of(100),
			},
		},
		After: time.Unix(5000, 0),
	}

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		gen := GenFromSlice(ctx, input)
		out := FilterMeasurements(ctx, gen, nil, fCodes, fLatest)
		resCh := SinkToSlice(ctx, out)
		actual := <-resCh
		assert.Equal(t, expected, actual)
	})
	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		gen := GenFromSlice(ctx, input)
		out := FilterMeasurements(ctx, gen, nil, fCodes, fLatest)
		resCh := SinkToSlice(ctx, out)
		cancel()
		actual, ok := <-resCh
		assert.Nil(t, actual)
		assert.False(t, ok)
	})
}

func BenchmarkFilterMeasurements(b *testing.B) {
	input := []Measurement{
		{
			// pass
			GaugeID:   GaugeID{"all_at_once", "a000"},
			Timestamp: unixHTime(12000),
		},
		{
			// fail by latest
			GaugeID:   GaugeID{"all_at_once", "a000"},
			Timestamp: unixHTime(7000),
		},
		{
			// fail by default
			GaugeID:   GaugeID{"all_at_once", "a001"},
			Timestamp: unixHTime(3000),
		},
		{
			// fail by codes
			GaugeID:   GaugeID{"all_at_once", "a002"},
			Timestamp: unixHTime(12000),
		},
	}
	fCodes := CodesFilter{Codes: StringSet{"a000": {}, "a001": {}}}
	fLatest := LatestFilter{
		Latest: map[GaugeID]Measurement{
			GaugeID{"all_at_once", "a000"}: {
				GaugeID:   GaugeID{"all_at_once", "a000"},
				Timestamp: unixHTime(10000),
				Level:     nulltype.NullFloat64Of(100),
				Flow:      nulltype.NullFloat64Of(100),
			},
		},
		After: time.Unix(5000, 0),
	}
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		gen := GenFromSlice(ctx, input)
		out := FilterMeasurements(ctx, gen, nil, fCodes, fLatest)
		resCh := SinkToSlice(ctx, out)
		<-resCh
	}
}
