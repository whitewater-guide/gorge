package schedule

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/scripts/testscripts"
	"github.com/whitewater-guide/gorge/storage"
)

type counterCron struct {
	mock.Mock
	counter cron.EntryID
	entries map[cron.EntryID]bool
}

func (c *counterCron) Entries() []cron.Entry {
	// no-op
	return nil
}

func (i *counterCron) Entry(id cron.EntryID) cron.Entry {
	return cron.Entry{}
}

func (c *counterCron) Remove(id cron.EntryID) {
	delete(c.entries, id)
}

func (c *counterCron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	if c.entries == nil {
		c.entries = make(map[cron.EntryID]bool)
	}
	c.entries[c.counter] = true
	result := c.counter
	c.counter++
	return result, nil
}

func (c *counterCron) Start() {
	// noop
}

func (c *counterCron) Stop() context.Context {
	withCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	return withCancel
}

type mockCron struct {
	mock.Mock
}

func (m *mockCron) Entries() []cron.Entry {
	args := m.Called()
	return args.Get(0).([]cron.Entry)
}

func (i *mockCron) Entry(id cron.EntryID) cron.Entry {
	return cron.Entry{}
}

func (m *mockCron) Remove(id cron.EntryID) {
	m.Called(id)
}

func (m *mockCron) AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	args := m.Called(spec, cmd)
	return cron.EntryID(args.Int(0)), args.Error(1)
}

func (m *mockCron) Start() {
	// noop
}

func (m *mockCron) Stop() context.Context {
	withCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	return withCancel
}

type mockScheduler struct {
	*simpleScheduler
}

func newMockScheduler(t *testing.T) *mockScheduler {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)

	db := storage.NewSqliteDb(logrus.NewEntry(logger), 0)
	cache := &storage.EmbeddedCacheManager{}
	registry := core.NewRegistry()
	registry.Register(testscripts.AllAtOnce)
	registry.Register(testscripts.OneByOne)
	registry.Register(testscripts.Batched)
	registry.Register(testscripts.Broken)

	return &mockScheduler{
		simpleScheduler: &simpleScheduler{
			Database: db,
			Cache:    cache,
			Cron:     &mockCron{},
			Logger:   logrus.NewEntry(logger),
			Registry: registry,
		},
	}
}
