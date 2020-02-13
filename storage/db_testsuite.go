package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/whitewater-guide/gorge/core"
)

func date(year int, month time.Month, day int) *time.Time {
	res := time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
	return &res
}

func seed(db *sqlx.DB) {
	almostNow := time.Now().Add(-1 * time.Hour).UTC()
	daysFromNow := time.Now().Add(-28 * time.Hour).UTC()
	measurements := []core.Measurement{
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a001",
			},
			Timestamp: core.HTime{Time: daysFromNow},
			Flow:      nulltype.NullFloat64Of(500),
			Level:     nulltype.NullFloat64Of(500),
		},
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a001",
			},
			Timestamp: core.HTime{Time: almostNow},
			Flow:      nulltype.NullFloat64Of(400),
			Level:     nulltype.NullFloat64Of(400),
		},
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a001",
			},
			Timestamp: core.HTime{Time: *date(2018, time.January, 1)},
			Flow:      nulltype.NullFloat64Of(100),
			Level:     nulltype.NullFloat64Of(100),
		},
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a002",
			},
			Timestamp: core.HTime{Time: *date(2018, time.January, 2)},
			Flow:      nulltype.NullFloat64Of(333),
			Level:     nulltype.NullFloat64Of(333),
		},
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a001",
			},
			Timestamp: core.HTime{Time: *date(2018, time.January, 3)},
			Flow:      nulltype.NullFloat64Of(101),
			Level:     nulltype.NullFloat64Of(101),
		},
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a001",
			},
			Timestamp: core.HTime{Time: *date(2018, time.January, 4)},
			Flow:      nulltype.NullFloat64{},
			Level:     nulltype.NullFloat64Of(103),
		},
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a001",
			},
			Timestamp: core.HTime{Time: *date(2018, time.January, 5)},
			Flow:      nulltype.NullFloat64Of(300),
			Level:     nulltype.NullFloat64Of(300),
		},
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a001",
			},
			Timestamp: core.HTime{Time: *date(2018, time.January, 6)},
			Flow:      nulltype.NullFloat64Of(102),
			Level:     nulltype.NullFloat64Of(102),
		},
		core.Measurement{
			GaugeID: core.GaugeID{
				Script: "all_at_once",
				Code:   "a001",
			},
			Timestamp: core.HTime{Time: *date(2018, time.January, 7)},
			Flow:      nulltype.NullFloat64Of(200),
			Level:     nulltype.NullFloat64Of(200),
		},
	}
	jobs := [][]string{
		{"01e99188-2189-11ea-978f-2e728ce88125", `{"id": "01e99188-2189-11ea-978f-2e728ce88125", "script": "all_at_once", "gauges": {"a001": {}, "a002": {}}, "cron": "1 * * * *", "options": {"foo": "bar"}}`},
		{"0d67638c-2189-11ea-978f-2e728ce88125", `{"id": "0d67638c-2189-11ea-978f-2e728ce88125", "script": "one_by_one", "gauges": {"o001": {"foo": "bar"}, "o002": {}}, "options": {"foo": 42.0}}`},
	}
	qs := "INSERT INTO measurements (timestamp, script, code, flow, level) VALUES (:timestamp, :script, :code, :flow, :level)"
	_, err := db.NamedExec(qs, measurements)
	if err != nil {
		log.Fatalf("failed to seed measurements: %v", err)
	}
	qs = "INSERT INTO jobs (id, description) VALUES"
	for _, d := range jobs {
		qs = qs + fmt.Sprintf(" ('%s', '%s'),", d[0], d[1])
	}
	qs = qs[:len(qs)-1]
	_, err = db.Exec(qs)
	if err != nil {
		log.Fatalf("failed to seed jobs: %v", err)
	}
}

func countJobs(db *sqlx.DB) int {
	row := db.QueryRow("SELECT count(*) FROM jobs")
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		log.Fatalf("failed to count raw jobs: %v", err)
	}
	return cnt
}

func cleanup(db *sqlx.DB) {
	_, err := db.Exec("DELETE FROM jobs")
	if err != nil {
		log.Fatalf("failed to clean up jobs")
	}
	_, err = db.Exec("DELETE FROM measurements")
	if err != nil {
		log.Fatalf("failed to clean up measurements")
	}
}

type DbTestSuite struct {
	suite.Suite
	mgr *DbManager
}

