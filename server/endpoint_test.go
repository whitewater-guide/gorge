package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
	"github.com/mattn/go-nulltype"
	"github.com/stretchr/testify/assert"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/scripts/testscripts"
)

type test struct {
	name   string
	path   string
	method string
	body   string
	// expected
	code int
	resp string
}

func prepareServer() *server {
	registry := core.NewRegistry()
	registry.Register(testscripts.AllAtOnce)
	registry.Register(testscripts.Broken)
	registry.Register(testscripts.OneByOne)
	srv := newServer(testConfig(), registry)

	// nolint:errcheck
	srv.database.AddJob(core.JobDescription{
		ID:      "48f979ec-268b-11ea-978f-2e728ce88125",
		Script:  "all_at_once",
		Gauges:  map[string]json.RawMessage{"g000": json.RawMessage("{}")}, // no data will be saved from this job because these gauge codes do not exist
		Cron:    "0 0 * * *",
		Options: json.RawMessage(`{"gauges": 11 }`),
	}, func(job core.JobDescription) error {
		return nil
	})
	srv.database.SaveMeasurements(context.Background(), core.GenFromSlice(context.Background(), []core.Measurement{
		{
			GaugeID: core.GaugeID{
				Script: "broken",
				Code:   "g000",
			},
			Timestamp: core.HTime{Time: time.Now().Add(-1 * time.Hour).UTC()},
			Level:     nulltype.NullFloat64Of(-100),
			Flow:      nulltype.NullFloat64Of(-100),
		},
	}))
	srv.cache.SaveStatus("48f979ec-268b-11ea-978f-2e728ce88125", "g000", nil, 10)                     // nolint:errcheck
	srv.cache.SaveStatus("48f979ec-268b-11ea-978f-2e728ce88125", "g001", errors.New("test error"), 0) // nolint:errcheck
	srv.cache.SaveLatestMeasurements(context.Background(), core.GenFromSlice(context.Background(), []core.Measurement{
		{
			GaugeID: core.GaugeID{
				Script: "broken",
				Code:   "g000",
			},
			Timestamp: core.HTime{Time: time.Now().Add(-1 * time.Hour).UTC()},
			Level:     nulltype.NullFloat64Of(-100),
			Flow:      nulltype.NullFloat64Of(-100),
		},
	}))

	srv.routes()
	srv.start()
	time.Sleep(10 * time.Millisecond)

	return srv
}

func runCase(t *testing.T, srv *server, tt test) (string, int) {
	ts := httptest.NewServer(srv.router)
	defer ts.Close()

	req, err := http.NewRequest(tt.method, fmt.Sprintf("%s%s", ts.URL, tt.path), bytes.NewBufferString(tt.body))
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}
	return string(body), res.StatusCode
}

