package core

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mattn/go-nulltype"
)

// HTime is just a wrapped around time which marshals/unmarshalt to/from RFC3339 string and can be stored in DB
type HTime struct {
	time.Time
}

// MarshalJSON implements json.Marshaler interface
func (t HTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.UTC().Format(time.RFC3339))
}

// UnmarshalJSON implements json.Unmarshaler interface
func (t *HTime) UnmarshalJSON(data []byte) error {
	v, err := time.Parse(time.RFC3339, string(data[1:len(data)-1]))
	t.Time = v
	return err
}

// Scan implements sql.Scanner interface
func (t *HTime) Scan(src interface{}) error {
	switch src := src.(type) {
	case time.Time:
		t.Time = src
	case string:
		ts, err := time.Parse(time.RFC3339, src)
		if err != nil {
			// sqlite format
			ts, err = time.Parse("2006-01-02 15:04:05+00:00", src)
			if err != nil {
				return fmt.Errorf("failed to scan HTime: invalid time string: %s", src)
			}
		}
		t.Time = ts
	default:
		return fmt.Errorf("failed to scan HTime: invalid source of type %T: %v", src, src)
	}
	return nil
}

// Value implements driver Valuer interface.
func (t HTime) Value() (driver.Value, error) {
	return t.Time, nil
}

// Measurement represents water level and/or flow value returned by gauge at the timestamp
type Measurement struct {
	GaugeID
	Timestamp HTime `json:"timestamp"`
	// Level is null when gauge doesn't provide it or is temporary broken
	Level nulltype.NullFloat64 `json:"level"`
	// Flow is null when gauge doesn't provide it or is temporary broken
	Flow nulltype.NullFloat64 `json:"flow"`
}

// Measurements is helper for sorting
type Measurements []*Measurement

func (m Measurements) Len() int {
	return len(m)
}

func (m Measurements) Less(i, j int) bool {
	// gauge id -> asc
	// timestamp -> desc
	if m[i].GaugeID == m[j].GaugeID {
		return m[i].Timestamp.After(m[j].Timestamp.Time)
	}
	return m[i].GaugeID.Less(&m[j].GaugeID)
}

func (m Measurements) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
