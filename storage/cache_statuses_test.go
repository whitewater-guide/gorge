package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/whitewater-guide/gorge/core"
	"gopkg.in/yaml.v3"
)

const (
	obo = "1d141432-5b1a-4a37-ab8e-78d912631fe5" // one-by-one job id

	aOk      = "622fd526-7fea-4870-be00-60c06576c5d5" // all-at-once job id (success)
	aErr     = "e2fb933f-cf9b-4956-a389-b95f7e2bba32" // all-at-once job id (success in past, then error)
	aErrOnly = "58cd6558-f356-450f-8ea6-755894059cd0" // all-at-once job id (error only)
)

var now = time.Date(2019, time.January, 1, 12, 0, 0, 0, time.UTC)
var nowStr = "2019-01-01T12:00:00Z"

var statusSeeds = `
jobs:
    # all-at-once job (success)

    622fd526-7fea-4870-be00-60c06576c5d5:
        success: 2017-01-01T12:00:00Z
        time: 2017-01-01T12:00:00Z
        count: 8
        error: ''

    # all-at-once job (success in past, then error)

    e2fb933f-cf9b-4956-a389-b95f7e2bba32:
        success: 2017-01-01T12:00:00Z
        time: 2017-01-02T12:00:00Z
        count: 0
        error: script error

    # all-at-once job (error only)

    58cd6558-f356-450f-8ea6-755894059cd0:
        time: 2017-01-02T12:00:00Z
        count: 0
        error: script error

# one-by-one job
1d141432-5b1a-4a37-ab8e-78d912631fe5:
    code_ok:
        success: 2017-01-01T12:00:00Z
        time: 2017-01-02T12:00:00Z
        count: 33
        error: ''
    code_err:
        success: 2017-01-01T12:00:00Z
        time: 2017-01-02T12:00:00Z
        count: 0
        error: gauge error
    code_err_only:
        time: 2017-01-01T12:00:00Z
        count: 0
        error: gauge error
`

type csTestSuite struct {
	suite.Suite
	mgr *EmbeddedCacheManager
}

func (s *csTestSuite) TearDownSuite() {
	s.mgr.Close()
}

// asserts that outer map contains inner map
func assertMapSubset(t *testing.T, outer map[string]string, inner map[string]string) {
	for k, v := range inner {
		o, ok := outer[k]
		if !ok {
			t.Errorf("inner map is not subset of outer map: key '%v' is missing", k)
			return
		}
		if v != o {
			t.Errorf("inner map is not subset of outer map: key '%v' has different value: '%v' in outer and '%v' in inner", k, o, v)
			return
		}
	}
}

func (s *csTestSuite) SetupTest() {
	conn := s.mgr.pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHALL")
	if err != nil {
		s.T().Fatalf("failed to flush redis: %v", err)
	}
	// seed data
	seeds := make(map[string]map[string]map[string]interface{})
	err = yaml.Unmarshal([]byte(statusSeeds), &seeds)
	if err != nil {
		s.T().Fatalf("yaml parse error: %v", err)
	}

	if err := conn.Send("MULTI"); err != nil {
		s.T().Fatalf("MULTI error: %v", err)
	}
	for key, hash := range seeds {
		args := []interface{}{fmt.Sprintf("%s:%s", NSStatus, key)}
		for prefix, status := range hash {
			for field, value := range status {
				val := fmt.Sprintf("%v", value)
				if t, ok := value.(time.Time); ok {
					val = t.Format(time.RFC3339)
				}
				args = append(args, fmt.Sprintf("%s:%s", prefix, field), val)
			}
		}
		if err := conn.Send("HMSET", args...); err != nil {
			s.T().Fatalf("HMSET error: %v", err)
		}
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		s.T().Fatalf("yaml save error: %v", err)
	}
}

func (s *csTestSuite) TestLoadJobStatuses() {
	t := s.T()
	expected := map[string]core.Status{
		aOk: {
			Success:     true,
			Timestamp:   core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			LastSuccess: &core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Count:       8,
		},
		aErr: {
			Success:     false,
			LastSuccess: &core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Timestamp:   core.HTime{Time: time.Date(2017, time.January, 2, 12, 0, 0, 0, time.UTC)},
			Count:       0,
			Error:       "script error",
		},
		aErrOnly: {
			Success:   false,
			Timestamp: core.HTime{Time: time.Date(2017, time.January, 2, 12, 0, 0, 0, time.UTC)},
			Count:     0,
			Error:     "script error",
		},
	}
	s.SetupTest()

	actual, err := s.mgr.LoadJobStatuses()
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
		// assert.Equal(t, expected[aOk], actual[aOk])
		// assert.Equal(t, expected[aErr], actual[aErr])
		// assert.Equal(t, expected[aErrOnly], actual[aErrOnly])
	}
}