func TestEndpoint(t *testing.T) {
	tests := []test{
		{
			name: "list scripts",
			path: "/scripts",
			resp: `[
				{"name": "all_at_once", "mode": "allAtOnce", "description": "Test script for all at once harvesting mode"},
				{"name": "broken", "mode": "allAtOnce", "description": "Test script that always returns error"},
				{"name": "one_by_one", "mode": "oneByOne", "description": "Test script for one by one harvesting mode"}
			]`,
		},
		{
			name:   "upstream gauges - success",
			method: "POST",
			path:   "/upstream/all_at_once/gauges",
			body:   `{"gauges": 1}`,
			resp: `[{
				"name": "Test gauge #0",
				"script": "all_at_once",
				"code": "g000",
				"url": "http://whitewater.guide/gauges/0",
				"flowUnit": "m3/s",
				"levelUnit": "m",
				"location": {
					"altitude": "<<PRESENCE>>",
					"latitude": "<<PRESENCE>>",
					"longitude": "<<PRESENCE>>"
				}
			}]`,
		},
		{
			name:   "upstream gauges - no location",
			method: "POST",
			path:   "/upstream/all_at_once/gauges",
			body:   `{"gauges": 1, "noLocation": true}`,
			resp: `[{
				"name": "Test gauge #0",
				"script": "all_at_once",
				"code": "g000",
				"url": "http://whitewater.guide/gauges/0",
				"flowUnit": "m3/s",
				"levelUnit": "m"
			}]`,
		},
		{
			name:   "upstream gauges - no altitude",
			method: "POST",
			path:   "/upstream/all_at_once/gauges",
			body:   `{"gauges": 1, "noAltitude": true}`,
			resp: `[{
				"name": "Test gauge #0",
				"script": "all_at_once",
				"code": "g000",
				"url": "http://whitewater.guide/gauges/0",
				"flowUnit": "m3/s",
				"levelUnit": "m",
				"location": {
					"latitude": "<<PRESENCE>>",
					"longitude": "<<PRESENCE>>"
				}
			}]`,
		},
		{
			name:   "upstream gauges - bad script",
			method: "POST",
			path:   "/upstream/foo/gauges",
			code:   http.StatusNotFound,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream gauges - bad payload",
			method: "POST",
			body:   "foo",
			path:   "/upstream/all_at_once/gauges",
			code:   http.StatusInternalServerError,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream gauges - internal error",
			method: "POST",
			body:   `{"gauges": "foo"}`,
			path:   "/upstream/all_at_once/gauges",
			code:   http.StatusInternalServerError,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream gauges - broken script",
			method: "POST",
			body:   "{}",
			path:   "/upstream/broken/gauges",
			code:   http.StatusInternalServerError,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream measurements - bad script",
			path:   "/upstream/foo/measurements",
			method: "POST",
			code:   http.StatusNotFound,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream measurements - broken script",
			path:   "/upstream/broken/measurements",
			method: "POST",
			code:   http.StatusInternalServerError,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream measurements - bad payload",
			path:   "/upstream/all_at_once/measurements",
			method: "POST",
			body:   "foo",
			code:   http.StatusInternalServerError,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream measurements - bad query",
			path:   "/upstream/all_at_once/measurements?since=foooo",
			method: "POST",
			body:   "{}",
			code:   http.StatusBadRequest,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream measurements - no code for one-by-one script",
			path:   "/upstream/one_by_one/measurements",
			method: "POST",
			code:   http.StatusMethodNotAllowed,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream measurements - many codes for one-by-one",
			path:   "/upstream/one_by_one/measurements?codes=g000,g001",
			method: "POST",
			code:   http.StatusMethodNotAllowed,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "upstream measurements - one_by_one success",
			path:   "/upstream/one_by_one/measurements?codes=g000",
			method: "POST",
			resp: `[{
				"script": "one_by_one",
				"code": "g000",
				"timestamp": "<<PRESENCE>>",
				"flow": "<<PRESENCE>>",
				"level": "<<PRESENCE>>"
			}]`,
		},
		{
			name:   "upstream measurements - all_at_once success, all gauges",
			path:   "/upstream/all_at_once/measurements",
			method: "POST",
			body:   `{"gauges": 3}`,
			resp: `[
				{
					"script": "all_at_once",
					"code": "g000",
					"timestamp": "<<PRESENCE>>",
					"flow": "<<PRESENCE>>",
					"level": "<<PRESENCE>>"
				},
				{
					"script": "all_at_once",
					"code": "g001",
					"timestamp": "<<PRESENCE>>",
					"flow": "<<PRESENCE>>",
					"level": "<<PRESENCE>>"
				},
				{
					"script": "all_at_once",
					"code": "g002",
					"timestamp": "<<PRESENCE>>",
					"flow": "<<PRESENCE>>",
					"level": "<<PRESENCE>>"
				}
			]`,
		},
		{
			name:   "upstream measurements - all_at_once success, some gauges",
			path:   "/upstream/all_at_once/measurements?codes=g000,g002",
			method: "POST",
			body:   `{"gauges": 3}`,
			resp: `[
				{
					"script": "all_at_once",
					"code": "g000",
					"timestamp": "<<PRESENCE>>",
					"flow": "<<PRESENCE>>",
					"level": "<<PRESENCE>>"
				},
				{
					"script": "all_at_once",
					"code": "g002",
					"timestamp": "<<PRESENCE>>",
					"flow": "<<PRESENCE>>",
					"level": "<<PRESENCE>>"
				}
			]`,
		},
		{
			name: "list jobs",
			path: "/jobs",
			resp: `[{
					"id": "48f979ec-268b-11ea-978f-2e728ce88125",
					"script": "all_at_once",
					"gauges":  {"g000": {}},
					"cron":   "0 0 * * *",
					"options": {"gauges": 11},
					"status": {
						"success": false,
						"error": "test error",
						"timestamp": "<<PRESENCE>>",
						"count": 0,
						"next": "<<PRESENCE>>"
					}
			}]`,
		},
		{
			name: "get job",
			path: "/jobs/48f979ec-268b-11ea-978f-2e728ce88125",
			resp: `{
					"id": "48f979ec-268b-11ea-978f-2e728ce88125",
					"script": "all_at_once",
					"gauges":  {"g000": {}},
					"cron":   "0 0 * * *",
					"options": {"gauges": 11}
			}`,
		},
		{
			name: "get job gauges",
			path: "/jobs/48f979ec-268b-11ea-978f-2e728ce88125/gauges",
			resp: `{
				"g000": {
					"success": true,
					"timestamp": "<<PRESENCE>>",
					"next": "<<PRESENCE>>",
					"count": 10
				},
				"g001": {
					"success": false,
					"error": "test error",
					"timestamp": "<<PRESENCE>>",
					"count": 0
				}
			}`,
		},
		{
			name:   "add job - all_at_once success",
			method: "POST",
			body: `{
				"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
				"script": "all_at_once",
				"gauges": {"g001": {}, "g002": {}},
				"cron": "* * * * *",
				"options": {"gauges": 3}
			}`,
			path: "/jobs",
			resp: `{
				"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
				"script": "all_at_once",
				"gauges": {"g001": {}, "g002": {}},
				"cron": "* * * * *",
				"options": {"gauges": 3}
			}`,
		},
		{
			name:   "add job - one_by_one success",
			method: "POST",
			body: `{
				"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
				"script": "one_by_one",
				"gauges": {"g001": {"value": 100}, "g002": {}},
				"cron": "* * * * *",
				"options": {"gauges": 3}
			}`,
			path: "/jobs",
			resp: `{
				"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
				"script": "one_by_one",
				"gauges": {"g001": {"value": 100}, "g002": {}},
				"cron": "* * * * *",
				"options": {"gauges": 3} 
			}`,
		},
		{
			name:   "add job - one_by_one with nulls success",
			method: "POST",
			body: `{
				"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
				"script": "one_by_one",
				"gauges": {"g001": {"value": 100}, "g002": null},
				"cron": null,
				"options": null
			}`,
			path: "/jobs",
			resp: `{
				"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
				"script": "one_by_one",
				"gauges": {"g001": {"value": 100}, "g002": null},
				"cron": "",
				"options": null
			}`,
		},
		{
			name:   "add job - bad payload",
			method: "POST",
			body:   "foo",
			path:   "/jobs",
			code:   http.StatusBadRequest,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "add job - invalid id",
			method: "POST",
			body: `{
				"id": "foo",
				"script": "all_at_once",
				"gauges":  {"g001": {}, "g002": {}},
				"cron": "* * * * *",
				"options": {"gauges": 3}
			}`,
			path: "/jobs",
			code: http.StatusBadRequest,
			resp: `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "add job - no codes",
			method: "POST",
			body: `{
				"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
				"script": "all_at_once",
				"gauges": {},
				"cron": "* * * * *",
				"options": {"gauges": 3}
			}`,
			path: "/jobs",
			code: http.StatusInternalServerError,
			resp: `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "add job - bad cron",
			method: "POST",
			body: `{
				"id": "24e45a47-7ae2-453a-afa3-153392e2460b",
				"script": "all_at_once",
				"gauges":  {"g001": {}, "g002": {}},
				"cron": "a * * * *",
				"options": {"gauges": 3}
			}`,
			path: "/jobs",
			code: http.StatusInternalServerError,
			resp: `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "add job - duplicate",
			method: "POST",
			body: `{
				"id": "48f979ec-268b-11ea-978f-2e728ce88125",
				"script": "all_at_once",
				"gauges":  {"g001": {}, "g002": {}},
				"cron": "a * * * *",
				"options": {"gauges": 3}
			}`,
			path: "/jobs",
			code: http.StatusInternalServerError,
			resp: `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "delete job - not found",
			method: "DELETE",
			path:   "/jobs/24e45a47-7ae2-453a-afa3-153392e2460b",
			code:   http.StatusInternalServerError,
			resp:   `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name:   "delete job - success",
			method: "DELETE",
			path:   "/jobs/48f979ec-268b-11ea-978f-2e728ce88125",
			resp:   `{ "success": true }`,
		},
		{
			name: "measurements/script success",
			path: "/measurements/broken",
			resp: `[{"script": "broken", "code": "g000", "timestamp": "<<PRESENCE>>", "flow": -100, "level": -100}]`,
		},
		{
			name: "measurements/gauge success",
			path: "/measurements/broken/g000",
			resp: `[{"script": "broken", "code": "g000", "timestamp": "<<PRESENCE>>", "flow": -100, "level": -100}]`,
		},
		{
			name: "measurements singl gauge latest success",
			path: "/measurements/broken/g000/latest",
			resp: `[{"script": "broken", "code": "g000", "timestamp": "<<PRESENCE>>", "flow": -100, "level": -100}]`,
		},
		{
			name: "measurements/latest success",
			path: "/measurements/latest?scripts=broken,all_at_once",
			resp: `[{"script": "broken", "code": "g000", "timestamp": "<<PRESENCE>>", "flow": -100, "level": -100}]`,
		},
		{
			name: "measurements - bad query",
			path: "/measurements/broken?from=foo&to=bar",
			code: http.StatusBadRequest,
			resp: `{ "error": "<<PRESENCE>>", "status": "<<PRESENCE>>", "request_id": "<<PRESENCE>>" }`,
		},
		{
			name: "measurements/nearest success",
			path: fmt.Sprintf("/measurements/broken/g000/nearest?to=%d", time.Now().Add(-15*time.Minute).UTC().Unix()),
			resp: `{"script": "broken", "code": "g000", "timestamp": "<<PRESENCE>>", "flow": -100, "level": -100}`,
		},
		{
			name: "measurements/nearest fail",
			path: fmt.Sprintf("/measurements/broken/g000/nearest?to=%d", time.Now().Add(333*time.Minute).UTC().Unix()),
			resp: `null`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exCode := http.StatusOK
			if tt.code != 0 {
				exCode = tt.code
			}
			srv := prepareServer()
			resp, code := runCase(t, srv, tt)
			respErr := ""
			if tt.code == 200 && code != 200 {
				respErr = resp
			}
			assert.Equal(t, exCode, code, "response code for %s is wrong %s", tt.name, respErr)
			if tt.resp != "" {
				ja := jsonassert.New(t)
				ja.Assertf(resp, tt.resp)
			}
			srv.shutdown()
		})
	}
}
