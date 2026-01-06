package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents the Pvdify API client
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New creates a new API client
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// App represents an application
type App struct {
	Name        string            `json:"name"`
	Environment string            `json:"environment,omitempty"`
	Status      string            `json:"status"`
	Image       string            `json:"image,omitempty"`
	BindPort    int               `json:"bind_port,omitempty"`
	Domains     []string          `json:"domains,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Config      map[string]string `json:"config,omitempty"`
}

// Release represents a deployment release
type Release struct {
	Version       int       `json:"version"`
	Image         string    `json:"image"`
	ConfigVersion int       `json:"config_version"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     string    `json:"created_by,omitempty"`
}

// Process represents a running process
type Process struct {
	Type    string `json:"type"`
	Command string `json:"command,omitempty"`
	Count   int    `json:"count"`
	Status  string `json:"status"`
}

// CreateAppRequest represents the request to create an app
type CreateAppRequest struct {
	Name        string `json:"name"`
	Environment string `json:"environment,omitempty"`
}

// CreateReleaseRequest represents a deploy request
type CreateReleaseRequest struct {
	Image string `json:"image"`
}

// ScaleRequest represents a scale request
type ScaleRequest struct {
	Processes map[string]int `json:"processes"`
}

// Error response from API
type APIError struct {
	Error string `json:"error"`
}

// do performs an HTTP request
func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// parseResponse parses the response body into the target
func parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Error != "" {
			return fmt.Errorf("API error: %s", apiErr.Error)
		}
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if target != nil {
		return json.NewDecoder(resp.Body).Decode(target)
	}
	return nil
}

// ListApps returns all apps
func (c *Client) ListApps() ([]App, error) {
	resp, err := c.do("GET", "/api/v1/apps", nil)
	if err != nil {
		return nil, err
	}

	var apps []App
	if err := parseResponse(resp, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

// GetApp returns a single app
func (c *Client) GetApp(name string) (*App, error) {
	resp, err := c.do("GET", "/api/v1/apps/"+name, nil)
	if err != nil {
		return nil, err
	}

	var app App
	if err := parseResponse(resp, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// CreateApp creates a new app
func (c *Client) CreateApp(name, environment string) (*App, error) {
	req := CreateAppRequest{Name: name, Environment: environment}
	resp, err := c.do("POST", "/api/v1/apps", req)
	if err != nil {
		return nil, err
	}

	var app App
	if err := parseResponse(resp, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// DeleteApp deletes an app
func (c *Client) DeleteApp(name string) error {
	resp, err := c.do("DELETE", "/api/v1/apps/"+name, nil)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// CreateRelease deploys a new image
func (c *Client) CreateRelease(appName, image string) (*Release, error) {
	req := CreateReleaseRequest{Image: image}
	resp, err := c.do("POST", "/api/v1/apps/"+appName+"/releases", req)
	if err != nil {
		return nil, err
	}

	var release Release
	if err := parseResponse(resp, &release); err != nil {
		return nil, err
	}
	return &release, nil
}

// ListReleases returns all releases for an app
func (c *Client) ListReleases(appName string) ([]Release, error) {
	resp, err := c.do("GET", "/api/v1/apps/"+appName+"/releases", nil)
	if err != nil {
		return nil, err
	}

	var releases []Release
	if err := parseResponse(resp, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

// Rollback rolls back to the previous release
func (c *Client) Rollback(appName string) (*Release, error) {
	resp, err := c.do("POST", "/api/v1/apps/"+appName+"/rollback", nil)
	if err != nil {
		return nil, err
	}

	var release Release
	if err := parseResponse(resp, &release); err != nil {
		return nil, err
	}
	return &release, nil
}

// GetConfig returns app configuration
func (c *Client) GetConfig(appName string) (map[string]string, error) {
	resp, err := c.do("GET", "/api/v1/apps/"+appName+"/config", nil)
	if err != nil {
		return nil, err
	}

	var config map[string]string
	if err := parseResponse(resp, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// SetConfig sets configuration variables
func (c *Client) SetConfig(appName string, config map[string]string) error {
	resp, err := c.do("PUT", "/api/v1/apps/"+appName+"/config", config)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// UnsetConfig removes a configuration variable
func (c *Client) UnsetConfig(appName, key string) error {
	resp, err := c.do("DELETE", "/api/v1/apps/"+appName+"/config/"+key, nil)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// ListDomains returns all domains for an app
func (c *Client) ListDomains(appName string) ([]string, error) {
	resp, err := c.do("GET", "/api/v1/apps/"+appName+"/domains", nil)
	if err != nil {
		return nil, err
	}

	var domains []string
	if err := parseResponse(resp, &domains); err != nil {
		return nil, err
	}
	return domains, nil
}

// AddDomain adds a domain to an app
func (c *Client) AddDomain(appName, domain string) error {
	resp, err := c.do("POST", "/api/v1/apps/"+appName+"/domains", map[string]string{"domain": domain})
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// RemoveDomain removes a domain from an app
func (c *Client) RemoveDomain(appName, domain string) error {
	resp, err := c.do("DELETE", "/api/v1/apps/"+appName+"/domains/"+domain, nil)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// ListProcesses returns all processes for an app
func (c *Client) ListProcesses(appName string) ([]Process, error) {
	resp, err := c.do("GET", "/api/v1/apps/"+appName+"/ps", nil)
	if err != nil {
		return nil, err
	}

	var processes []Process
	if err := parseResponse(resp, &processes); err != nil {
		return nil, err
	}
	return processes, nil
}

// Scale scales processes
func (c *Client) Scale(appName string, processes map[string]int) error {
	req := ScaleRequest{Processes: processes}
	resp, err := c.do("POST", "/api/v1/apps/"+appName+"/ps/scale", req)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// Restart restarts all processes
func (c *Client) Restart(appName string) error {
	resp, err := c.do("POST", "/api/v1/apps/"+appName+"/ps/restart", nil)
	if err != nil {
		return err
	}
	return parseResponse(resp, nil)
}

// GetLogs returns logs for an app (non-streaming)
func (c *Client) GetLogs(appName string, lines int, follow bool) (io.ReadCloser, error) {
	path := fmt.Sprintf("/api/v1/apps/%s/logs?lines=%d&follow=%t", appName, lines, follow)
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Error != "" {
			return nil, fmt.Errorf("API error: %s", apiErr.Error)
		}
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return resp.Body, nil
}