func (s *DbTestSuite) TearDownSuite() {
	s.mgr.Close()
}

func (s *DbTestSuite) SetupTest() {
	cleanup(s.mgr.db)
	seed(s.mgr.db)
}

func (s *DbTestSuite) TestGetMeasurements() {
	t := s.T()
	tests := []struct {
		name     string
		query    MeasurementsQuery
		expected []float64
	}{
		{
			name: "default",
			query: MeasurementsQuery{
				Script: "all_at_once",
				Code:   "a001",
			},
			expected: []float64{400, 500},
		},
		{
			name: "all 2018",
			query: MeasurementsQuery{
				Script: "all_at_once",
				Code:   "a001",
				From:   date(2018, time.January, 1),
				To:     date(2019, time.January, 1),
			},
			expected: []float64{200, 102, 300, math.NaN(), 101, 100},
		},
		{
			name: "open end",
			query: MeasurementsQuery{
				Script: "all_at_once",
				Code:   "a001",
				From:   date(2018, time.February, 1),
				To:     nil,
			},
			expected: []float64{400, 500},
		},
		{
			name: "edge case",
			query: MeasurementsQuery{
				Script: "all_at_once",
				Code:   "a001",
				From:   date(2018, time.January, 7),
				To:     nil,
			},
			expected: []float64{400, 500, 200},
		},
		{
			name: "all gauges",
			query: MeasurementsQuery{
				Script: "all_at_once",
				From:   date(2018, time.January, 1),
				To:     date(2018, time.January, 3),
			},
			expected: []float64{101, 333, 100},
		},
		{
			name: "empty result",
			query: MeasurementsQuery{
				Script: "all_at_once",
				Code:   "a001",
				From:   date(2017, time.January, 3),
				To:     date(2017, time.February, 3),
			},
			expected: []float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetupTest()
			measurements, err := s.mgr.GetMeasurements(tt.query)
			if assert.NoError(t, err) {
				if assert.Equal(t, len(tt.expected), len(measurements), "got %v", measurements) {
					for i, m := range measurements {
						expected := nulltype.NullFloat64Of(tt.expected[i])
						if math.IsNaN(tt.expected[i]) {
							expected.Reset()
						}
						assert.Equal(t, expected, m.Flow, "value at position %d was %.0f, but expected %.0f", i, m.Flow.Float64Value(), expected.Float64Value())
					}
				}
			}
		})
	}
}

