package main

import (
	"net/http"

	"github.com/go-chi/render"
)

func (s *Server) handleListScripts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, s.registry.List())
	}
}
