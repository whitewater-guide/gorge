package canada

import (
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestGetPairedGauge(t *testing.T) {
	assert.Equal(t, "10KAX01", getPairedGauge("10KA001"))
	assert.Equal(t, "10KA001", getPairedGauge("10KAX01"))
	assert.Equal(t, "11AB108", getPairedGauge("11AB108"))
}

func TestCanada_MeasurementFromRow(t *testing.T) {
	// 02YD002,2020-01-17T00:00:00-03:30,0.368,,,1,2.86,,,1
	s := scriptCanada{
		name: "canada",
	}
	actual, err := s.measurementFromRow([]string{
		"02YD002", "2020-01-17T00:00:00-03:30", "0.368", "", "", "1", "2.86", "", "", "1",
	})
	expected := &core.Measurement{
		GaugeID: core.GaugeID{
			Script: "canada",
			Code:   "02YD002",
		},
		Timestamp: core.HTime{
			Time: time.Date(2020, time.January, 17, 3, 30, 0, 0, time.UTC),
		},
		Flow:  nulltype.NullFloat64Of(2.86),
		Level: nulltype.NullFloat64Of(0.368),
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
