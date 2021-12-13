package schedule

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/scripts/testscripts"
)

func setupScheduler(t *testing.T) (*mockScheduler, *mockCron) {
	scheduler := newMockScheduler(t)
	cron := scheduler.Cron.(*mockCron)
	cron.On("AddJob", mock.Anything, mock.Anything).Return(0, nil)
	return scheduler, cron
}

func TestAddJobBadScript(t *testing.T) {
	scheduler, _ := setupScheduler(t)
	defer scheduler.Stop()

	err := scheduler.AddJob(core.JobDescription{
		ID:     "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
		Script: "fooo",
		Gauges: map[string]json.RawMessage{},
	})
	assert.Error(t, err)
}

func TestAddJobWithoutGauges(t *testing.T) {
	scheduler, _ := setupScheduler(t)
	defer scheduler.Stop()

	t.Run("all at once", func(t *testing.T) {
		oneByOneErr := scheduler.AddJob(core.JobDescription{
			ID:     "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
			Script: "one_by_one",
			Gauges: map[string]json.RawMessage{},
		})
		assert.Error(t, oneByOneErr, "should fail for one-by-one without gauges")
	})

	t.Run("one by one", func(t *testing.T) {
		allAtOnceErr := scheduler.AddJob(core.JobDescription{
			ID:     "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
			Script: "all_at_once",
			Gauges: map[string]json.RawMessage{},
		})
		assert.Error(t, allAtOnceErr, "should fail for all-at-once without gauges")
	})
}

func TestAddJobWithBadCron(t *testing.T) {
	scheduler, _ := setupScheduler(t)
	defer scheduler.Stop()

	err := scheduler.AddJob(core.JobDescription{
		ID:     "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
		Script: "all_at_once",
		Gauges: map[string]json.RawMessage{"g001": []byte("{}")},
		Cron:   "foo",
	})
	assert.Error(t, err, "should fail for bad cron")
}

func TestAddJobWithBadOptions(t *testing.T) {
	scheduler, _ := setupScheduler(t)
	defer scheduler.Stop()

	err := scheduler.AddJob(core.JobDescription{
		ID:      "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
		Script:  "all_at_once",
		Gauges:  map[string]json.RawMessage{"g001": []byte("")},
		Cron:    "* * * * *",
		Options: json.RawMessage(`{"foo": "bar"}`),
	})
	assert.Error(t, err)
}

func TestAddJobAllAtOnce(t *testing.T) {
	scheduler, cron := setupScheduler(t)
	defer scheduler.Stop()

	err := scheduler.AddJob(core.JobDescription{
		ID:      "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
		Script:  "all_at_once",
		Gauges:  map[string]json.RawMessage{"g001": []byte("{}")},
		Cron:    "* * * * *",
		Options: json.RawMessage(`{"gauges": 3}`),
	})

	if assert.NoError(t, err) {
		cron.AssertCalled(t, "AddJob", "* * * * *", mock.Anything)
	}
}

func TestAddJobOneByOne(t *testing.T) {
	scheduler, cron := setupScheduler(t)
	defer scheduler.Stop()

	err := scheduler.AddJob(core.JobDescription{
		ID:     "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
		Script: "one_by_one",
		Gauges: map[string]json.RawMessage{
			"g001": []byte("{}"),
			"g002": []byte(`{"min": 100.0, "max": 300.0}`),
			"g003": []byte("{}"),
			"g004": []byte("{}"),
			"g005": []byte("{}"),
			"g006": []byte("{}"),
			"g007": []byte("{}"),
		},
		Cron:    "",
		Options: json.RawMessage(`{"min": 200.0}`),
	})

	if assert.NoError(t, err) {
		cron.AssertNumberOfCalls(t, "AddJob", 7)

		assert.Equal(t, "0 * * * *", cron.Calls[0].Arguments[0])
		assert.Equal(t, &harvestJob{
			database: scheduler.Database,
			cache:    scheduler.Cache,
			logger:   scheduler.Logger,
			registry: scheduler.Registry,
			cron:     "0 * * * *",
			jobID:    "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
			script:   "one_by_one",
			codes:    core.StringSet{"g001": {}},
			options:  &testscripts.OneByOneOptions{Gauges: 10, Min: 200.0, Max: 20},
		}, cron.Calls[0].Arguments[1])

		assert.Equal(t, "9 * * * *", cron.Calls[1].Arguments[0])
		assert.Equal(t, &harvestJob{
			database: scheduler.Database,
			cache:    scheduler.Cache,
			logger:   scheduler.Logger,
			registry: scheduler.Registry,
			cron:     "9 * * * *",
			script:   "one_by_one",
			jobID:    "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
			codes:    core.StringSet{"g002": {}},
			options:  &testscripts.OneByOneOptions{Gauges: 10, Min: 100.0, Max: 300},
		}, cron.Calls[1].Arguments[1])
		assert.Equal(t, "51 * * * *", cron.Calls[6].Arguments[0])
	}
}

