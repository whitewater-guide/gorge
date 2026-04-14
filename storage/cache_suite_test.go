package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/whitewater-guide/gorge/core"
)

// testableCacheManager extends CacheManager with test-only helpers for seeding
// state at specific timestamps and flushing all data between test cases.
type testableCacheManager interface {
	CacheManager
	saveStatusAt(jobID, code string, err error, count int, ts time.Time) error
	flushAll() error
}

// EmbeddedCacheManager test helpers

func (cache *EmbeddedCacheManager) saveStatusAt(jobID, code string, err error, count int, ts time.Time) error {
	return cache.RedisCacheManager.saveStatusWithTime(jobID, code, err, count, ts)
}

func (cache *EmbeddedCacheManager) flushAll() error {
	conn := cache.pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHALL")
	return err
}

// Seed timestamps used across status suites.
var (
	seedT1 = time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)
	seedT2 = time.Date(2017, time.January, 2, 12, 0, 0, 0, time.UTC)
)

// seedStatuses seeds a fixed set of job/gauge statuses into mgr using only
// saveStatusAt so any testableCacheManager implementation can be used.
func seedStatuses(t *testing.T, mgr testableCacheManager) {
	t.Helper()
	// aOk: job that always succeeded
	require.NoError(t, mgr.saveStatusAt(aOk, "", nil, 8, seedT1))
	// aErr: job with prior success, then an error
	require.NoError(t, mgr.saveStatusAt(aErr, "", nil, 1, seedT1))
	require.NoError(t, mgr.saveStatusAt(aErr, "", errors.New("script error"), 0, seedT2))
	// aErrOnly: job that only ever errored
	require.NoError(t, mgr.saveStatusAt(aErrOnly, "", errors.New("script error"), 0, seedT2))
	// one-by-one job gauges
	require.NoError(t, mgr.saveStatusAt(obo, "code_ok", nil, 33, seedT2))
	require.NoError(t, mgr.saveStatusAt(obo, "code_err", nil, 1, seedT1))
	require.NoError(t, mgr.saveStatusAt(obo, "code_err", errors.New("gauge error"), 0, seedT2))
	require.NoError(t, mgr.saveStatusAt(obo, "code_err_only", errors.New("gauge error"), 0, seedT1))
}

type cacheStatusSuite struct {
	suite.Suite
	mgr testableCacheManager
}

func (s *cacheStatusSuite) TestLoadJobStatuses() {
	t := s.T()
	expected := map[string]core.Status{
		aOk: {
			LastRun:     core.HTime{Time: seedT1},
			LastSuccess: &core.HTime{Time: seedT1},
			Count:       8,
		},
		aErr: {
			LastRun:     core.HTime{Time: seedT2},
			LastSuccess: &core.HTime{Time: seedT1},
			Count:       0,
			Error:       "script error",
		},
		aErrOnly: {
			LastRun: core.HTime{Time: seedT2},
			Count:   0,
			Error:   "script error",
		},
	}
	actual, err := s.mgr.LoadJobStatuses()
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func (s *cacheStatusSuite) TestLoadGaugeStatuses() {
	t := s.T()
	expected := map[string]core.Status{
		"code_ok": {
			LastRun:     core.HTime{Time: seedT2},
			LastSuccess: &core.HTime{Time: seedT2},
			Count:       33,
		},
		"code_err": {
			LastRun:     core.HTime{Time: seedT2},
			LastSuccess: &core.HTime{Time: seedT1},
			Count:       0,
			Error:       "gauge error",
		},
		"code_err_only": {
			LastRun: core.HTime{Time: seedT1},
			Count:   0,
			Error:   "gauge error",
		},
	}
	actual, err := s.mgr.LoadGaugeStatuses(obo)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func (s *cacheStatusSuite) TestLoadGaugeStatusesForInexistingJob() {
	t := s.T()
	actual, err := s.mgr.LoadGaugeStatuses("does_not_exist")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]core.Status{}, actual)
	}
}

