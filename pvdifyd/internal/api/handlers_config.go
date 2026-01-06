package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/philoveracity/pvdifyd/internal/models"
	"gopkg.in/yaml.v3"
)

// handleGetConfig returns current config vars
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	cfg, err := s.db.GetLatestConfig(name)
	if err != nil {
		s.logger.Error("get config", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to get config")
		return
	}

	if cfg == nil {
		s.json(w, http.StatusOK, map[string]interface{}{
			"version": 0,
			"vars":    map[string]string{},
		})
		return
	}

	// Decrypt and parse config
	// TODO: Implement SOPS decryption
	var vars models.ConfigData
	if err := yaml.Unmarshal(cfg.Data, &vars); err != nil {
		s.logger.Error("parse config", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to parse config")
		return
	}

	s.json(w, http.StatusOK, map[string]interface{}{
		"version": cfg.Version,
		"vars":    vars,
	})
}

// handleSetConfig sets config vars
func (s *Server) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	var req models.SetConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get current config and merge
	currentVars := make(models.ConfigData)
	currentCfg, _ := s.db.GetLatestConfig(name)
	if currentCfg != nil {
		yaml.Unmarshal(currentCfg.Data, &currentVars)
	}

	// Merge new vars
	for k, v := range req.Vars {
		currentVars[k] = v
	}

	// Serialize
	data, err := yaml.Marshal(currentVars)
	if err != nil {
		s.logger.Error("marshal config", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to serialize config")
		return
	}

	// TODO: Encrypt with SOPS before storing

	cfg, err := s.db.CreateConfigVersion(name, data)
	if err != nil {
		s.logger.Error("create config version", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to save config")
		return
	}

	s.logger.Info("config updated", "app", name, "version", cfg.Version)
	s.json(w, http.StatusOK, map[string]interface{}{
		"version": cfg.Version,
		"vars":    currentVars,
	})
}

// handleUnsetConfig removes a config var
func (s *Server) handleUnsetConfig(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	key := chi.URLParam(r, "key")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	// Get current config
	currentVars := make(models.ConfigData)
	currentCfg, _ := s.db.GetLatestConfig(name)
	if currentCfg != nil {
		yaml.Unmarshal(currentCfg.Data, &currentVars)
	}

	// Remove key
	if _, exists := currentVars[key]; !exists {
		s.error(w, http.StatusNotFound, "config key not found")
		return
	}
	delete(currentVars, key)

	// Serialize and save
	data, _ := yaml.Marshal(currentVars)
	cfg, err := s.db.CreateConfigVersion(name, data)
	if err != nil {
		s.logger.Error("create config version", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to save config")
		return
	}

	s.logger.Info("config key removed", "app", name, "key", key, "version", cfg.Version)
	w.WriteHeader(http.StatusNoContent)
}
