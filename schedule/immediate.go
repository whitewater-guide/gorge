package schedule

import (
	"context"

	"github.com/robfig/cron/v3"
)

// ImmediateCron is cronLike implementation for test purposes
// it will execute job imeediately after it has been added
type ImmediateCron struct{}

// AddJob implements schedule.Cron interface
func (i *ImmediateCron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	cmd.Run()
	return 0, nil
}

// Entries implements schedule.Cron interface
func (i *ImmediateCron) Entries() []cron.Entry {
	return []cron.Entry{}
}

// Entry implements schedule.Cron interface
func (i *ImmediateCron) Entry(id cron.EntryID) cron.Entry {
	return cron.Entry{}
}

// Remove implements schedule.Cron interface
func (i *ImmediateCron) Remove(id cron.EntryID) {}

// Start implements schedule.Cron interface
func (i *ImmediateCron) Start() {}

// Stop implements schedule.Cron interface
func (i *ImmediateCron) Stop() context.Context {
	withCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	return withCancel
}

func newImmediateCron() Cron {
	return &ImmediateCron{}
}
