package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/philoveracity/pvdifyd/internal/models"
)

// handleListReleases returns all releases for an app
func (s *Server) handleListReleases(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	releases, err := s.db.ListReleases(name, limit)
	if err != nil {
		s.logger.Error("list releases", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to list releases")
		return
	}
	if releases == nil {
		releases = []*models.Release{}
	}
	s.json(w, http.StatusOK, releases)
}

// handleCreateRelease creates a new release (deploy)
func (s *Server) handleCreateRelease(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	var req models.CreateReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Image == "" {
		s.error(w, http.StatusBadRequest, "image is required")
		return
	}

	// Get current config version
	var configVersion int
	cfg, _ := s.db.GetLatestConfig(name)
	if cfg != nil {
		configVersion = cfg.Version
	}

	release := &models.Release{
		AppName:       name,
		Image:         req.Image,
		ConfigVersion: configVersion,
		Status:        models.ReleaseStatusPending,
		CreatedBy:     req.CreatedBy,
	}

	if err := s.db.CreateRelease(release); err != nil {
		s.logger.Error("create release", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to create release")
		return
	}

	// Update app with new image
	if err := s.db.UpdateApp(name, &req.Image, nil, nil); err != nil {
		s.logger.Error("update app image", "error", err)
	}

	// TODO: Actually deploy the release
	// 1. Pull image
	// 2. Generate/update systemd unit
	// 3. Start new container
	// 4. Health check
	// 5. Update tunnel config
	// 6. Stop old container
	// 7. Update release status

	s.logger.Info("release created", "app", name, "version", release.Version, "image", req.Image)
	s.json(w, http.StatusCreated, release)
}

// handleGetRelease returns a specific release
func (s *Server) handleGetRelease(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	versionStr := chi.URLParam(r, "version")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		s.error(w, http.StatusBadRequest, "invalid version")
		return
	}

	release, err := s.db.GetRelease(name, version)
	if err != nil {
		s.logger.Error("get release", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to get release")
		return
	}
	if release == nil {
		s.error(w, http.StatusNotFound, "release not found")
		return
	}

	s.json(w, http.StatusOK, release)
}

// handleRollback rolls back to a previous release
func (s *Server) handleRollback(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	var req models.RollbackRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Get target release
	var target *models.Release
	if req.Version > 0 {
		target, err = s.db.GetRelease(name, req.Version)
	} else {
		// Get previous release (latest - 1)
		releases, _ := s.db.ListReleases(name, 2)
		if len(releases) >= 2 {
			target = releases[1]
		}
	}

	if target == nil {
		s.error(w, http.StatusBadRequest, "no release to rollback to")
		return
	}

	// Create new release with old image
	release := &models.Release{
		AppName:       name,
		Image:         target.Image,
		ConfigVersion: target.ConfigVersion,
		Status:        models.ReleaseStatusPending,
		CreatedBy:     "rollback",
	}

	if err := s.db.CreateRelease(release); err != nil {
		s.logger.Error("create rollback release", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to create rollback")
		return
	}

	// TODO: Actually deploy the rollback

	s.logger.Info("rollback initiated", "app", name, "to_version", target.Version, "new_version", release.Version)
	s.json(w, http.StatusCreated, release)
}
