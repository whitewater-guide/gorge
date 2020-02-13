package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/schedule"
	"github.com/whitewater-guide/gorge/scripts/testscripts"
)

func TestStartup(t *testing.T) {
	ja := jsonassert.New(t)
	registry := core.NewRegistry()
	registry.Register(testscripts.AllAtOnce)
	registry.Register(testscripts.Broken)
	registry.Register(testscripts.OneByOne)
	srv := newServer(testConfig(), registry)
	srv.scheduler = &schedule.SimpleScheduler{
		Database: srv.database,
		Cache:    srv.cache,
		Registry: srv.registry,
		Cron:     &schedule.ImmediateCron{},
		Logger:   srv.logger,
	}
	err := srv.database.AddJob(core.JobDescription{
		ID:     "24e45a47-7ae2-453a-afa3-153392e2460b",
		Script: "all_at_once",
		Gauges: map[string]json.RawMessage{"g000": []byte("{}")},
		Cron:   "* * * * *",
	}, func(job core.JobDescription) error { return nil })
	if err != nil {
		t.Fatalf("failed to seed startup jobs %v", err)
	}

	srv.routes()
	srv.start()
	defer srv.shutdown()

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
	err = json.Unmarshal([]byte(resp), &data)
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
