package core

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// JobDescription represents task that will run on schedule and harvest measurements from some source
type JobDescription struct {
	// UUID. Represents paired entity id in client's service and are generated on client side
	ID string `json:"id" structs:"id"`
	// id of script from gorge's script registry
	Script string `json:"script" structs:"script"`
	// a map with keys being gauge codes and value being pieces of json representing harvest options for that gauge
	// pass `{}` or `null` if no options are given for the gauge
	Gauges map[string]json.RawMessage `json:"gauges" structs:"codes" ts_type:"{[key: string]: any} | null"`
	// cron expression, ignored for OneByOne scripts. AllAtOnce script will run on this cron schedule
	Cron string `json:"cron" structs:"cron"`
	// harvest options for the entire script. For example, upstream credentials
	Options json.RawMessage `json:"options" structs:"options,omitempty" ts_type:"{[key: string]: any} | null"`
	// When used as input this must be nil
	Status *Status `json:"status,omitempty"`
}

// JobScheduler is responsible for running harvest jobs on schedule
type JobScheduler interface {
	Start()
	Stop()
	AddJob(description JobDescription) error
	DeleteJob(jobID string) error
	// ListNext returns map where values are times when scripts will run next time
	// If jobID is empty, ListNext lists next times for all running scripts. And map keys are script ids
	// If jobID is not empty, this will return next times for all codes of this one-by-one job, and map keys are gauge codes
	ListNext(jobID string) map[string]HTime
}

// Bind implements go-chi Binder interface
func (j *JobDescription) Bind(r *http.Request) error {
	_, err := uuid.Parse(j.ID)
	if err != nil {
		e := &Error{
			Err: nil,
			Msg: "job id must be valid uuid",
		}
		return e.With("jobId", j.ID)
	}
	return nil
}

// Scan implements sql nner interface https://golang.org/pkg/database/sql/#Scanner
func (j *JobDescription) Scan(src interface{}) error {
	var source []byte
	switch src := src.(type) {
	case string:
		source = []byte(src)
	case []byte:
		source = src
	default:
		return errors.New("Incompatible type for JobDescription")
	}
	return json.Unmarshal(source, j)
}

// Status contains information about job's last execution
type Status struct {
	// When was this job executed last time (has nothing to do with measurements timestamps)
	LastRun HTime `json:"lastRun" ts_type:"string"`
	// Error, if last job execution failed
	Error string `json:"error,omitempty"`
	// Number of measurements harvested during last execution (0 in case of error)
	Count int `json:"count"`
	// When did this job run successfully (collected some measurements) last time
	// Is less or equal than Timestamp, or nil pointer if never ran successfully
	LastSuccess *HTime `json:"lastSuccess,omitempty" ts_type:"string"`
	// When will this job run next time
	NextRun *HTime `json:"nextRun,omitempty" ts_type:"string"`
}

// UnhealthyJob describes a job that haven't run successfully for a certain period of time
// It's used to notify interested parties
type UnhealthyJob struct {
	// JobID is same as in job description
	JobID string `json:"id"`
	// Script of the job
	Script string `json:"script"`
	// When was this job executed last time (has nothing to do with measurements timestamps)
	LastRun HTime `json:"lastRun" ts_type:"string"`
	// When did this job run successfully (collected some measurements) last time
	// Is less or equal than LastRun, or nil pointer if never ran successfully
	LastSuccess *HTime `json:"lastSuccess,omitempty" ts_type:"string"`
}
