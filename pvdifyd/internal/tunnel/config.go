package tunnel

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents a Cloudflare Tunnel configuration
type Config struct {
	Tunnel          string        `yaml:"tunnel"`
	CredentialsFile string        `yaml:"credentials-file"`
	Ingress         []IngressRule `yaml:"ingress"`
}

// IngressRule represents a tunnel ingress rule
type IngressRule struct {
	Hostname string `yaml:"hostname,omitempty"`
	Service  string `yaml:"service"`
}

// Manager manages tunnel configuration
type Manager struct {
	configPath      string
	credentialsFile string
	tunnelID        string
}

// NewManager creates a new tunnel manager
func NewManager(configPath, credentialsFile string) (*Manager, error) {
	// Try to extract tunnel ID from existing config
	var tunnelID string
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg Config
		if yaml.Unmarshal(data, &cfg) == nil {
			tunnelID = cfg.Tunnel
		}
	}

	return &Manager{
		configPath:      configPath,
		credentialsFile: credentialsFile,
		tunnelID:        tunnelID,
	}, nil
}

// SetTunnelID sets the tunnel ID
func (m *Manager) SetTunnelID(id string) {
	m.tunnelID = id
}

// Load reads the current tunnel configuration
func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Tunnel:          m.tunnelID,
				CredentialsFile: m.credentialsFile,
				Ingress:         []IngressRule{{Service: "http_status:404"}},
			}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the tunnel configuration
func (m *Manager) Save(cfg *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// AddRoute adds a route to the tunnel configuration
func (m *Manager) AddRoute(hostname string, port int) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	// Check if route already exists
	for i, rule := range cfg.Ingress {
		if rule.Hostname == hostname {
			// Update existing rule
			cfg.Ingress[i].Service = fmt.Sprintf("http://localhost:%d", port)
			return m.Save(cfg)
		}
	}

	// Insert before catch-all (last rule)
	newRule := IngressRule{
		Hostname: hostname,
		Service:  fmt.Sprintf("http://localhost:%d", port),
	}

	if len(cfg.Ingress) > 0 {
		// Insert before last rule (catch-all)
		cfg.Ingress = append(cfg.Ingress[:len(cfg.Ingress)-1],
			append([]IngressRule{newRule}, cfg.Ingress[len(cfg.Ingress)-1])...)
	} else {
		cfg.Ingress = append([]IngressRule{newRule}, IngressRule{Service: "http_status:404"})
	}

	return m.Save(cfg)
}

// RemoveRoute removes a route from the tunnel configuration
func (m *Manager) RemoveRoute(hostname string) error {
	cfg, err := m.Load()
	if err != nil {
		return err
	}

	var filtered []IngressRule
	for _, rule := range cfg.Ingress {
		if rule.Hostname != hostname {
			filtered = append(filtered, rule)
		}
	}

	cfg.Ingress = filtered
	return m.Save(cfg)
}

// ListRoutes returns all configured routes
func (m *Manager) ListRoutes() ([]IngressRule, error) {
	cfg, err := m.Load()
	if err != nil {
		return nil, err
	}

	// Filter out catch-all
	var routes []IngressRule
	for _, rule := range cfg.Ingress {
		if rule.Hostname != "" {
			routes = append(routes, rule)
		}
	}

	return routes, nil
}