func (s *DbTestSuite) TestSaveMeasurements() {
	t := s.T()
	null := nulltype.NullFloat64Of(0)
	null.Reset()
	tests := []struct {
		name    string
		input   []core.Measurement
		result  int
		content []nulltype.NullFloat64
	}{
		{
			name: "unique values",
			input: []core.Measurement{
				{
					GaugeID: core.GaugeID{
						Script: "all_at_once",
						Code:   "a002",
					},
					Timestamp: core.HTime{Time: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)},
					Level:     nulltype.NullFloat64Of(999),
					Flow:      nulltype.NullFloat64Of(999),
				},
				{
					GaugeID: core.GaugeID{
						Script: "all_at_once",
						Code:   "a002",
					},
					Timestamp: core.HTime{Time: time.Date(2019, time.February, 1, 0, 0, 0, 0, time.UTC)},
					Level:     nulltype.NullFloat64Of(777),
					Flow:      nulltype.NullFloat64Of(777),
				},
			},
			result:  2,
			content: []nulltype.NullFloat64{nulltype.NullFloat64Of(777), nulltype.NullFloat64Of(999), nulltype.NullFloat64Of(333)},
		},
		{
			name: "dupe value",
			input: []core.Measurement{
				{
					GaugeID: core.GaugeID{
						Script: "all_at_once",
						Code:   "a002",
					},
					Timestamp: core.HTime{Time: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)},
					Level:     nulltype.NullFloat64Of(999),
					Flow:      nulltype.NullFloat64Of(999),
				},
				{
					GaugeID: core.GaugeID{
						Script: "all_at_once",
						Code:   "a002",
					},
					Timestamp: core.HTime{Time: time.Date(2018, time.January, 2, 12, 0, 0, 0, time.UTC)},
					Level:     nulltype.NullFloat64Of(777),
					Flow:      nulltype.NullFloat64Of(777),
				},
			},
			result:  1,
			content: []nulltype.NullFloat64{nulltype.NullFloat64Of(999), nulltype.NullFloat64Of(333)},
		},
		{
			name: "null values",
			input: []core.Measurement{
				{
					GaugeID: core.GaugeID{
						Script: "all_at_once",
						Code:   "a002",
					},
					Timestamp: core.HTime{Time: time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)},
					Level:     null,
					Flow:      null,
				},
				{
					GaugeID: core.GaugeID{
						Script: "all_at_once",
						Code:   "a002",
					},
					Timestamp: core.HTime{Time: time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC)},
					Level:     nulltype.NullFloat64Of(0),
					Flow:      nulltype.NullFloat64Of(0),
				},
				{
					GaugeID: core.GaugeID{
						Script: "all_at_once",
						Code:   "a002",
					},
					Timestamp: core.HTime{Time: time.Date(2015, time.January, 1, 12, 1, 0, 0, time.UTC)},
					Level:     nulltype.NullFloat64Of(777),
					Flow:      null,
				},
			},
			result:  1,
			content: []nulltype.NullFloat64{nulltype.NullFloat64Of(333), null},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetupTest()
			in := core.GenFromSlice(context.Background(), tt.input)
			savedCh, errCh := s.mgr.SaveMeasurements(context.Background(), in)
			cnt := <-savedCh
			err := <-errCh
			if assert.NoError(t, err) {
				assert.Equal(t, tt.result, cnt)
				actual := make([]nulltype.NullFloat64, 0)
				rows, err := s.mgr.db.Query("SELECT flow FROM measurements WHERE script = 'all_at_once' AND code = 'a002' ORDER BY timestamp DESC")
				if assert.NoError(t, err, "failed to fetch raw rows") {
					var v nulltype.NullFloat64
					for rows.Next() {
						err := rows.Scan(&v)
						assert.NoError(t, err, "failed to scan raw rows")
						actual = append(actual, v)
					}
				}
				assert.Equal(t, tt.content, actual)
			}
		})
	}

}

func (s *DbTestSuite) TestSaveMeasurementsChunks() {
	t := s.T()

	tests := []struct {
		total     int
		chunkSize int
	}{
		{total: 30, chunkSize: 0},
		{total: 30, chunkSize: 10},
		{total: 30, chunkSize: 1},
		{total: 30, chunkSize: 7},
		{total: 30, chunkSize: 50},
	}
	factory := core.MeasurementsFactory{
		Script: "all_at_once",
		Code:   "g003",
	}
	inp := factory.GenMany(30)

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d in chunks of %d", tt.total, tt.chunkSize), func(t *testing.T) {
			s.SetupTest()
			s.mgr.saveChunkSize = tt.chunkSize

			in := core.GenFromSlice(context.Background(), inp[:tt.total])
			savedCh, errCh := s.mgr.SaveMeasurements(context.Background(), in)
			cnt := <-savedCh
			err := <-errCh

			if assert.NoError(t, err) {
				assert.Equal(t, tt.total, cnt)
				var actual int
				if assert.NoError(t, s.mgr.db.QueryRowx("SELECT count(*) FROM measurements WHERE script = 'all_at_once' AND code = 'g003'").Scan(&actual)) {
					assert.Equal(t, tt.total, actual)
				}
			}
		})
	}
}

func (s *DbTestSuite) TestSaveMeasurementsCancel() {
	t := s.T()
	s.SetupTest()
	factory := core.MeasurementsFactory{
		Script: "all_at_once",
		Code:   "g003",
	}
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan *core.Measurement)
	savedCh, errCh := s.mgr.SaveMeasurements(ctx, in)
	in <- factory.GenOnePtr(1)
	cancel()
	cnt, cok := <-savedCh
	assert.Zero(t, cnt)
	assert.False(t, cok)
	err := <-errCh
	assert.Equal(t, err, context.Canceled)
	_, eok := <-errCh
	assert.False(t, eok)
}

