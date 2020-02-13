package schedule

import (
	"context"

	"github.com/robfig/cron/v3"
)

// ImmediateCron is cronLike implementation for test purposes
// it will execute job imeediately after it has been added
type ImmediateCron struct{}

// AddJob implement cronLike interface
func (i *ImmediateCron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	cmd.Run()
	return 0, nil
}

// Entries implement cronLike interface
func (i *ImmediateCron) Entries() []cron.Entry {
	return []cron.Entry{}
}

// Remove implement cronLike interface
func (i *ImmediateCron) Remove(id cron.EntryID) {}

// Start implement cronLike interface
func (i *ImmediateCron) Start() {}

// Stop implement cronLike interface
func (i *ImmediateCron) Stop() context.Context {
	withCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	return withCancel
}
