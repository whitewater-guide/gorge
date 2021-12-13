package schedule

import (
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestGetSince(t *testing.T) {
	assert := assert.New(t)
	job := &harvestJob{
		cron:   "10 * * * *",
		script: "one-by-one",
		jobID:  "f45829f1-357c-4b48-aa77-ee1edfa02e38",
		codes:  core.StringSet{"g001": {}},
	}
	cache := map[core.GaugeID]core.Measurement{
		{
			Script: "one-by-one",
			Code:   "g001",
		}: {
			GaugeID: core.GaugeID{
				Script: "one-by-one",
				Code:   "g001",
			},
			Timestamp: core.HTime{Time: time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC)},
			Level:     nulltype.NullFloat64Of(100),
			Flow:      nulltype.NullFloat64Of(100),
		},
	}
	assert.Equal(
		time.Date(2000, time.January, 1, 1, 1, 1, 1, time.UTC).Unix(),
		getSince(job, cache),
	)
}
