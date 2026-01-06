package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents daemon configuration
type Config struct {
	Listen    string       `yaml:"listen"`
	StateDir  string       `yaml:"state_dir"`
	Database  string       `yaml:"database"`
	StaticDir string       `yaml:"static_dir"` // Directory for Admin UI static files
	Dev       bool         `yaml:"dev"`
	Log       LogConfig    `yaml:"log"`
	TLS       TLSConfig    `yaml:"tls"`
	Auth      AuthConfig   `yaml:"auth"`
	Podman    PodmanConfig `yaml:"podman"`
	Ports     PortConfig   `yaml:"ports"`
	Tunnel    TunnelConfig `yaml:"tunnel"`
	SOPS      SOPSConfig   `yaml:"sops"`
}

// LogConfig for logging settings
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// TLSConfig for HTTPS
type TLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

// AuthConfig for authentication
type AuthConfig struct {
	CFAccessTeam string `yaml:"cf_access_team"`
	CFAccessAUD  string `yaml:"cf_access_aud"`
}

// PodmanConfig for container runtime
type PodmanConfig struct {
	Socket string `yaml:"socket"`
}

// PortConfig for port allocation
type PortConfig struct {
	Start int `yaml:"start"`
	End   int `yaml:"end"`
}

// TunnelConfig for Cloudflare tunnel
type TunnelConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Config      string `yaml:"config"`
	Credentials string `yaml:"credentials"`
}

// SOPSConfig for secrets encryption
type SOPSConfig struct {
	AgeKey string `yaml:"age_key"`
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		Listen:    "0.0.0.0:9443",
		StateDir:  "/var/lib/pvdify",
		Database:  "/var/lib/pvdify/pvdifyd.db",
		StaticDir: "/opt/pvdify/admin-ui/dist",
		Dev:       false,
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		TLS: TLSConfig{
			Enabled: true,
			Cert:    "/etc/pvdify/tls/cert.pem",
			Key:     "/etc/pvdify/tls/key.pem",
		},
		Podman: PodmanConfig{
			Socket: "unix:///run/user/1000/podman/podman.sock",
		},
		Ports: PortConfig{
			Start: 3000,
			End:   3999,
		},
		Tunnel: TunnelConfig{
			Enabled: true,
			Config:  "/var/lib/pvdify/tunnels/pvdify-apps.yml",
		},
	}
}

// Load reads configuration from file and environment
func Load(path string) (*Config, error) {
	cfg := Default()

	// Load from file if exists
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config: %w", err)
		}
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse config: %w", err)
			}
		}
	}

	// Override with environment variables
	if v := os.Getenv("PVDIFY_LISTEN"); v != "" {
		cfg.Listen = v
	}
	if v := os.Getenv("PVDIFY_STATE_DIR"); v != "" {
		cfg.StateDir = v
	}
	if v := os.Getenv("PVDIFY_DB"); v != "" {
		cfg.Database = v
	}
	if v := os.Getenv("PVDIFY_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if os.Getenv("PVDIFY_DEV") == "true" {
		cfg.Dev = true
	}

	return cfg, nil
}
