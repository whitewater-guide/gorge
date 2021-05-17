package main

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/whitewater-guide/gorge/core"
)

func (s *server) handleUpstreamGauges() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "script")

		script, _, err := s.registry.CreateFromReader(name, r.Body)

		if err == core.ErrScriptNotFound {
			s.renderError(w, r, core.WrapErr(err, "script not found").With("script", name), "script not found", http.StatusNotFound)
			return
		}
		if err != nil {
			s.renderError(w, r, core.NewErr(err).With("script", name), "failed to create script", http.StatusInternalServerError)
			return
		}

		result, err := script.ListGauges()
		sort.Sort(result)
		if err != nil {
			s.renderError(w, r, err, "failed to list gauges", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, result)
	}
}

func (s *server) handleUpstreamMeasurements() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errorMsg := "failed to harvest measurements from upstream"
		name := chi.URLParam(r, "script")
		logger := s.logger.WithField("script", name)
		codes := core.StringSet{}
		for _, s := range strings.Split(r.URL.Query().Get("codes"), ",") {
			if s != "" {
				codes[s] = struct{}{}
			}
		}
		logger.WithField("codes", codes).Debug("harvesting measurements")
		sinceS := r.URL.Query().Get("since")
		var since int64
		var err error

		if sinceS != "" {
			since, err = strconv.ParseInt(sinceS, 10, 64)
			if err != nil {
				s.renderError(w, r, core.WrapErr(err, "failed to parse since query parameter").With("since", sinceS), errorMsg, http.StatusBadRequest)
				return
			}
		}

		script, mode, err := s.registry.CreateFromReader(name, r.Body)
		if err == core.ErrScriptNotFound {
			s.renderError(w, r, core.WrapErr(err, "script not found").With("script", name), "script not found", http.StatusNotFound)
			return
		}
		if err != nil {
			s.renderError(w, r, core.NewErr(err).With("script", name), "failed to create script", http.StatusInternalServerError)
			return
		}
		if mode == core.OneByOne && len(codes) != 1 {
			s.renderError(w, r, errors.New("exactly one code must be provided for one-by-one scripts"), errorMsg, http.StatusMethodNotAllowed)
			return
		}
		script.SetLogger(logger)

		in := make(chan *core.Measurement)
		errCh := make(chan error, 1)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		var out <-chan *core.Measurement = in

		go script.Harvest(ctx, in, errCh, codes, since)
		if len(codes) >= 1 {
			out = core.FilterMeasurements(ctx, in, logger, core.CodesFilter{Codes: codes})
		}
		resultCh := core.SinkToSlice(context.Background(), out)
		var result core.Measurements = <-resultCh
		err = <-errCh
		sort.Sort(result)

		if err != nil {
			s.renderError(w, r, err, errorMsg, http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, result)
	}

}
