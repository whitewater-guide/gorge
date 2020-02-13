package storage

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/whitewater-guide/gorge/core"
)

// MeasurementsQuery is intermediate data struct to convert HTTP request to database queries
type MeasurementsQuery struct {
	Script string
	Code   string
	From   *time.Time
	To     *time.Time
}

func parseTimeWindow(start, end string) (*time.Time, *time.Time, error) {
	if start == "" && end == "" {
		return nil, nil, nil
	}

	if end == "" { // period till now, trim to 30 days
		startI, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			return nil, nil, core.WrapErr(err, "invalid start timestamp").With("start", start)
		}
		startT := time.Unix(startI, 0)
		monthBefore := time.Now().Add(-30 * 24 * time.Hour)
		if startT.Before(monthBefore) {
			return &monthBefore, nil, nil
		}
		return &startT, nil, nil
	}

	endI, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		return nil, nil, core.WrapErr(err, "invalid start timestamp").With("start", start)
	}
	endT := time.Unix(endI, 0)
	monthBefore := endT.Add(-30 * 24 * time.Hour)

	if start == "" { // period with known end, trim to 30 days
		return &monthBefore, &endT, nil
	}
	startI, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return nil, nil, core.WrapErr(err, "invalid start timestamp").With("start", start)
	}
	startT := time.Unix(startI, 0)
	if startT.Before(monthBefore) {
		return &monthBefore, &endT, nil
	}
	return &startT, &endT, nil
}

// NewMeasurementsQuery builds db query from raw string arguments (passed via URL)
// If both TO and FROM are empty, a period of 30 days from current db time will be used
// If TO is empty string, current time from db will be used
// If given period is longer than 30 days, it will be trimmed to 30 days endig at TO timestamp
func NewMeasurementsQuery(script, code, fromS, toS string) (*MeasurementsQuery, error) {
	if script == "" {
		return nil, errors.New("script name is required")
	}
	from, to, err := parseTimeWindow(fromS, toS)
	if err != nil {
		return nil, err
	}
	return &MeasurementsQuery{
		Script: script,
		Code:   code,
		From:   from,
		To:     to,
	}, nil
}
