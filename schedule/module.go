package schedule

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/storage"
	"go.uber.org/fx"
)

type SchedulerParams struct {
	fx.In

	Cron     Cron
	Database storage.DatabaseManager
	Cache    storage.CacheManager
	Registry *core.ScriptRegistry
	Logger   *logrus.Logger
}

func newCron() Cron {
	return cron.New(cron.WithLocation(time.UTC))
}

func newSimpleScheduler(lc fx.Lifecycle, p SchedulerParams) core.JobScheduler {
	scheduler := &simpleScheduler{
		Database: p.Database,
		Cache:    p.Cache,
		Registry: p.Registry,
		Cron:     p.Cron,
		Logger:   p.Logger.WithField("logger", "scheduler"),
	}
	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			scheduler.Logger.Debug("starting")
			scheduler.Start()

			// Load initial jobs
			jobs, err := scheduler.Database.ListJobs()
			if err != nil {
				scheduler.Logger.Fatalf("failed to load initial jobs: %v", err)
			}
			for _, job := range jobs {
				err := scheduler.AddJob(job)
				if err != nil {
					scheduler.Logger.Errorf("failed to schedule initial jobs: %v", err)
					return err
				}
				scheduler.Logger.WithFields(logrus.Fields{"script": job.Script, "jobID": job.ID}).Info("started job")
			}

			scheduler.Logger.Info("started")
			return nil
		},
		OnStop: func(c context.Context) error {
			scheduler.Logger.Debug("stopping")
			scheduler.Stop()
			scheduler.Logger.Info("stopped")
			return nil
		},
	})
	return scheduler
}

var Module = fx.Provide(
	newCron,
	newSimpleScheduler,
)

var TestModule = fx.Provide(
	newImmediateCron,
	newSimpleScheduler,
)