func (s *cacheStatusSuite) TestSaveStatus() {
	t := s.T()
	nowHTime := core.HTime{Time: now}
	seedT1HTime := core.HTime{Time: seedT1}
	seedT2HTime := core.HTime{Time: seedT2}

	newJobID := "ade7ffa1-1f4f-405f-9065-cefcd0b5f72c"

	tests := []struct {
		name     string
		jobID    string
		code     string
		err      error
		count    int
		expected core.Status
	}{
		{
			name:  "save completely new job success",
			jobID: newJobID,
			count: 7,
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &nowHTime,
				Count:       7,
			},
		},
		{
			name:  "save completely new job error",
			jobID: newJobID,
			err:   errors.New("job failed"),
			expected: core.Status{
				LastRun: nowHTime,
				Count:   0,
				Error:   "job failed",
			},
		},
		{
			name:  "update existing job success -> success",
			jobID: aOk,
			count: 33,
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &nowHTime,
				Count:       33,
			},
		},
		{
			name:  "update existing job success -> error",
			jobID: aOk,
			err:   errors.New("job failed"),
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &seedT1HTime,
				Count:       0,
				Error:       "job failed",
			},
		},
		{
			name:  "update existing job success+error -> success",
			jobID: aErr,
			count: 33,
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &nowHTime,
				Count:       33,
			},
		},
		{
			name:  "update existing job success+error -> error",
			jobID: aErr,
			err:   errors.New("job failed"),
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &seedT1HTime,
				Count:       0,
				Error:       "job failed",
			},
		},
		{
			name:  "update existing job error only -> success",
			jobID: aErrOnly,
			count: 33,
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &nowHTime,
				Count:       33,
			},
		},
		{
			name:  "update existing job error only -> error",
			jobID: aErrOnly,
			err:   errors.New("job failed"),
			expected: core.Status{
				LastRun: nowHTime,
				Count:   0,
				Error:   "job failed",
			},
		},
		{
			name:  "save completely new gauge success",
			jobID: newJobID,
			code:  "gauge01",
			count: 7,
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &nowHTime,
				Count:       7,
			},
		},
		{
			name:  "save completely new gauge error",
			jobID: newJobID,
			code:  "gauge01",
			err:   errors.New("job failed"),
			expected: core.Status{
				LastRun: nowHTime,
				Count:   0,
				Error:   "job failed",
			},
		},
		{
			name:  "update existing gauge success -> success",
			jobID: obo,
			code:  "code_ok",
			count: 44,
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &nowHTime,
				Count:       44,
			},
		},
		{
			name:  "update existing gauge success -> error",
			jobID: obo,
			code:  "code_ok",
			err:   errors.New("code_ok failed"),
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &seedT2HTime,
				Count:       0,
				Error:       "code_ok failed",
			},
		},
		{
			name:  "update existing gauge success+error -> success",
			jobID: obo,
			code:  "code_err",
			count: 45,
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &nowHTime,
				Count:       45,
			},
		},
		{
			name:  "update existing gauge success+error -> error",
			jobID: obo,
			code:  "code_err",
			err:   errors.New("code_err failed"),
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &seedT1HTime,
				Count:       0,
				Error:       "code_err failed",
			},
		},
		{
			name:  "update existing gauge error only -> success",
			jobID: obo,
			code:  "code_err_only",
			count: 22,
			expected: core.Status{
				LastRun:     nowHTime,
				LastSuccess: &nowHTime,
				Count:       22,
			},
		},
		{
			name:  "update existing gauge error only -> error",
			jobID: obo,
			code:  "code_err_only",
			err:   errors.New("code_err_only failed"),
			expected: core.Status{
				LastRun: nowHTime,
				Count:   0,
				Error:   "code_err_only failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, s.mgr.flushAll())
			seedStatuses(t, s.mgr)
			require.NoError(t, s.mgr.saveStatusAt(tt.jobID, tt.code, tt.err, tt.count, now))

			if tt.code == "" {
				statuses, err := s.mgr.LoadJobStatuses()
				if assert.NoError(t, err) {
					assert.Equal(t, tt.expected, statuses[tt.jobID])
				}
			} else {
				statuses, err := s.mgr.LoadGaugeStatuses(tt.jobID)
				if assert.NoError(t, err) {
					assert.Equal(t, tt.expected, statuses[tt.code])
				}
			}
		})
	}
}

