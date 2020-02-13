package sepa

import (
	"testing"
	"time"

	"github.com/mattn/go-nulltype"

	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestSepa_MeasurementFromRow(t *testing.T) {
	assert := assert.New(t)

	actual, err := measurementFromRow([]string{"18/01/2020 01:15:00", "1.708"})
	expected := &core.Measurement{
		Timestamp: core.HTime{Time: time.Date(2020, time.January, 18, 1, 15, 0, 0, time.UTC)},
		Level:     nulltype.NullFloat64Of(1.708),
	}
	if assert.NoError(err) {
		assert.Equal(expected, actual)
	}
}
