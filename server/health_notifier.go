package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/config"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/schedule"
	"github.com/whitewater-guide/gorge/storage"
	"go.uber.org/fx"
)

type healthNotifierParams struct {
	fx.In

	Logger *logrus.Logger
	Db     storage.DatabaseManager
	Cache  storage.CacheManager
	Cfg    *config.Config
	Cron   schedule.Cron
}

type healthNotifierJob struct {
	cfg      config.HealthConfig
	database storage.DatabaseManager
	cache    storage.CacheManager
	logger   *logrus.Entry
}

func (job healthNotifierJob) Run() {
	job.logger.Debug("running health notifier job")

	jobs, err := job.database.ListJobs()

	if err != nil {
		job.logger.Errorf("failed to list jobs: %v", err)
		return
	}

	statuses, err := job.cache.LoadJobStatuses()
	if err != nil {
		job.logger.Errorf("failed to load job statuses: %v", err)
		return
	}

	res := []core.UnhealthyJob{}

	// Jobs that didn't return measurements within last 48 hours are considered unhealthy
	threshold := time.Now().Add(-time.Duration(job.cfg.Threshold) * time.Hour)

	for _, j := range jobs {
		if status, ok := statuses[j.ID]; ok {
			if status.LastSuccess == nil || status.LastSuccess.Before(threshold) {
				res = append(res, core.UnhealthyJob{
					JobID:       j.ID,
					Script:      j.Script,
					LastRun:     status.LastRun,
					LastSuccess: status.LastSuccess,
				})
			}
		}
	}

	job.logger.Debugf("found %d unhelathy jobs", len(res))

	if len(res) == 0 {
		return
	}

	job.logger.Debug("notifying about unhealthy jobs")
	msg, err := json.Marshal(res)
	if err != nil {
		job.logger.Errorf("failed to marshal unhealthy jobs: %v", err)
		return
	}

	req, err := http.NewRequest("POST", job.cfg.URL, bytes.NewBuffer(msg))
	if err != nil {
		job.logger.Errorf("failed to create http request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	for _, h := range job.cfg.Headers {
		parts := strings.Split(h, ": ")
		if len(parts) != 2 {
			job.logger.Warnf("invalid header name-value pair: %s", h)
			continue
		}
		req.Header.Set(parts[0], parts[1])
	}

	_, err = core.Client.Do(req, &core.RequestOptions{})
	if err == nil {
		job.logger.Debug("notified about unhealthy jobs")
	} else {
		job.logger.Errorf("failed to notify about unhealthy jobs: %v", err)
	}
}

func startHealthNotifier(lc fx.Lifecycle, p healthNotifierParams) {
	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			log := p.Logger.WithField("logger", "health")

			if p.Cfg.Endpoint == "" {
				log.Debug("health webhook url not configured")
				return nil
			}

			log.Debugf("starting")
			job := healthNotifierJob{
				cfg:      p.Cfg.Hooks.Health,
				database: p.Db,
				cache:    p.Cache,
				logger:   log,
			}
			_, err := p.Cron.AddJob(job.cfg.Cron, job)
			if err == nil {
				log.Infof("started notfier at '%s'", job.cfg.Cron)
			}
			return err
		},
	})
}
