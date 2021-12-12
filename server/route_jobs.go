package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/whitewater-guide/gorge/core"
)

func (s *Server) handleAddJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var description core.JobDescription
		err := render.Bind(r, &description)
		if err != nil {
			s.renderError(w, r, err, "bad job description", http.StatusBadRequest)
			return
		}

		err = s.database.AddJob(description, s.scheduler.AddJob)
		if err != nil {
			s.renderError(w, r, err, "failed to add job", http.StatusInternalServerError)
			return
		}

		s.logger.WithField("codes", core.GaugesCodes(description.Gauges)).
			WithField("script", description.Script).
			WithField("id", description.ID).
			Info("added job")

		render.JSON(w, r, description)
	}
}

func (s *Server) handleDeleteJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := chi.URLParam(r, "jobId")
		err := s.database.DeleteJob(jobID, s.scheduler.DeleteJob)
		if err != nil {
			s.renderError(w, r, err, "failed to delete job", http.StatusInternalServerError)
			return
		}
		s.logger.WithField("id", jobID).Infof("deleted job")
		render.JSON(w, r, map[string]interface{}{"success": true})
	}
}

func (s *Server) handleGetJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := chi.URLParam(r, "jobId")
		job, err := s.database.GetJob(jobID)
		if err != nil {
			s.renderError(w, r, err, "failed to delete job", http.StatusInternalServerError)
			return
		}
		if job == nil {
			s.renderError(w, r, errors.New("not found"), "not found", http.StatusNotFound)
			return
		}
		render.JSON(w, r, *job)
	}
}

func (s *Server) handleGetJobGauges() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID := chi.URLParam(r, "jobId")
		statuses, err := s.cache.LoadGaugeStatuses(jobID)
		nexts := s.scheduler.ListNext(jobID)
		for k, v := range nexts {
			next := v
			if status, ok := statuses[k]; ok {
				status.NextRun = &next
				statuses[k] = status
			} else {
				statuses[k] = core.Status{NextRun: &next}
			}
		}
		if err != nil {
			s.renderError(w, r, err, "failed to list job gauge statuses", http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, statuses)
	}
}

func (s *Server) handleListJobs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs, err := s.database.ListJobs()
		if err != nil {
			s.renderError(w, r, err, "failed to list jobs", http.StatusInternalServerError)
			return
		}
		statuses, err := s.cache.LoadJobStatuses()
		if err != nil {
			s.renderError(w, r, err, "failed to list jobs statuses", http.StatusInternalServerError)
			return
		}
		nexts := s.scheduler.ListNext("")
		for i, job := range jobs {
			status, ok := statuses[job.ID]
			if ok {
				jobs[i].Status = &status
				if next, ok := nexts[job.ID]; ok {
					jobs[i].Status.NextRun = &next
				}
			} else if next, ok := nexts[job.ID]; ok {
				jobs[i].Status = &core.Status{NextRun: &next}

			}
		}
		render.JSON(w, r, jobs)
	}
}
