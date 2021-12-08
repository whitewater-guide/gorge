package schedule

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/storage"
)

// cronLike interface is subset of cron https://github.com/robfig/cron
// it's used to run harvesting jobs on schedule
// various test helpers implement it
type cronLike interface {
	AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
	Entries() []cron.Entry
	Remove(id cron.EntryID)
	Start()
	Stop() context.Context
}

// SimpleScheduler is implementation of JobScheduler with cron scheduler
// it will run jobs and save their results and statuses
type SimpleScheduler struct {
	Database storage.DatabaseManager
	Cache    storage.CacheManager
	Registry *core.ScriptRegistry
	Cron     cronLike
	Logger   *logrus.Entry
}

// Start implements JobScheduler interface
func (s *SimpleScheduler) Start() {
	s.Logger.Info("starting")
	s.Cron.Start()
}

// Stop implements JobScheduler interface
func (s *SimpleScheduler) Stop() {
	s.Logger.Info("stopping")
	schedCtx := s.Cron.Stop()
	<-schedCtx.Done()
}
