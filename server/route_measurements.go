package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/whitewater-guide/gorge/core"
	"github.com/whitewater-guide/gorge/storage"
)

func (s *server) handleGetMeasurements() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		script := chi.URLParam(r, "script")
		code := chi.URLParam(r, "code")
		q := r.URL.Query()
		toS := q.Get("to")
		fromS := q.Get("from")

		query, err := storage.NewMeasurementsQuery(script, code, fromS, toS)
		if err != nil {
			s.renderError(w, r, err, "failed to create measurements query", http.StatusBadRequest)
			return
		}

		measurements, err := s.database.GetMeasurements(*query)

		if err != nil {
			s.renderError(w, r, err, "failed to get measurements", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, measurements)
	}
}

func (s *server) handleGetNearest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		script := chi.URLParam(r, "script")
		code := chi.URLParam(r, "code")
		q := r.URL.Query()

		toI, err := strconv.ParseInt(q.Get("to"), 10, 64)
		if err != nil {
			s.renderError(w, r, err, "failed to get nearest measurement", http.StatusInternalServerError)
			return
		}

		measurement, err := s.database.GetNearestMeasurement(script, code, time.Unix(toI, 0), time.Hour)

		if err != nil {
			s.renderError(w, r, err, "failed to get nearest measurement", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, measurement)
	}
}

func (s *server) handleGetLatest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		script := chi.URLParam(r, "script")
		code := chi.URLParam(r, "code")
		scripts := r.URL.Query().Get("scripts")
		q := map[string]core.StringSet{}
		if script != "" && code != "" {
			q[script] = core.StringSet{code: {}}
		} else {
			for _, s := range strings.Split(scripts, ",") {
				if s != "" {
					q[s] = core.StringSet{}
				}
			}
		}
		measurements, err := s.cache.LoadLatestMeasurements(q)
		res := make([]core.Measurement, len(measurements))
		i := 0
		for _, m := range measurements {
			res[i] = m
			i++
		}
		if err != nil {
			s.renderError(w, r, err, "failed to get latest measurements", http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, res)
	}
}
