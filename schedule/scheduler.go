package schedule

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/storage"
)

// Cron is a subset of Cron from  https://github.com/robfig/cron
// It's extracted into interface so it can be mocked
type Cron interface {
	AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
	Entries() []cron.Entry
	Remove(id cron.EntryID)
	Start()
	Stop() context.Context
}

// simpleScheduler is implementation of core.JobScheduler with cron scheduler
// it will run jobs and save their results and statuses
type simpleScheduler struct {
	Database storage.DatabaseManager
	Cache    storage.CacheManager
	Registry *core.ScriptRegistry
	Cron     Cron
	Logger   *logrus.Entry
}

// Start implements core.JobScheduler interface
func (s *simpleScheduler) Start() {
	s.Logger.Info("starting")
	s.Cron.Start()
}

// Stop implements core.JobScheduler interface
func (s *simpleScheduler) Stop() {
	s.Logger.Info("stopping")
	schedCtx := s.Cron.Stop()
	<-schedCtx.Done()
}
