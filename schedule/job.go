package schedule

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/storage"
)

type harvestJob struct {
	database storage.DatabaseManager
	cache    storage.CacheManager
	registry *core.ScriptRegistry
	logger   *logrus.Logger
	jobID    string
	cron     string
	codes    core.StringSet
	script   string
	options  interface{}
}

func getSince(job *harvestJob, cache map[core.GaugeID]core.Measurement) int64 {
	code, err := job.codes.Only()
	if err == nil {
		key := core.GaugeID{Script: job.script, Code: code}
		if cached, ok := cache[key]; ok {
			return cached.Timestamp.UTC().Unix()
		}
	}
	return 0
}

func logError(logger *logrus.Entry, err error) {
	if e, ok := err.(*core.Error); ok {
		logger.WithFields(e.Ctx).Error(e)
	} else {
		logger.Error(e)
	}
}

func (job harvestJob) Run() {
	logger := job.logger.WithField("script", job.script).WithField("id", job.jobID)
	code, _ := job.codes.Only()
	if code != "" {
		logger = logger.WithField("code", code)
	}
	defer func() {
		if r := recover(); r != nil {
			logger.Error(r)
		}
	}()

	// get last values from redis cache
	cache, err := job.cache.LoadLatestMeasurements(map[string]core.StringSet{job.script: job.codes})
	if err != nil {
		job.logger.Warnf("failed to load last measurements from cache: %v", err)
	}

	script, _, err := job.registry.Create(job.script, job.options)
	if err != nil {
		logError(logger, err)
		ssErr := job.cache.SaveStatus(job.script, code, err, 0)
		if ssErr != nil {
			logError(logger, ssErr)
		}
		return
	}
	script.SetLogger(logger)

	in := make(chan *core.Measurement)
	errCh := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	go script.Harvest(ctx, in, errCh, job.codes, getSince(&job, cache))
	filteredCh := core.FilterMeasurements(
		ctx,
		in,
		logger,
		core.CodesFilter{Codes: job.codes},
		core.LatestFilter{Latest: cache, After: time.Now().Add(time.Duration(-30*24) * time.Hour)},
	)
	cacheIn, dbIn := core.Split(ctx, filteredCh)
	savedCh, savedErrCh := job.database.SaveMeasurements(ctx, dbIn)
	cachedErrCh := job.cache.SaveLatestMeasurements(ctx, cacheIn)
	harvestErr, saved, savedErr, cachedErr := <-errCh, <-savedCh, <-savedErrCh, <-cachedErrCh

	statusErr := <-errCh
	if statusErr == nil {
		statusErr = savedErr
	}
	if statusErr == nil {
		statusErr = cachedErr
	}
	ssErr := job.cache.SaveStatus(job.jobID, code, statusErr, saved)
	if ssErr != nil {
		logError(logger, core.WrapErr(ssErr, "save status error"))
	}

	if harvestErr != nil {
		logError(logger, core.WrapErr(harvestErr, "harvest error"))
	}
	if savedErr != nil {
		logError(logger, core.WrapErr(savedErr, "db save error"))
	}
	if cachedErr != nil {
		logError(logger, core.WrapErr(cachedErr, "cache save error"))
	}
	if saved == 0 {
		logger.Warn("saved 0 measurements")
	} else {
		logger.Debugf("saved %d measurements", saved)
	}
}
