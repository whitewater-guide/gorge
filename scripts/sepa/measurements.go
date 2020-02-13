package sepa

import (
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/whitewater-guide/gorge/core"
)

func measurementFromRow(row []string) (*core.Measurement, error) {
	timestamp, err := time.Parse("02/01/2006 15:04:05", row[0])
	if err != nil {
		return nil, err
	}

	level := nulltype.NullFloat64{}
	err = level.UnmarshalJSON([]byte(row[1]))

	if err != nil {
		return nil, err
	}
	return &core.Measurement{
		Timestamp: core.HTime{Time: timestamp},
		Level:     level,
	}, nil
}
