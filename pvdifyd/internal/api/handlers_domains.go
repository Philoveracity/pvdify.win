package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/philoveracity/pvdifyd/internal/models"
)

// handleListDomains returns all domains for an app
func (s *Server) handleListDomains(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	domains, err := s.db.ListDomains(name)
	if err != nil {
		s.logger.Error("list domains", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to list domains")
		return
	}
	if domains == nil {
		domains = []*models.Domain{}
	}
	s.json(w, http.StatusOK, domains)
}

// handleAddDomain adds a domain to an app
func (s *Server) handleAddDomain(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	var req models.AddDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Domain == "" {
		s.error(w, http.StatusBadRequest, "domain is required")
		return
	}

	// Check if domain already exists
	existing, _ := s.db.GetDomain(req.Domain)
	if existing != nil {
		s.error(w, http.StatusConflict, "domain already in use")
		return
	}

	domain := &models.Domain{
		Domain:  req.Domain,
		AppName: name,
		Status:  models.DomainStatusPending,
	}

	if err := s.db.CreateDomain(domain); err != nil {
		s.logger.Error("create domain", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to add domain")
		return
	}

	// TODO: Update tunnel config
	// TODO: Validate domain ownership

	s.logger.Info("domain added", "app", name, "domain", req.Domain)
	s.json(w, http.StatusCreated, domain)
}

// handleRemoveDomain removes a domain from an app
func (s *Server) handleRemoveDomain(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	domainName := chi.URLParam(r, "domain")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	domain, _ := s.db.GetDomain(domainName)
	if domain == nil || domain.AppName != name {
		s.error(w, http.StatusNotFound, "domain not found")
		return
	}

	// TODO: Update tunnel config

	if err := s.db.DeleteDomain(domainName); err != nil {
		s.logger.Error("delete domain", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to remove domain")
		return
	}

	s.logger.Info("domain removed", "app", name, "domain", domainName)
	w.WriteHeader(http.StatusNoContent)
}
