package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/philoveracity/pvdifyd/internal/models"
)

// handleListProcesses returns all processes for an app
func (s *Server) handleListProcesses(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	processes, err := s.db.ListProcesses(name)
	if err != nil {
		s.logger.Error("list processes", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to list processes")
		return
	}

	// TODO: Get actual runtime status from Podman/systemd
	var statuses []models.ProcessStatus
	for _, p := range processes {
		for i := 1; i <= p.Count; i++ {
			statuses = append(statuses, models.ProcessStatus{
				Name:     p.Name,
				Instance: i,
				State:    "unknown", // TODO: Query actual state
			})
		}
	}

	s.json(w, http.StatusOK, map[string]interface{}{
		"definitions": processes,
		"instances":   statuses,
	})
}

// handleScale scales processes
func (s *Server) handleScale(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	var req models.ScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Processes) == 0 {
		s.error(w, http.StatusBadRequest, "processes is required")
		return
	}

	for procName, count := range req.Processes {
		if count < 0 {
			s.error(w, http.StatusBadRequest, "count must be non-negative")
			return
		}

		if err := s.db.ScaleProcess(name, procName, count); err != nil {
			s.logger.Error("scale process", "error", err, "process", procName)
			s.error(w, http.StatusInternalServerError, "failed to scale process")
			return
		}

		// TODO: Actually start/stop systemd units

		s.logger.Info("process scaled", "app", name, "process", procName, "count", count)
	}

	// Return updated process list
	processes, _ := s.db.ListProcesses(name)
	s.json(w, http.StatusOK, processes)
}

// handleRestart restarts all processes
func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	processes, err := s.db.ListProcesses(name)
	if err != nil {
		s.logger.Error("list processes", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to list processes")
		return
	}

	// TODO: Actually restart systemd units
	for _, p := range processes {
		s.logger.Info("restarting process", "app", name, "process", p.Name, "count", p.Count)
	}

	s.json(w, http.StatusOK, map[string]string{
		"status": "restarting",
	})
}
