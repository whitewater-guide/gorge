package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/whitewater-guide/gorge/config"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/schedule"
	"github.com/whitewater-guide/gorge/storage"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func seedHealthNotifierTests(db storage.DatabaseManager, cache storage.CacheManager) {
	// Healthy job
	// nolint:errcheck
	db.AddJob(core.JobDescription{
		ID:     "48f979ec-268b-11ea-978f-2e728ce88125",
		Script: "all_at_once",
		Gauges: map[string]json.RawMessage{"a000": json.RawMessage("{}")},
		Cron:   "0 0 * * *",
	}, func(job core.JobDescription) error {
		return nil
	})
	cache.SaveStatus("48f979ec-268b-11ea-978f-2e728ce88125", "g000", nil, 10) // nolint:errcheck

	// Unhealthy job
	// nolint:errcheck
	db.AddJob(core.JobDescription{
		ID:     "e0b198ad-d7cd-4d2b-aeb0-ad83992bc851",
		Script: "one_by_one",
		Gauges: map[string]json.RawMessage{"o000": json.RawMessage("{}")},
	}, func(job core.JobDescription) error {
		return nil
	})
	cache.SaveStatus("e0b198ad-d7cd-4d2b-aeb0-ad83992bc851", "", errors.New("test error"), 0) // nolint:errcheck
}

func TestHealthNotifier(t *testing.T) {
	// start test server that is used to listen to notifier calls
	// notifier will trigger immediate because we inject immediate cron
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		ja := jsonassert.New(t)
		ja.Assertf(string(data), `[{"id":"e0b198ad-d7cd-4d2b-aeb0-ad83992bc851","script":"one_by_one","lastRun":"<<PRESENCE>>"}]`)

		if h := r.Header["Authorization"][0]; h != "Bearer __token__" {
			t.Errorf("invalid authorization header: %s", h)
		}
		if h := r.Header["Foo"][0]; h != "Bar" {
			t.Errorf("invalid foo header: %s", h)
		}
	}))
	defer srv.Close()

	cfg := config.TestConfig()
	cfg.Hooks.Health.URL = srv.URL
	cfg.Hooks.Health.Headers = []string{"Authorization: Bearer __token__", "Foo: Bar"}

	app := fx.New(
		fx.Invoke(func(lc fx.Lifecycle, db storage.DatabaseManager, cache storage.CacheManager) {
			lc.Append(fx.Hook{
				OnStart: func(c context.Context) error {
					seedHealthNotifierTests(db, cache)
					return nil
				},
			})
		}),
		fx.Options(
			fx.Supply(cfg),
			fx.Provide(testLogger),
			schedule.TestModule,
			storage.Module,
			fx.Invoke(startHealthNotifier),
		),
		fx.WithLogger(
			func() fxevent.Logger {
				return fxevent.NopLogger
			},
		),
	)
	if err := app.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer app.Stop(context.Background()) //nolint:errcheck
}