type cacheLatestSuite struct {
	suite.Suite
	mgr testableCacheManager
}

func (s *cacheLatestSuite) SetupTest() {
	require.NoError(s.T(), s.mgr.flushAll())
	measurements := []core.Measurement{
		{
			GaugeID:   core.GaugeID{Script: "all_at_once", Code: "a000"},
			Timestamp: core.HTime{Time: time.Date(2018, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(100),
			Level:     nulltype.NullFloat64Of(100),
		},
		{
			GaugeID:   core.GaugeID{Script: "all_at_once", Code: "a001"},
			Timestamp: core.HTime{Time: time.Date(2018, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(101),
			Level:     nulltype.NullFloat64Of(101),
		},
		{
			GaugeID:   core.GaugeID{Script: "all_at_once", Code: "a002"},
			Timestamp: core.HTime{Time: time.Date(2018, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(102),
			Level:     nulltype.NullFloat64Of(102),
		},
		{
			GaugeID:   core.GaugeID{Script: "one_by_one", Code: "o000"},
			Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(0),
			Level:     nulltype.NullFloat64{},
		},
		{
			GaugeID:   core.GaugeID{Script: "one_by_one", Code: "o001"},
			Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(1),
			Level:     nulltype.NullFloat64{},
		},
		{
			GaugeID:   core.GaugeID{Script: "one_by_one", Code: "o002"},
			Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(2),
			Level:     nulltype.NullFloat64{},
		},
	}
	ctx := context.Background()
	in := core.GenFromSlice(ctx, measurements)
	require.NoError(s.T(), <-s.mgr.SaveLatestMeasurements(ctx, in))
}

func (s *cacheLatestSuite) TestSaveLatestMeasurements() {
	t := s.T()
	data := []core.Measurement{
		{
			GaugeID:   core.GaugeID{Script: "all_at_once", Code: "a000"},
			Timestamp: core.HTime{Time: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(111),
			Level:     nulltype.NullFloat64Of(0),
		},
		{
			GaugeID:   core.GaugeID{Script: "all_at_once", Code: "a002"},
			Timestamp: core.HTime{Time: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64{},
			Level:     nulltype.NullFloat64{},
		},
		{
			GaugeID:   core.GaugeID{Script: "all_at_once", Code: "a003"},
			Timestamp: core.HTime{Time: time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(222),
			Level:     nulltype.NullFloat64Of(222),
		},
		{
			GaugeID:   core.GaugeID{Script: "all_at_once", Code: "a003"},
			Timestamp: core.HTime{Time: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(113),
			Level:     nulltype.NullFloat64Of(113),
		},
		{
			GaugeID:   core.GaugeID{Script: "one_by_one", Code: "o000"},
			Timestamp: core.HTime{Time: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)},
			Flow:      nulltype.NullFloat64Of(4),
			Level:     nulltype.NullFloat64Of(4),
		},
	}
	ctx := context.Background()
	in := core.GenFromSlice(ctx, data)
	err := <-s.mgr.SaveLatestMeasurements(ctx, in)
	if !assert.NoError(t, err) {
		return
	}

	res, err := s.mgr.LoadLatestMeasurements(map[string]core.StringSet{
		"all_at_once": {},
		"one_by_one":  {},
	})
	if !assert.NoError(t, err) {
		return
	}

	t2019 := core.HTime{Time: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)}
	t2018 := core.HTime{Time: time.Date(2018, time.January, 1, 12, 0, 0, 0, time.UTC)}
	t2017 := core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)}

	assert.Equal(t, core.Measurement{GaugeID: core.GaugeID{Script: "all_at_once", Code: "a000"}, Timestamp: t2019, Flow: nulltype.NullFloat64Of(111), Level: nulltype.NullFloat64Of(0)}, res[core.GaugeID{Script: "all_at_once", Code: "a000"}])
	assert.Equal(t, core.Measurement{GaugeID: core.GaugeID{Script: "all_at_once", Code: "a001"}, Timestamp: t2018, Flow: nulltype.NullFloat64Of(101), Level: nulltype.NullFloat64Of(101)}, res[core.GaugeID{Script: "all_at_once", Code: "a001"}])
	assert.Equal(t, core.Measurement{GaugeID: core.GaugeID{Script: "all_at_once", Code: "a002"}, Timestamp: t2018, Flow: nulltype.NullFloat64Of(102), Level: nulltype.NullFloat64Of(102)}, res[core.GaugeID{Script: "all_at_once", Code: "a002"}])
	assert.Equal(t, core.Measurement{GaugeID: core.GaugeID{Script: "all_at_once", Code: "a003"}, Timestamp: t2019, Flow: nulltype.NullFloat64Of(113), Level: nulltype.NullFloat64Of(113)}, res[core.GaugeID{Script: "all_at_once", Code: "a003"}])
	assert.Equal(t, core.Measurement{GaugeID: core.GaugeID{Script: "one_by_one", Code: "o000"}, Timestamp: t2019, Flow: nulltype.NullFloat64Of(4), Level: nulltype.NullFloat64Of(4)}, res[core.GaugeID{Script: "one_by_one", Code: "o000"}])
	assert.Equal(t, core.Measurement{GaugeID: core.GaugeID{Script: "one_by_one", Code: "o001"}, Timestamp: t2017, Flow: nulltype.NullFloat64Of(1), Level: nulltype.NullFloat64{}}, res[core.GaugeID{Script: "one_by_one", Code: "o001"}])
	assert.Equal(t, core.Measurement{GaugeID: core.GaugeID{Script: "one_by_one", Code: "o002"}, Timestamp: t2017, Flow: nulltype.NullFloat64Of(2), Level: nulltype.NullFloat64{}}, res[core.GaugeID{Script: "one_by_one", Code: "o002"}])
}

func (s *cacheLatestSuite) TestSaveLatestMeasurementsCanceled() {
	t := s.T()
	factory := core.MeasurementsFactory{Script: "broken", Code: "b000"}
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan *core.Measurement)
	errCh := s.mgr.SaveLatestMeasurements(ctx, in)
	in <- factory.GenOnePtr(1)
	cancel()
	err := <-errCh
	assert.Equal(t, context.Canceled, err)

	res, loadErr := s.mgr.LoadLatestMeasurements(map[string]core.StringSet{"broken": {}})
	assert.NoError(t, loadErr)
	assert.Empty(t, res)
}

func (s *cacheLatestSuite) TestGetLatestMeasurements() {
	t := s.T()
	tests := []struct {
		name     string
		input    map[string]core.StringSet
		expected []float64
		err      bool
	}{
		{
			name:     "all gauges for one source",
			input:    map[string]core.StringSet{"all_at_once": {}},
			expected: []float64{100, 101, 102},
		},
		{
			name:     "some gauges for one source",
			input:    map[string]core.StringSet{"all_at_once": {"a000": {}}},
			expected: []float64{100},
		},
		{
			name:     "all gauges for many sources",
			input:    map[string]core.StringSet{"all_at_once": {}, "one_by_one": {}},
			expected: []float64{100, 101, 102, 0, 1, 2},
		},
		{
			name:     "some gauges for many sources",
			input:    map[string]core.StringSet{"all_at_once": {"a000": {}}, "one_by_one": {"o000": {}}},
			expected: []float64{100, 0},
		},
		{
			name:     "all gauges for one source and some for another",
			input:    map[string]core.StringSet{"all_at_once": {}, "one_by_one": {"o000": {}}},
			expected: []float64{100, 101, 102, 0},
		},
		{
			name:     "unknown gauges",
			input:    map[string]core.StringSet{"all_at_once": {"a000": {}, "a005": {}}},
			expected: []float64{100},
		},
		{
			name:     "unknown scripts",
			input:    map[string]core.StringSet{"all_at_once": {}, "foo": {}},
			expected: []float64{100, 101, 102},
		},
		{
			name:     "empty list",
			input:    map[string]core.StringSet{},
			expected: []float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetupTest()
			res, err := s.mgr.LoadLatestMeasurements(tt.input)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				flows := make([]float64, 0)
				for _, m := range res {
					flows = append(flows, m.Flow.Float64Value())
				}
				assert.ElementsMatch(t, tt.expected, flows)
			}
		})
	}
}
