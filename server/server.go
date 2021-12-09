package main

import (
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/config"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/storage"
	"go.uber.org/fx"
)

type ServerParams struct {
	fx.In

	Logger    *logrus.Logger
	Db        storage.DatabaseManager
	Cache     storage.CacheManager
	Cfg       *config.Config
	Registry  *core.ScriptRegistry
	Scheduler core.JobScheduler
}

type Server struct {
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

func (s *Server) routes() {
	s.logger.Debug("creating routes")
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
	s.logger.Debug("created routes")
}

func newServer(lc fx.Lifecycle, p ServerParams) *Server {
	result := &Server{
		endpoint:  p.Cfg.Endpoint,
		port:      p.Cfg.Port,
		registry:  p.Registry,
		debug:     p.Cfg.Debug,
		cache:     p.Cache,
		database:  p.Db,
		scheduler: p.Scheduler,
		logger:    p.Logger,
	}

	core.Client = core.NewClient(p.Cfg.HTTP, result.logger.WithField("client", "http"))

	return result
}
