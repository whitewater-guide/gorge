package schedule

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
)

func TestListNext(t *testing.T) {
	scheduler := newMockScheduler(t)
	sched := scheduler.Cron.(*mockCron)

	entries := []cron.Entry{
		{
			ID:         0,
			Schedule:   nil,
			Next:       time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC),
			Prev:       time.Time{},
			WrappedJob: nil,
			Job: &harvestJob{
				database: scheduler.Database,
				cache:    scheduler.Cache,
				logger:   scheduler.Logger,
				cron:     "1 * * * *",
				script:   "all_at_once",
				jobID:    "3816a33f-5511-4795-84e0-d6371de2dc2b",
				codes:    core.StringSet{"g001": {}, "g002": {}},
			},
		},
		{
			ID:         1,
			Schedule:   nil,
			Next:       time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC),
			Prev:       time.Time{},
			WrappedJob: nil,
			Job: &harvestJob{
				database: scheduler.Database,
				cache:    scheduler.Cache,
				logger:   scheduler.Logger,
				cron:     "2 * * * *",
				script:   "one_by_one",
				jobID:    "f45829f1-357c-4b48-aa77-ee1edfa02e38",
				codes:    core.StringSet{"g002": {}},
			},
		},
		{
			ID:         2,
			Schedule:   nil,
			Next:       time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC),
			Prev:       time.Time{},
			WrappedJob: nil,
			Job: &harvestJob{
				database: scheduler.Database,
				cache:    scheduler.Cache,
				logger:   scheduler.Logger,
				cron:     "3 * * * *",
				script:   "one_by_one",
				jobID:    "f45829f1-357c-4b48-aa77-ee1edfa02e38",
				codes:    core.StringSet{"g001": {}},
			},
		},
	}
	sched.On("Entries").Return(entries)

	actualJobs := scheduler.ListNext("")
	expectedJobs := map[string]core.HTime{
		"3816a33f-5511-4795-84e0-d6371de2dc2b": {Time: time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC)},
		"f45829f1-357c-4b48-aa77-ee1edfa02e38": {Time: time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)},
	}
	assert.Equal(t, expectedJobs, actualJobs)

	actualGauges := scheduler.ListNext("f45829f1-357c-4b48-aa77-ee1edfa02e38")
	expectedGauges := map[string]core.HTime{
		"g001": {Time: time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)},
		"g002": {Time: time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC)},
	}
	assert.Equal(t, expectedGauges, actualGauges)
}
