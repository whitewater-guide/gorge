package main

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/whitewater-guide/gorge/version"
)

func (s *server) handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		render.JSON(w, r, map[string]string{"version": version.Version})
	}
}
