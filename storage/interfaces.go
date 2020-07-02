package storage

import (
	"context"
	"time"

	"github.com/whitewater-guide/gorge/core"
)

// DatabaseManager is used to store all harvested measurements
// it's also used to store jobs so that they persist between service restarts
type DatabaseManager interface {
	// ListJobs returns slice of currently active jobs
	ListJobs() ([]core.JobDescription, error)
	// GetJob returns active job by its id
	GetJob(id string) (*core.JobDescription, error)
	// AddJob creates new job from description and starts it immediately
	// onSave argument is used to ensure transactional behavior when adding job to scheduler
	AddJob(job core.JobDescription, onSave func(job core.JobDescription) error) error
	// DeleteJon stops running job and deletes it
	// onDelete argument is used to ensure transactional behavior when adding job to scheduler
	DeleteJob(id string, onDelete func(id string) error) error

	// SaveMeasurements saves measurements from the channel in db, until the channel is closed
	// It supports context cancelation
	// returns channel where one single int will be written: total number of saved mesurements
	SaveMeasurements(ctx context.Context, in <-chan *core.Measurement) (<-chan int, <-chan error)
	// GetMeasurements returns measurements stored in db
	GetMeasurements(query MeasurementsQuery) ([]core.Measurement, error)
	// GetNearestMeasurement returns nearest measurement to timestamp (without interpolation)
	GetNearestMeasurement(script, code string, to time.Time, tolerance time.Duration) (*core.Measurement, error)

	// Close is called when db should be shut down
	Close()
}

// CacheManager manager is used to store latest measurement for each gauge and auxiliary information that is safe to lose
type CacheManager interface {
	// LoadJobStatuses returns statuses of currenly running jobs.
	// returns map where keys are job ids
	LoadJobStatuses() (map[string]core.Status, error)
	// LoadGaugeStatuses returns statuses of gauges for given job
	// returns map where keys are gauge codes
	LoadGaugeStatuses(jobID string) (map[string]core.Status, error)
	// SaveStatus saves harvest status for entire job (if code is empty) or single gauge
	// count means number of saved measurements
	SaveStatus(jobID, code string, err error, count int) error

	// LoadLatestMeasurements returns latest measurements
	// it accepts a map where keys are scripts (not job ids!) and values are sets of gauge codes
	LoadLatestMeasurements(from map[string]core.StringSet) (map[core.GaugeID]core.Measurement, error)
	// SaveLatestMeasurements saves given measurements. If there're multiple values per gauge, the most recent one will be saved
	// Input measurements are supposed to be filtered against previous latest values from cache
	// This is done inside job (it also ensures we don't save dupe measurements in db)
	SaveLatestMeasurements(ctx context.Context, in <-chan *core.Measurement) <-chan error

	// Close is callled when cache must be shut down
	Close()
}
