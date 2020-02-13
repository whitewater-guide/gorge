package storage

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/whitewater-guide/gorge/core"
)

const (
	oOld     = "1d141432-5b1a-4a37-ab8e-78d912631fe5" // old one-by-one job id
	aOld     = "622fd526-7fea-4870-be00-60c06576c5d5" // old all-at-once job id
	oNew     = "56134f8b-7b5f-4976-84b5-e8b0aa8b043f" // new one-by-one job id
	brokenID = "467d8504-1380-4efe-bd1a-d579eae090ed" // broken job id
)

var now = time.Date(2019, time.January, 1, 12, 0, 0, 0, time.UTC)

type csTestSuite struct {
	suite.Suite
	mgr *EmbeddedCacheManager
}

func (s *csTestSuite) TearDownSuite() {
	s.mgr.Close()
}

func (s *csTestSuite) SetupTest() {
	conn := s.mgr.pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHALL")
	if err != nil {
		s.T().Fatalf("failed to flush redis: %v", err)
	}
	_, err = conn.Do(
		"HSET",
		NSStatus,
		oOld,
		`{"timestamp": "2017-01-01T12:00:00Z", "success": false, "count": 0, "error": "crash"}`,
		aOld,
		`{"timestamp": "2017-01-01T12:00:00Z", "success": true, "count": 10}`,
	)
	if err != nil {
		s.T().Fatalf("failed to save job statuses: %v", err)
	}
	_, err = conn.Do(
		"HSET",
		fmt.Sprintf("%s:%s", NSStatus, oOld),
		"o001",
		`{"timestamp": "2017-01-01T12:00:00Z", "success": false, "count": 0, "error": "boom"}`,
		"o000",
		`{"timestamp": "2017-01-01T12:00:00Z", "success": true, "count": 10}`,
	)
	if err != nil {
		s.T().Fatalf("failed to save one-by-one job gauge statuses: %v", err)
	}
	_, err = conn.Do(
		"HSET",
		fmt.Sprintf("%s:%s", NSStatus, brokenID),
		"b000",
		`foo{`,
	)
	if err != nil {
		s.T().Fatalf("failed to broken job gauge statuses: %v", err)
	}
}

func (s *csTestSuite) TestLoadJobStatuses() {
	t := s.T()
	expected := map[string]core.Status{
		oOld: {
			Success:   false,
			Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Error:     "crash",
		},
		aOld: {
			Success:   true,
			Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Count:     10,
		},
	}
	s.SetupTest()

	actual, err := s.mgr.LoadJobStatuses()
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func (s *csTestSuite) TestLoadGaugeStatuses() {
	t := s.T()
	tests := []struct {
		name     string
		jobID    string
		expected map[string]core.Status
		err      bool
	}{
		{
			name:  "success",
			jobID: oOld,
			expected: map[string]core.Status{
				"o000": {
					Success:   true,
					Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
					Count:     10,
				},
				"o001": {
					Success:   false,
					Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
					Error:     "boom",
				},
			},
			err: false,
		},
		{
			name:  "broken gauge",
			jobID: brokenID,
			err:   true,
		},
		{
			name:     "missing job",
			jobID:    "missing",
			expected: map[string]core.Status{},
			err:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetupTest()

			res, err := s.mgr.LoadGaugeStatuses(tt.jobID)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func (s *csTestSuite) TestSaveStatus() {
	t := s.T()
	tests := []struct {
		name      string
		jobID     string
		code      string
		err       error
		count     int
		expectedJ map[string]string // expected job statuses
		expectedG map[string]string // expected gauge statuses for this job
	}{
		{
			name:  "existing job success",
			jobID: aOld,
			count: 7,
			expectedJ: map[string]string{
				aOld: `{"timestamp": "2019-01-01T12:00:00Z", "success": true, "count": 7}`,
				oOld: `{"timestamp": "2017-01-01T12:00:00Z", "success": false, "count": 0, "error": "crash"}`,
			},
			expectedG: map[string]string{},
		},
		{
			name:  "existing gauge error",
			jobID: oOld,
			code:  "o000",
			count: 0,
			err:   errors.New("fail"),
			expectedJ: map[string]string{
				oOld: `{"timestamp": "2019-01-01T12:00:00Z", "success": false, "count": 0, "error": "fail"}`,
				aOld: `{"timestamp": "2017-01-01T12:00:00Z", "success": true, "count": 10}`,
			},
			expectedG: map[string]string{
				"o000": `{"timestamp": "2019-01-01T12:00:00Z", "success": false, "count": 0, "error": "fail"}`,
				"o001": `{"timestamp": "2017-01-01T12:00:00Z", "success": false, "count": 0, "error": "boom"}`,
			},
		},
		{
			name:  "new gauge error",
			jobID: oNew,
			code:  "n000",
			count: 0,
			err:   errors.New("broken"),
			expectedJ: map[string]string{
				oNew: `{"timestamp": "2019-01-01T12:00:00Z", "success": false, "count": 0, "error": "broken"}`,
				oOld: `{"timestamp": "2017-01-01T12:00:00Z", "success": false, "count": 0, "error": "crash"}`,
				aOld: `{"timestamp": "2017-01-01T12:00:00Z", "success": true, "count": 10}`,
			},
			expectedG: map[string]string{
				"n000": `{"timestamp": "2019-01-01T12:00:00Z", "success": false, "count": 0, "error": "broken"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetupTest()
			err := s.mgr.saveStatusWithTime(tt.jobID, tt.code, tt.err, tt.count, now)
			if assert.NoError(t, err) {
				conn := s.mgr.pool.Get()
				defer conn.Close()
				actJ, _ := redis.StringMap(conn.Do("HGETALL", NSStatus))
				actG, _ := redis.StringMap(conn.Do("HGETALL", fmt.Sprintf("%s:%s", NSStatus, tt.jobID)))
				assert.Equal(t, len(tt.expectedJ), len(actJ))
				for k, exS := range tt.expectedJ {
					assert.JSONEq(t, exS, actJ[k])
				}
				assert.Equal(t, len(tt.expectedG), len(actG))
				for k, exG := range tt.expectedG {
					assert.JSONEq(t, exG, actG[k])
				}
			}
		})
	}
}

func TestCacheStatuses(t *testing.T) {
	mgr, err := NewEmbeddedCacheManager()
	if err != nil {
		t.Fatalf("failed to init embedded redis cache manager: %v", err)
	}
	sqliteSuite := &csTestSuite{mgr: mgr}
	suite.Run(t, sqliteSuite)
}
