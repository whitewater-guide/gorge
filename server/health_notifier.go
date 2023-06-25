package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
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
	job.logger.Info("running health notifier")

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
	scripts := []string{}

	// Jobs that didn't return measurements within last 48 hours are considered unhealthy
	threshold := time.Now().Add(-time.Duration(job.cfg.Threshold) * time.Hour)

	for _, j := range jobs {
		if status, ok := statuses[j.ID]; ok {
			// because this is supposed to run daily and our job are scheduled hourly or more frequently,
			// having last success == nil for a day is cosidered unhealthy
			// it can produce some misfires when service restarts, but it's ok
			if status.LastSuccess == nil || status.LastSuccess.Before(threshold) {
				res = append(res, core.UnhealthyJob{
					JobID:       j.ID,
					Script:      j.Script,
					LastRun:     status.LastRun,
					LastSuccess: status.LastSuccess,
				})
				scripts = append(scripts, j.Script)
			}
		}
	}

	job.logger.Infof("found %d unhealthy jobs: %s", len(res), strings.Join(scripts, ", "))

	if len(res) == 0 {
		return
	}

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
		parts := strings.Split(h, ":")
		if len(parts) != 2 {
			job.logger.Warnf("invalid header name-value pair: %s", h)
			continue
		}
		// it's possible to use env variables in header values
		// e.g. '--hooks-health-headers "x-api-key: $GORGE_HEALTH_KEY"'
		req.Header.Set(strings.TrimSpace(parts[0]), os.ExpandEnv(strings.TrimSpace(parts[1])))
	}

	resp, err := core.Client.Do(req, &core.RequestOptions{})
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		job.logger.Infof("notified about unhealthy jobs, got %d %s", resp.StatusCode, string(body))
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
			eId, err := p.Cron.AddJob(job.cfg.Cron, job)
			if err == nil {
				entry := p.Cron.Entry(eId)
				log.Infof("started notfier with cron expression '%s', next run at '%v'", job.cfg.Cron, entry.Next.UTC())
			}
			return err
		},
	})
}