func TestAddJobManyGauges(t *testing.T) {
	// this tests schedule generation
	scheduler, cron := setupScheduler(t)
	defer scheduler.Stop()

	gauges := make(map[string]json.RawMessage)
	for i := 0; i < 71; i++ {
		gauges[fmt.Sprintf("g%03d", i)] = []byte("{}")
	}

	err := scheduler.AddJob(core.JobDescription{
		ID:     "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
		Script: "one_by_one",
		Gauges: gauges,
		Cron:   "",
	})

	if assert.NoError(t, err) {
		cron.AssertNumberOfCalls(t, "AddJob", 71)
		for _, c := range cron.Calls {
			assert.NotContains(t, c.Arguments[0], "60")
		}
	}
}

func TestAddJobOneByOneTransaction(t *testing.T) {
	counter := counterCron{}
	scheduler := newMockScheduler(t)
	scheduler.Cron = &counter
	defer scheduler.Stop()

	err := scheduler.AddJob(core.JobDescription{
		ID:     "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
		Script: "one_by_one",
		Gauges: map[string]json.RawMessage{
			"g001": []byte("{}"),
			"g002": []byte(`{"sss`), // invalid json
			"g003": []byte("{}"),
		},
		Cron: "",
	})

	assert.Error(t, err)
	assert.Len(t, counter.entries, 0)
}

func TestAddJobBatched(t *testing.T) {
	scheduler, cron := setupScheduler(t)
	defer scheduler.Stop()

	err := scheduler.AddJob(core.JobDescription{
		ID:     "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
		Script: "batched",
		Gauges: map[string]json.RawMessage{
			"g001": []byte("{}"),
			"g002": []byte(`{"min": 100.0, "max": 300.0}`),
			"g003": []byte("{}"),
			"g004": []byte("{}"),
			"g005": []byte("{}"),
			"g006": []byte("{}"),
			"g007": []byte("{}"),
		},
		Cron:    "",
		Options: json.RawMessage(`{"min": 200.0, "batchSize": 3}`),
	})

	if assert.NoError(t, err) {
		cron.AssertNumberOfCalls(t, "AddJob", 3)

		assert.Equal(t, "0 * * * *", cron.Calls[0].Arguments[0])
		assert.Equal(t, &harvestJob{
			database: scheduler.Database,
			cache:    scheduler.Cache,
			logger:   scheduler.Logger,
			registry: scheduler.Registry,
			cron:     "0 * * * *",
			jobID:    "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
			script:   "batched",
			codes:    core.StringSet{"g001": {}, "g002": {}, "g003": {}},
			options:  &testscripts.BatchedOptions{Gauges: 10, BatchSize: 3, Min: 200.0, Max: 20},
		}, cron.Calls[0].Arguments[1])

		assert.Equal(t, "20 * * * *", cron.Calls[1].Arguments[0])
		assert.Equal(t, &harvestJob{
			database: scheduler.Database,
			cache:    scheduler.Cache,
			logger:   scheduler.Logger,
			registry: scheduler.Registry,
			cron:     "20 * * * *",
			script:   "batched",
			jobID:    "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
			codes:    core.StringSet{"g004": {}, "g005": {}, "g006": {}},
			options:  &testscripts.BatchedOptions{Gauges: 10, BatchSize: 3, Min: 200.0, Max: 20},
		}, cron.Calls[1].Arguments[1])

		assert.Equal(t, "40 * * * *", cron.Calls[2].Arguments[0])
		assert.Equal(t, &harvestJob{
			database: scheduler.Database,
			cache:    scheduler.Cache,
			logger:   scheduler.Logger,
			registry: scheduler.Registry,
			cron:     "40 * * * *",
			script:   "batched",
			jobID:    "7bf5a9c4-d406-46dd-b596-1cdfd343e121",
			codes:    core.StringSet{"g007": {}},
			options:  &testscripts.BatchedOptions{Gauges: 10, BatchSize: 3, Min: 200.0, Max: 20},
		}, cron.Calls[2].Arguments[1])

	}
}
