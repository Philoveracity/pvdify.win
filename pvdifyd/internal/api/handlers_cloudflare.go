package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/go-chi/chi/v5"
)

// CloudflareZone represents a Cloudflare zone
type CloudflareZone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Plan   string `json:"plan"`
	Status string `json:"status"`
}

// CloudflareDNSRequest represents a request to create a DNS record
type CloudflareDNSRequest struct {
	ZoneID   string `json:"zone_id"`
	ZoneName string `json:"zone_name"`
	Type     string `json:"type"`      // CNAME, A, etc.
	Name     string `json:"name"`      // subdomain or @ for root
	Content  string `json:"content"`   // target
	Proxied  bool   `json:"proxied"`
}

// CloudflareDNSRecord represents a DNS record
type CloudflareDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

// handleListCloudflareZones returns all Cloudflare zones
func (s *Server) handleListCloudflareZones(w http.ResponseWriter, r *http.Request) {
	// Execute cf zone list command
	cmd := exec.Command("cf", "zone", "list")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		s.logger.Error("cloudflare zone list failed", "error", err, "stderr", stderr.String())
		s.error(w, http.StatusInternalServerError, "failed to list Cloudflare zones")
		return
	}

	// Parse the output (tab-separated)
	zones := []CloudflareZone{}
	lines := strings.Split(stdout.String(), "\n")
	for i, line := range lines {
		// Skip header line and empty lines
		if i == 0 || i == 1 || strings.TrimSpace(line) == "" {
			continue
		}

		// Parse tab-separated values
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			zones = append(zones, CloudflareZone{
				ID:     strings.TrimSpace(parts[0]),
				Name:   strings.TrimSpace(parts[1]),
				Plan:   strings.TrimSpace(parts[2]),
				Status: strings.TrimSpace(parts[3]),
			})
		}
	}

	s.json(w, http.StatusOK, map[string]interface{}{
		"zones": zones,
	})
}

// handleCreateCloudflareDNS creates a DNS record in Cloudflare
func (s *Server) handleCreateCloudflareDNS(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "name")
	domainName := chi.URLParam(r, "domain")

	// Verify app exists
	app, err := s.db.GetApp(appName)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	var req CloudflareDNSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ZoneName == "" {
		s.error(w, http.StatusBadRequest, "zone_name is required")
		return
	}

	// Default values
	if req.Type == "" {
		req.Type = "CNAME"
	}
	if req.Content == "" {
		// Point to the app's default domain
		req.Content = appName + ".pvdify.win"
	}

	// Determine record name based on domain
	recordName := domainName
	if strings.HasSuffix(domainName, "."+req.ZoneName) {
		// Extract subdomain (e.g., "app.example.com" -> "app")
		recordName = strings.TrimSuffix(domainName, "."+req.ZoneName)
	} else if domainName == req.ZoneName {
		// Root domain
		recordName = "@"
	}

	// Build cf dns create command
	args := []string{
		"dns", "create",
		"--zone", req.ZoneName,
		"--type", req.Type,
		"--name", recordName,
		"--content", req.Content,
	}
	if req.Proxied {
		args = append(args, "--proxy")
	}

	s.logger.Info("creating DNS record", "args", args)

	cmd := exec.Command("cf", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		s.logger.Error("cloudflare dns create failed", "error", err, "stderr", stderr.String(), "stdout", stdout.String())
		// Check for common errors
		errMsg := stderr.String() + stdout.String()
		if strings.Contains(errMsg, "already exists") {
			s.error(w, http.StatusConflict, "DNS record already exists")
			return
		}
		s.error(w, http.StatusInternalServerError, "failed to create DNS record: "+errMsg)
		return
	}

	s.logger.Info("DNS record created", "domain", domainName, "zone", req.ZoneName, "output", stdout.String())

	s.json(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "DNS record created successfully",
		"domain":  domainName,
		"zone":    req.ZoneName,
		"type":    req.Type,
		"content": req.Content,
		"proxied": req.Proxied,
	})
}

// handleListCloudflareDNS lists DNS records for a zone
func (s *Server) handleListCloudflareDNS(w http.ResponseWriter, r *http.Request) {
	zoneName := r.URL.Query().Get("zone")
	if zoneName == "" {
		s.error(w, http.StatusBadRequest, "zone query parameter is required")
		return
	}

	cmd := exec.Command("cf", "dns", "list", "--zone", zoneName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		s.logger.Error("cloudflare dns list failed", "error", err, "stderr", stderr.String())
		s.error(w, http.StatusInternalServerError, "failed to list DNS records")
		return
	}

	// Parse the output
	records := []CloudflareDNSRecord{}
	lines := strings.Split(stdout.String(), "\n")
	for i, line := range lines {
		// Skip header lines and empty lines
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			proxied := strings.TrimSpace(parts[4]) == "true"
			records = append(records, CloudflareDNSRecord{
				ID:      strings.TrimSpace(parts[0]),
				Type:    strings.TrimSpace(parts[1]),
				Name:    strings.TrimSpace(parts[2]),
				Content: strings.TrimSpace(parts[3]),
				Proxied: proxied,
			})
		}
	}

	s.json(w, http.StatusOK, map[string]interface{}{
		"records": records,
	})
}
