package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/config"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/schedule"
	"github.com/whitewater-guide/gorge/scripts"
	"github.com/whitewater-guide/gorge/storage"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func seedStartupTest(db storage.DatabaseManager) error {
	return db.AddJob(core.JobDescription{
		ID:     "24e45a47-7ae2-453a-afa3-153392e2460b",
		Script: "all_at_once",
		Gauges: map[string]json.RawMessage{"g000": []byte("{}")},
		Cron:   "* * * * *",
	}, func(job core.JobDescription) error { return nil })
}

func TestStartup(t *testing.T) {
	ja := jsonassert.New(t)

	var srv *Server
	app := fx.New(
		fx.Invoke(func(lc fx.Lifecycle, db storage.DatabaseManager) {
			// First startup hook: seed database
			lc.Append(fx.Hook{
				OnStart: func(c context.Context) error {
					return seedStartupTest(db)
				},
			})
		}),
		fx.Options(
			config.TestModule,
			fx.Provide(testLogger),
			scripts.TestModule,
			storage.Module,
			schedule.Module,
			fx.Provide(newServer),
		),
		fx.Invoke(func(s *Server) {
			srv = s
			srv.scheduler.(*schedule.SimpleScheduler).Cron = &schedule.ImmediateCron{}
			srv.routes()
		}),
		fx.WithLogger(
			func() fxevent.Logger {
				return fxevent.NopLogger
			},
		),
	)
	if err := app.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer app.Stop(context.Background())

	resp, _ := runCase(t, srv, test{
		method: "GET",
		path:   "/jobs",
	})

	ja.Assertf(resp, `[{
		"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
		"script": "all_at_once",
		"gauges": {"g000": {}},
		"cron": "* * * * *",
		"options": null,
		"status": {
			"success": true,
			"timestamp": "<<PRESENCE>>", 
			"count": 1
		}
	}]`)

	resp, _ = runCase(t, srv, test{
		method: "GET",
		path:   "/jobs/24e45a47-7ae2-453a-afa3-153392e2460b",
	})

	ja.Assertf(resp, `{
		"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
		"script": "all_at_once",
		"gauges": {"g000": {}},
		"cron": "* * * * *",
		"options": null
	}`)

	resp, _ = runCase(t, srv, test{
		method: "GET",
		path:   "/jobs/24e45a47-7ae2-453a-afa3-153392e2460b/gauges",
	})

	ja.Assertf(resp, `{
		"g000": {
			"success": true,
			"timestamp": "<<PRESENCE>>", 
			"count": 1
		}
	}`)
	// time.Sleep(600 * time.Millisecond) // this place is flaky

	resp, _ = runCase(t, srv, test{
		method: "GET",
		path:   "/measurements/all_at_once/g000",
	})

	var data []core.Measurement
	err := json.Unmarshal([]byte(resp), &data)
	if assert.NoError(t, err) && assert.GreaterOrEqual(t, len(data), 1) {
		assert.Equal(t, "all_at_once", data[0].Script)
		assert.Equal(t, "g000", data[0].Code)
		assert.InDelta(t, time.Now().UTC().Add(-1*time.Second).Unix(), data[0].Timestamp.Unix(), 500)
	}

	resp, _ = runCase(t, srv, test{
		method: "GET",
		path:   "/measurements/latest?scripts=all_at_once,one_by_one",
	})

	err = json.Unmarshal([]byte(resp), &data)
	if assert.NoError(t, err) {
		assert.Len(t, data, 1)
		assert.Equal(t, "all_at_once", data[0].Script)
		assert.Equal(t, "g000", data[0].Code)
		assert.InDelta(t, time.Now().UTC().Add(-1*time.Second).Unix(), data[0].Timestamp.Unix(), 500)
	}
}
