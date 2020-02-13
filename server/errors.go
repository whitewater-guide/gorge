package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/whitewater-guide/gorge/core"
)

func (s *server) renderError(w http.ResponseWriter, r *http.Request, e error, msg string, status int) {
	if e == nil {
		return
	}
	reqID := middleware.GetReqID(r.Context())
	resp := core.NewErrorResponse(e, msg, status)
	resp.ReqID = reqID

	entry := s.logger.WithField("uri", r.RequestURI).WithField("request_id", reqID)

	if resp.Ctx != nil {
		entry = entry.WithFields(resp.Ctx)
	}

	script := chi.URLParam(r, "script")
	code := chi.URLParam(r, "code")

	if script != "" {
		entry = entry.WithField("script", script)
	}
	if code != "" {
		entry = entry.WithField("code", code)
	}

	entry.Error(e)

	render.Render(w, r, resp) // nolint:errcheck
}
