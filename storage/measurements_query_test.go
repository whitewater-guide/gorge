package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func days(n int64) *time.Time {
	res := time.Now().Add(time.Duration(-24*n) * time.Hour)
	return &res
}

func TestParseTimeWindow(t *testing.T) {
	var tests = []struct {
		name  string
		start string
		end   string
		to    *time.Time
		from  *time.Time
		err   bool
	}{
		{
			name:  "no start + no end",
			start: "",
			end:   "",
			from:  nil,
			to:    nil,
			err:   false,
		},
		{
			name:  "recent start + no end",
			start: fmt.Sprint(days(1).Unix()),
			end:   "",
			from:  days(1),
			to:    nil,
			err:   false,
		},
		{
			name:  "old start + no end",
			start: fmt.Sprint(days(365).Unix()),
			end:   "",
			from:  days(30),
			to:    nil,
			err:   false,
		},
		{
			name:  "bad start + no end",
			start: "fffffuuu",
			end:   "",
			err:   true,
		},
		{
			name:  "no start + bad end",
			start: "",
			end:   "fffuuuuu",
			err:   true,
		},
		{
			name:  "no start + fixed end",
			start: "",
			end:   fmt.Sprint(days(1).Unix()),
			from:  days(31),
			to:    days(1),
			err:   false,
		},
		{
			name:  "recent start + fixed end",
			start: fmt.Sprint(days(2).Unix()),
			end:   fmt.Sprint(days(1).Unix()),
			from:  days(2),
			to:    days(1),
			err:   false,
		},
		{
			name:  "old start + fixed end",
			start: fmt.Sprint(days(82).Unix()),
			end:   fmt.Sprint(days(1).Unix()),
			from:  days(31),
			to:    days(1),
			err:   false,
		},
		{
			name:  "bad start + fixed end",
			start: "ffffuuuu",
			end:   fmt.Sprint(days(1).Unix()),
			err:   true,
		},
	}
	for _, tt := range tests {
		from, to, err := parseTimeWindow(tt.start, tt.end)
		if tt.err {
			assert.Error(t, err, "expected error in case of %s", tt.name)
		} else {
			if tt.from == nil {
				assert.Nil(t, from)
			} else {
				assert.InDelta(t, tt.from.Unix(), from.Unix(), 1000)
			}
			if tt.to == nil {
				assert.Nil(t, to)
			} else {
				assert.InDelta(t, tt.to.Unix(), to.Unix(), 1000)
			}
		}
	}
}
