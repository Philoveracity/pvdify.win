package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/philoveracity/pvdifyd/internal/models"
)

// handleListApps returns all apps
func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	apps, err := s.db.ListApps()
	if err != nil {
		s.logger.Error("list apps", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to list apps")
		return
	}
	if apps == nil {
		apps = []*models.App{}
	}
	s.json(w, http.StatusOK, apps)
}

// handleCreateApp creates a new app
func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		s.error(w, http.StatusBadRequest, "name is required")
		return
	}

	// Check if app exists
	existing, err := s.db.GetApp(req.Name)
	if err != nil {
		s.logger.Error("check app exists", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to check app")
		return
	}
	if existing != nil {
		s.error(w, http.StatusConflict, "app already exists")
		return
	}

	// Allocate port
	port, err := s.db.AllocatePort(s.cfg.Ports.Start, s.cfg.Ports.End)
	if err != nil {
		s.logger.Error("allocate port", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to allocate port")
		return
	}

	app := &models.App{
		Name:        req.Name,
		Environment: req.Environment,
		Status:      models.AppStatusCreated,
		BindPort:    port,
	}
	if app.Environment == "" {
		app.Environment = "production"
	}

	if err := s.db.CreateApp(app); err != nil {
		s.logger.Error("create app", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to create app")
		return
	}

	// Create default web process
	process := &models.Process{
		AppName: app.Name,
		Name:    "web",
		Count:   1,
	}
	if err := s.db.UpsertProcess(process); err != nil {
		s.logger.Error("create default process", "error", err)
	}

	s.logger.Info("app created", "name", app.Name, "port", app.BindPort)
	s.json(w, http.StatusCreated, app)
}

// handleGetApp returns a single app
func (s *Server) handleGetApp(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil {
		s.logger.Error("get app", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to get app")
		return
	}
	if app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	// Enrich with processes and domains
	processes, _ := s.db.ListProcesses(name)
	domains, _ := s.db.ListDomains(name)
	activeRelease, _ := s.db.GetActiveRelease(name)

	response := map[string]interface{}{
		"app":       app,
		"processes": processes,
		"domains":   domains,
	}
	if activeRelease != nil {
		response["current_release"] = activeRelease
	}

	s.json(w, http.StatusOK, response)
}

// handleUpdateApp updates an existing app
func (s *Server) handleUpdateApp(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil {
		s.logger.Error("get app", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to get app")
		return
	}
	if app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	var req models.UpdateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.db.UpdateApp(name, req.Image, nil, nil); err != nil {
		s.logger.Error("update app", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to update app")
		return
	}

	app, _ = s.db.GetApp(name)
	s.json(w, http.StatusOK, app)
}

// handleDeleteApp removes an app
func (s *Server) handleDeleteApp(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil {
		s.logger.Error("get app", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to get app")
		return
	}
	if app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	// TODO: Stop running containers
	// TODO: Remove systemd units
	// TODO: Update tunnel config

	if err := s.db.DeleteApp(name); err != nil {
		s.logger.Error("delete app", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to delete app")
		return
	}

	s.logger.Info("app deleted", "name", name)
	w.WriteHeader(http.StatusNoContent)
}