func (s *DbTestSuite) TestListJobs() {
	t := s.T()
	expected := []core.JobDescription{
		{
			ID:      "01e99188-2189-11ea-978f-2e728ce88125",
			Script:  "all_at_once",
			Gauges:  map[string]json.RawMessage{"a001": []byte("{}"), "a002": []byte("{}")},
			Cron:    "1 * * * *",
			Options: json.RawMessage(`{"foo": "bar"}`),
		},
		{
			ID:      "0d67638c-2189-11ea-978f-2e728ce88125",
			Script:  "one_by_one",
			Gauges:  map[string]json.RawMessage{"o001": []byte(`{"foo": "bar"}`), "o002": []byte("{}")},
			Options: json.RawMessage(`{"foo": 42.0}`),
		},
	}
	actual, err := s.mgr.ListJobs()
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func (s *DbTestSuite) TestGetJobSuccess() {
	t := s.T()
	actual, err := s.mgr.GetJob("01e99188-2189-11ea-978f-2e728ce88125")
	expected := &core.JobDescription{
		ID:      "01e99188-2189-11ea-978f-2e728ce88125",
		Script:  "all_at_once",
		Gauges:  map[string]json.RawMessage{"a001": []byte("{}"), "a002": []byte("{}")},
		Cron:    "1 * * * *",
		Options: json.RawMessage(`{"foo": "bar"}`),
	}
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func (s *DbTestSuite) TestGetJobNotFound() {
	t := s.T()
	actual, err := s.mgr.GetJob("7ec72ae0-403c-11ea-b77f-2e728ce88125")
	if assert.NoError(t, err) {
		assert.Nil(t, actual)
	}
}

func (s *DbTestSuite) TestDeleteJobOk() {
	t := s.T()
	err := s.mgr.DeleteJob("0d67638c-2189-11ea-978f-2e728ce88125", func(id string) error {
		return nil
	})
	assert.NoError(t, err)
}

func (s *DbTestSuite) TestDeleteJobMissing() {
	t := s.T()
	err := s.mgr.DeleteJob("b2162fe8-218a-11ea-978f-2e728ce88125", func(id string) error {
		return nil
	})
	assert.Error(t, err)
}

func (s *DbTestSuite) TestDeleteJobCallbackError() {
	t := s.T()
	err := s.mgr.DeleteJob("0d67638c-2189-11ea-978f-2e728ce88125", func(id string) error {
		return errors.New("boom")
	})
	if assert.Error(t, err) {
		assert.Equal(t, "boom", err.Error())
		cnt := countJobs(s.mgr.db)
		assert.Equal(t, 2, cnt)
	}
}

func (s *DbTestSuite) TestAddJobOk() {
	t := s.T()
	input := core.JobDescription{
		ID:      "bbe699a2-0d36-4d3e-8581-542034e26602",
		Script:  "all_at_once",
		Gauges:  map[string]json.RawMessage{"a003": []byte("{}"), "a004": []byte("{}")},
		Cron:    "7 * * * *",
		Options: json.RawMessage(`{"foo": "baz"}`),
	}
	err := s.mgr.AddJob(input, func(job core.JobDescription) error {
		return nil
	})
	if assert.NoError(t, err) {
		cnt := countJobs(s.mgr.db)
		assert.Equal(t, 3, cnt)
	}
}

func (s *DbTestSuite) TestAddJobDupe() {
	t := s.T()
	input := core.JobDescription{
		ID:      "0d67638c-2189-11ea-978f-2e728ce88125",
		Script:  "all_at_once",
		Gauges:  map[string]json.RawMessage{"a003": []byte("{}"), "a004": []byte("{}")},
		Cron:    "7 * * * *",
		Options: json.RawMessage(`{"foo": "baz"}`),
	}
	err := s.mgr.AddJob(input, func(job core.JobDescription) error {
		return nil
	})
	if assert.Error(t, err) {
		cnt := countJobs(s.mgr.db)
		assert.Equal(t, 2, cnt)
	}
}

func (s *DbTestSuite) TestAddJobCallbackError() {
	t := s.T()
	input := core.JobDescription{
		ID:      "bbe699a2-0d36-4d3e-8581-542034e26602",
		Script:  "all_at_once",
		Gauges:  map[string]json.RawMessage{"a003": []byte("{}"), "a004": []byte("{}")},
		Cron:    "7 * * * *",
		Options: json.RawMessage(`{"foo": "baz"}`),
	}
	err := s.mgr.AddJob(input, func(job core.JobDescription) error {
		return errors.New("boom")
	})
	if assert.Error(t, err) {
		assert.Equal(t, "boom", err.Error())
		cnt := countJobs(s.mgr.db)
		assert.Equal(t, 2, cnt)
	}
}
