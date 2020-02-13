package ecuador

import (
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestEcuador_ParseMeasurement(t *testing.T) {
	assert := assert.New(t)
	raw := []interface{}{
		"20190210170000", 0.36, 50.0, 0.33, 50.0, 0.35, 50.0, 0.35, 50.0, 0.0, 50.0, 13.77, 50.0,
	}
	s := scriptEcuador{name: "ecuador"}
	m, err := s.parseMeasurement(raw, "H0064", 0, 7)
	if assert.NoError(err) {
		assert.Equal(&core.Measurement{
			GaugeID: core.GaugeID{
				Script: "ecuador",
				Code:   "H0064",
			},
			Timestamp: core.HTime{Time: time.Date(2019, time.February, 10, 17, 0, 0, 0, time.UTC)},
			Level:     nulltype.NullFloat64Of(0.35),
		}, m)
	}
}