func (s *csTestSuite) TestLoadGaugeStatuses() {
	t := s.T()
	expected := map[string]core.Status{
		"code_ok": {
			Success:     true,
			LastSuccess: &core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Timestamp:   core.HTime{Time: time.Date(2017, time.January, 2, 12, 0, 0, 0, time.UTC)},
			Count:       33,
		},
		"code_err": {
			Success:     false,
			LastSuccess: &core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Timestamp:   core.HTime{Time: time.Date(2017, time.January, 2, 12, 0, 0, 0, time.UTC)},
			Count:       0,
			Error:       "gauge error",
		},
		"code_err_only": {
			Success:   false,
			Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 12, 0, 0, 0, time.UTC)},
			Count:     0,
			Error:     "gauge error",
		},
	}
	s.SetupTest()

	actual, err := s.mgr.LoadGaugeStatuses(obo)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
		// assert.Equal(t, expected[aOk], actual[aOk])
		// assert.Equal(t, expected[aErr], actual[aErr])
		// assert.Equal(t, expected[aErrOnly], actual[aErrOnly])
	}
}

func (s *csTestSuite) TestLoadGaugeStatusesForInexistingJob() {
	t := s.T()
	expected := map[string]core.Status{}
	s.SetupTest()

	actual, err := s.mgr.LoadGaugeStatuses("does_not_exist")
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func (s *csTestSuite) TestSaveStatus() {
	// update existing gauge error only -> success
	// update existing gauge error only -> error
	t := s.T()
	tests := []struct {
		name     string
		jobID    string
		code     string
		err      error
		count    int
		expected map[string]string // expected raw redis response
	}{
		{
			name:  "save completely new job success",
			jobID: "ade7ffa1-1f4f-405f-9065-cefcd0b5f72c",
			count: 7,
			expected: map[string]string{
				"ade7ffa1-1f4f-405f-9065-cefcd0b5f72c:success": nowStr,
				"ade7ffa1-1f4f-405f-9065-cefcd0b5f72c:time":    nowStr,
				"ade7ffa1-1f4f-405f-9065-cefcd0b5f72c:count":   "7",
				"ade7ffa1-1f4f-405f-9065-cefcd0b5f72c:error":   "",
			},
		},
		{
			name:  "save completely new job error",
			jobID: "ade7ffa1-1f4f-405f-9065-cefcd0b5f72c",
			err:   fmt.Errorf("job failed"),
			expected: map[string]string{
				"ade7ffa1-1f4f-405f-9065-cefcd0b5f72c:time":  nowStr,
				"ade7ffa1-1f4f-405f-9065-cefcd0b5f72c:count": "0",
				"ade7ffa1-1f4f-405f-9065-cefcd0b5f72c:error": "job failed",
			},
		},
		{
			name:  "save completely new gauge success",
			jobID: "ade7ffa1-1f4f-405f-9065-cefcd0b5f72c",
			code:  "gauge01",
			count: 7,
			expected: map[string]string{
				"gauge01:success": nowStr,
				"gauge01:time":    nowStr,
				"gauge01:count":   "7",
				"gauge01:error":   "",
			},
		},
		{
			name:  "save completely new gauge error",
			jobID: "ade7ffa1-1f4f-405f-9065-cefcd0b5f72c",
			code:  "gauge01",
			err:   fmt.Errorf("job failed"),
			expected: map[string]string{
				"gauge01:time":  nowStr,
				"gauge01:count": "0",
				"gauge01:error": "job failed",
			},
		},
		{
			name:  "update existing job success -> success",
			jobID: aOk,
			count: 33,
			expected: map[string]string{
				"622fd526-7fea-4870-be00-60c06576c5d5:success": nowStr,
				"622fd526-7fea-4870-be00-60c06576c5d5:time":    nowStr,
				"622fd526-7fea-4870-be00-60c06576c5d5:count":   "33",
				"622fd526-7fea-4870-be00-60c06576c5d5:error":   "",
			},
		},
		{
			name:  "update existing job success -> error",
			jobID: aOk,
			err:   fmt.Errorf("job failed"),
			expected: map[string]string{
				"622fd526-7fea-4870-be00-60c06576c5d5:success": "2017-01-01T12:00:00Z",
				"622fd526-7fea-4870-be00-60c06576c5d5:time":    nowStr,
				"622fd526-7fea-4870-be00-60c06576c5d5:count":   "0",
				"622fd526-7fea-4870-be00-60c06576c5d5:error":   "job failed",
			},
		},
		{
			name:  "update existing job succes + error -> success",
			jobID: aErr,
			count: 33,
			expected: map[string]string{
				"e2fb933f-cf9b-4956-a389-b95f7e2bba32:success": nowStr,
				"e2fb933f-cf9b-4956-a389-b95f7e2bba32:time":    nowStr,
				"e2fb933f-cf9b-4956-a389-b95f7e2bba32:count":   "33",
				"e2fb933f-cf9b-4956-a389-b95f7e2bba32:error":   "",
			},
		},
		{
			name:  "update existing job succes + error -> error",
			jobID: aErr,
			err:   fmt.Errorf("job failed"),
			expected: map[string]string{
				"e2fb933f-cf9b-4956-a389-b95f7e2bba32:success": "2017-01-01T12:00:00Z",
				"e2fb933f-cf9b-4956-a389-b95f7e2bba32:time":    nowStr,
				"e2fb933f-cf9b-4956-a389-b95f7e2bba32:count":   "0",
				"e2fb933f-cf9b-4956-a389-b95f7e2bba32:error":   "job failed",
			},
		},
		{
			name:  "update existing job error only -> success",
			jobID: aErrOnly,
			count: 33,
			expected: map[string]string{
				"58cd6558-f356-450f-8ea6-755894059cd0:success": nowStr,
				"58cd6558-f356-450f-8ea6-755894059cd0:time":    nowStr,
				"58cd6558-f356-450f-8ea6-755894059cd0:count":   "33",
				"58cd6558-f356-450f-8ea6-755894059cd0:error":   "",
			},
		},
		{
			name:  "update existing job error only -> error",
			jobID: aErrOnly,
			err:   fmt.Errorf("job failed"),
			expected: map[string]string{
				"58cd6558-f356-450f-8ea6-755894059cd0:time":  nowStr,
				"58cd6558-f356-450f-8ea6-755894059cd0:count": "0",
				"58cd6558-f356-450f-8ea6-755894059cd0:error": "job failed",
			},
		},
		{
			name:  "update existing gauge success -> success",
			jobID: obo,
			code:  "code_ok",
			count: 44,
			expected: map[string]string{
				"code_ok:success": nowStr,
				"code_ok:time":    nowStr,
				"code_ok:count":   "44",
				"code_ok:error":   "",
			},
		},
		{
			name:  "update existing gauge success -> error",
			jobID: obo,
			code:  "code_ok",
			err:   fmt.Errorf("code_ok failed"),
			expected: map[string]string{
				"code_ok:success": "2017-01-01T12:00:00Z",
				"code_ok:time":    nowStr,
				"code_ok:count":   "0",
				"code_ok:error":   "code_ok failed",
			},
		},
		{
			name:  "update existing gauge success + error -> success",
			jobID: obo,
			code:  "code_err",
			count: 45,
			expected: map[string]string{
				"code_err:success": nowStr,
				"code_err:time":    nowStr,
				"code_err:count":   "45",
				"code_err:error":   "",
			},
		},
		{
			name:  "update existing gauge success + error -> error",
			jobID: obo,
			code:  "code_err",
			err:   fmt.Errorf("code_err failed"),
			expected: map[string]string{
				"code_err:success": "2017-01-01T12:00:00Z",
				"code_err:time":    nowStr,
				"code_err:count":   "0",
				"code_err:error":   "code_err failed",
			},
		},
		{
			name:  "update existing gauge error only -> success",
			jobID: obo,
			code:  "code_err_only",
			count: 22,
			expected: map[string]string{
				"code_err_only:success": nowStr,
				"code_err_only:time":    nowStr,
				"code_err_only:count":   "22",
				"code_err_only:error":   "",
			},
		},
		{
			name:  "update existing gauge error only -> error",
			jobID: obo,
			code:  "code_err_only",
			err:   fmt.Errorf("code_err_only failed"),
			expected: map[string]string{
				"code_err_only:time":  nowStr,
				"code_err_only:count": "0",
				"code_err_only:error": "code_err_only failed",
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

				var key string
				if tt.code == "" {
					key = fmt.Sprintf("%s:jobs", NSStatus)
				} else {
					key = fmt.Sprintf("%s:%s", NSStatus, tt.jobID)
				}

				actual, _ := redis.StringMap(conn.Do("HGETALL", key))
				assertMapSubset(t, actual, tt.expected)
			}
		})
	}
}

func TestCacheStatuses(t *testing.T) {
	mgr := &EmbeddedCacheManager{}
	mgr.Start() //nolint:errcheck
	tests := &csTestSuite{mgr: mgr}
	suite.Run(t, tests)
}
