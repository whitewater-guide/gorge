package schedule

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/whitewater-guide/gorge/core"
)

func TestDeleteJob(t *testing.T) {
	scheduler := newMockScheduler(t)
	sched := scheduler.Cron.(*mockCron)

	entries := []cron.Entry{
		{
			ID:         0,
			Schedule:   nil,
			Next:       time.Time{},
			Prev:       time.Time{},
			WrappedJob: nil,
			Job: &harvestJob{
				database: scheduler.Database,
				cache:    scheduler.Cache,
				logger:   scheduler.Logger,
				cron:     "1 * * * *",
				script:   "one_by_one",
				jobID:    "f45829f1-357c-4b48-aa77-ee1edfa02e38",
				codes:    core.StringSet{"g001": {}},
			},
		},
		{
			ID:         1,
			Schedule:   nil,
			Next:       time.Time{},
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
			Next:       time.Time{},
			Prev:       time.Time{},
			WrappedJob: nil,
			Job: &harvestJob{
				database: scheduler.Database,
				cache:    scheduler.Cache,
				logger:   scheduler.Logger,
				cron:     "3 * * * *",
				script:   "one_by_one",
				jobID:    "6865c63f-0e02-467d-ad52-e53c2aca24e2",
				codes:    core.StringSet{"a001": {}},
			},
		},
	}
	sched.On("Entries").Return(entries)
	sched.On("Remove", mock.Anything).Return()
	scheduler.DeleteJob("f45829f1-357c-4b48-aa77-ee1edfa02e38") //nolint:errcheck
	sched.AssertNumberOfCalls(t, "Remove", 2)
	sched.AssertCalled(t, "Remove", cron.EntryID(0))
	sched.AssertCalled(t, "Remove", cron.EntryID(1))
	sched.AssertNotCalled(t, "Remove", cron.EntryID(2))
}

func TestDeleteJobError(t *testing.T) {
	scheduler := newMockScheduler(t)
	sched := scheduler.Cron.(*mockCron)
	sched.On("Entries").Return(make([]cron.Entry, 0))
	sched.On("Remove", mock.Anything).Return()
	err := scheduler.DeleteJob("foo")
	assert.Error(t, err, "must return error when job id is not found")
}
