package main

import (
	"fmt"
	"io/ioutil"
	"net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/schedule"
	"github.com/whitewater-guide/gorge/storage"
)

type server struct {
	port      string
	endpoint  string
	cache     storage.CacheManager
	database  storage.DatabaseManager
	logger    *logrus.Logger
	registry  *core.ScriptRegistry
	router    *chi.Mux
	scheduler core.JobScheduler
	debug     bool
}

func (s *server) routes() {
	s.router = chi.NewRouter()
	s.router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.RequestID,
		middleware.RedirectSlashes,
		middleware.Compress(5),
		middleware.NoCache,
		middleware.Recoverer,
		middleware.Heartbeat("/healthcheck"),
	)
	s.router.Route(s.endpoint, func(r chi.Router) {
		r.Get("/version", s.handleVersion())
		r.Get("/scripts", s.handleListScripts())

		r.Post("/upstream/{script}/gauges", s.handleUpstreamGauges())
		r.Post("/upstream/{script}/measurements", s.handleUpstreamMeasurements())

		r.Get("/jobs", s.handleListJobs())
		r.Get("/jobs/{jobId}", s.handleGetJob())
		r.Get("/jobs/{jobId}/gauges", s.handleGetJobGauges())
		r.Post("/jobs", s.handleAddJob())
		r.Delete("/jobs/{jobId}", s.handleDeleteJob())

		r.Get("/measurements/{script}", s.handleGetMeasurements())
		r.Get("/measurements/{script}/{code}", s.handleGetMeasurements())
		r.Get("/measurements/{script}/{code}/latest", s.handleGetLatest())
		r.Get("/measurements/{script}/{code}/nearest", s.handleGetNearest())
		r.Get("/measurements/latest", s.handleGetLatest())
	})
	if s.debug {
		s.router.HandleFunc("/debug/pprof", pprof.Index)
		s.router.HandleFunc("/debug/cmdline", pprof.Cmdline)
		s.router.HandleFunc("/debug/profile", pprof.Profile)
		s.router.HandleFunc("/debug/symbol", pprof.Symbol)
		s.router.HandleFunc("/debug/trace", pprof.Trace)
		s.router.Handle("/debug/heap", pprof.Handler("heap"))
		s.router.Handle("/debug/goroutine", pprof.Handler("goroutine"))
	}
}

func (s *server) start() {
	s.scheduler.Start()
	// Load initial jobs
	jobs, err := s.database.ListJobs()
	if err != nil {
		s.logger.Fatalf("failed to load initial jobs: %v", err)
	}
	for _, job := range jobs {
		err := s.scheduler.AddJob(job)
		if err != nil {
			s.logger.Fatalf("failed to schedule initial jobs: %v", err)
		}
		s.logger.WithFields(logrus.Fields{"script": job.Script, "jobID": job.ID}).Info("started job")
	}
}

func (s *server) shutdown() {
	s.database.Close()
	s.cache.Close()
	s.scheduler.Stop()
}

func newServer(cfg *config, registry *core.ScriptRegistry) *server {
	result := &server{registry: registry, debug: cfg.Debug}
	result.logger = logrus.New()
	if cfg.Log.Format == "json" {
		result.logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		result.logger.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true})
	}
	logLevel := cfg.Log.Level
	if cfg.Debug {
		logLevel = "debug"
	}
	if logLevel == "" {
		result.logger.SetOutput(ioutil.Discard)
	} else {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			lvl = logrus.DebugLevel
		}
		result.logger.SetLevel(lvl)
	}
	result.logger.WithFields(logrus.Fields{
		"cache": cfg.Cache,
		"db":    cfg.Db,
	}).Info("starting storage...")

	switch cfg.Cache {
	case "redis":
		cache, err := storage.NewRedisCacheManager(cfg.Redis.Host, cfg.Redis.Port)
		if err != nil {
			result.logger.Fatal(err)
		}
		result.cache = cache
	case "inmemory":
		cache, err := storage.NewEmbeddedCacheManager()
		if err != nil {
			result.logger.Fatal(err)
		}
		result.cache = cache
	default:
		result.logger.Fatal("invalid cache manager")
	}

	switch cfg.Db {
	case "postgres":
		pgConnStr := fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			cfg.Pg.User,
			cfg.Pg.Password,
			cfg.Pg.Host,
			cfg.Pg.Db,
		)
		db, err := storage.NewPostgresManager(pgConnStr, cfg.DbChunkSize)
		if err != nil {
			result.logger.Fatal(err)
		}
		result.database = db
	case "inmemory":
		db, err := storage.NewSqliteDb(cfg.DbChunkSize)
		if err != nil {
			result.logger.Fatal(err)
		}
		result.database = db
	default:
		result.logger.Fatal("invalid database manager")
	}
	result.logger.Info("storage started")

	result.port = cfg.Port
	result.endpoint = cfg.Endpoint

	result.scheduler = &schedule.SimpleScheduler{
		Database: result.database,
		Cache:    result.cache,
		Logger:   result.logger,
		Registry: result.registry,
		Cron:     cron.New(),
	}

	core.Client = core.NewClient(cfg.HTTP)

	return result
}
