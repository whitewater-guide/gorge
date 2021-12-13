package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/kinbiko/jsonassert"
	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/whitewater-guide/gorge/core"
)

type clTestSuite struct {
	suite.Suite
	mgr *EmbeddedCacheManager
}

func (s *clTestSuite) TearDownSuite() {
	s.mgr.Close()
}

func (s *clTestSuite) SetupTest() {
	conn := s.mgr.pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHALL")
	if err != nil {
		s.T().Fatalf("failed to flush redis: %v", err)
	}
	_, err = conn.Do(
		"HSET",
		fmt.Sprintf("%s:all_at_once", NSLatest),
		"a000",
		`{"script": "all_at_once", "code": "a000", "timestamp": "2018-01-01T12:00:00Z", "flow": 100, "level": 100}`,
		"a001",
		`{"script": "all_at_once", "code": "a001", "timestamp": "2018-01-01T12:00:00Z", "flow": 101, "level": 101}`,
		"a002",
		`{"script": "all_at_once", "code": "a002", "timestamp": "2018-01-01T12:00:00Z", "flow": 102, "level": 102}`,
	)
	if err != nil {
		s.T().Fatalf("failed to seed all_at_once data: %v", err)
	}
	_, err = conn.Do(
		"HSET",
		fmt.Sprintf("%s:one_by_one", NSLatest),
		"o000",
		`{"script": "one_by_one", "code": "o000", "timestamp": "2017-01-01T12:00:00Z", "flow": 0, "level": null}`,
		"o001",
		`{"script": "one_by_one", "code": "o001", "timestamp": "2017-01-01T12:00:00Z", "flow": 1, "level": null}`,
		"o002",
		`{"script": "one_by_one", "code": "o002", "timestamp": "2017-01-01T12:00:00Z", "flow": 2, "level": null}`,
	)
	if err != nil {
		s.T().Fatalf("failed to seed one_by_one data: %v", err)
	}
}

func (s *clTestSuite) TestSaveLastestMeasurements() {
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
	assert := assert.New(s.T())
	if assert.NoError(err) {
		conn := s.mgr.pool.Get()
		defer conn.Close()
		res, err := redis.StringMap(conn.Do("HGETALL", fmt.Sprintf("%s:all_at_once", NSLatest)))
		ja := jsonassert.New(s.T())
		if assert.NoError(err) {
			ja.Assertf(res["a000"], `{"script": "all_at_once", "code": "a000", "timestamp": "2019-01-01T00:00:00Z", "flow": 111, "level": 0}`)
			ja.Assertf(res["a001"], `{"script": "all_at_once", "code": "a001", "timestamp": "2018-01-01T12:00:00Z", "flow": 101, "level": 101}`)
			ja.Assertf(res["a002"], `{"script": "all_at_once", "code": "a002", "timestamp": "2018-01-01T12:00:00Z", "flow": 102, "level": 102}`)
			ja.Assertf(res["a003"], `{"script": "all_at_once", "code": "a003", "timestamp": "2019-01-01T00:00:00Z", "flow": 113, "level": 113}`)
		}
		res, err = redis.StringMap(conn.Do("HGETALL", fmt.Sprintf("%s:one_by_one", NSLatest)))
		if assert.NoError(err) {
			ja.Assertf(res["o000"], `{"script": "one_by_one", "code": "o000", "timestamp": "2019-01-01T00:00:00Z", "flow": 4, "level": 4}`)
			ja.Assertf(res["o001"], `{"script": "one_by_one", "code": "o001", "timestamp": "2017-01-01T12:00:00Z", "flow": 1, "level": null}`)
			ja.Assertf(res["o002"], `{"script": "one_by_one", "code": "o002", "timestamp": "2017-01-01T12:00:00Z", "flow": 2, "level": null}`)
		}
	}
}

func (s *clTestSuite) TestSaveLastestMeasurementsCanceled() {
	factory := core.MeasurementsFactory{Script: "broken", Code: "b000"}
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan *core.Measurement)
	errCh := s.mgr.SaveLatestMeasurements(ctx, in)
	in <- factory.GenOnePtr(1)
	cancel()
	err := <-errCh
	assert := assert.New(s.T())
	// assert.False(ok)
	assert.Equal(err, context.Canceled)
	conn := s.mgr.pool.Get()
	defer conn.Close()
	res, _ := redis.StringMap(conn.Do("HGETALL", fmt.Sprintf("%s:broken", NSLatest)))
	assert.Empty(res)
}

func (s *clTestSuite) TestGetLatestMeasurements() {
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
			err:      false,
		},
		{
			name:     "some gauges for one source",
			input:    map[string]core.StringSet{"all_at_once": {"a000": {}}},
			expected: []float64{100},
			err:      false,
		},
		{
			name:     "all gauges for many sources",
			input:    map[string]core.StringSet{"all_at_once": {}, "one_by_one": {}},
			expected: []float64{100, 101, 102, 0, 1, 2},
			err:      false,
		},
		{
			name:     "some gauges for many sources",
			input:    map[string]core.StringSet{"all_at_once": {"a000": {}}, "one_by_one": {"o000": {}}},
			expected: []float64{100, 0},
			err:      false,
		},
		{
			name:     "all gauges for one source and some for another",
			input:    map[string]core.StringSet{"all_at_once": {}, "one_by_one": {"o000": {}}},
			expected: []float64{100, 101, 102, 0},
			err:      false,
		},
		{
			name:     "unknown gauges",
			input:    map[string]core.StringSet{"all_at_once": {"a000": {}, "a005": {}}},
			expected: []float64{100},
			err:      false,
		},
		{
			name:     "unknown scripts",
			input:    map[string]core.StringSet{"all_at_once": {}, "foo": {}},
			expected: []float64{100, 101, 102},
			err:      false,
		},
		{
			name:     "empty list",
			input:    map[string]core.StringSet{},
			expected: []float64{},
			err:      false,
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

func TestCacheLatest(t *testing.T) {
	mgr := &EmbeddedCacheManager{}
	mgr.Start() //nolint:errcheck
	sqliteSuite := &clTestSuite{mgr: mgr}
	suite.Run(t, sqliteSuite)
}
