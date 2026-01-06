package systemd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// UnitConfig holds parameters for generating systemd units
type UnitConfig struct {
	App           string
	Process       string
	Image         string
	Port          int
	ContainerPort int
	Memory        string
	CPU           string
	Command       string
	EnvFile       string
	User          string
	// Health check configuration
	HealthCheckPath     string // HTTP path for health check (e.g., /health)
	HealthCheckInterval int    // Interval in seconds between checks (default: 30)
	HealthCheckTimeout  int    // Timeout in seconds for health check (default: 5)
	HealthCheckRetries  int    // Number of retries before marking unhealthy (default: 3)
}

// Generator creates systemd unit files
type Generator struct {
	unitDir string
	tmpl    *template.Template
}

// New creates a new systemd generator
func New(unitDir string) (*Generator, error) {
	tmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	return &Generator{
		unitDir: unitDir,
		tmpl:    tmpl,
	}, nil
}

// Generate creates a systemd unit file for an app process
func (g *Generator) Generate(cfg *UnitConfig) (string, error) {
	if cfg.ContainerPort == 0 {
		cfg.ContainerPort = 3000
	}
	if cfg.Memory == "" {
		cfg.Memory = "512M"
	}
	if cfg.CPU == "" {
		cfg.CPU = "0.5"
	}
	if cfg.User == "" {
		cfg.User = "pvdify"
	}
	// Health check defaults
	if cfg.HealthCheckInterval == 0 {
		cfg.HealthCheckInterval = 30
	}
	if cfg.HealthCheckTimeout == 0 {
		cfg.HealthCheckTimeout = 5
	}
	if cfg.HealthCheckRetries == 0 {
		cfg.HealthCheckRetries = 3
	}

	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	unitName := fmt.Sprintf("pvdify-%s-%s@.service", cfg.App, cfg.Process)
	unitPath := filepath.Join(g.unitDir, unitName)

	if err := os.WriteFile(unitPath, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("write unit file: %w", err)
	}

	return unitPath, nil
}

// Remove deletes a systemd unit file
func (g *Generator) Remove(app, process string) error {
	unitName := fmt.Sprintf("pvdify-%s-%s@.service", app, process)
	unitPath := filepath.Join(g.unitDir, unitName)

	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove unit file: %w", err)
	}
	return nil
}

// UnitPath returns the path for an app's process unit
func (g *Generator) UnitPath(app, process string) string {
	return filepath.Join(g.unitDir, fmt.Sprintf("pvdify-%s-%s@.service", app, process))
}

const unitTemplate = `[Unit]
Description=Pvdify {{.App}} {{.Process}} process %i
After=network.target
Wants=network.target

[Service]
Type=simple
User={{.User}}
Restart=always
RestartSec=5
TimeoutStartSec=120
TimeoutStopSec=30

# Pull image before starting
ExecStartPre=/usr/bin/podman pull {{.Image}}

# Run container with health check
ExecStart=/usr/bin/podman run --rm \
    --name pvdify-{{.App}}-{{.Process}}-%i \
    -p {{.Port}}:{{.ContainerPort}} \
    --memory={{.Memory}} \
    --cpus={{.CPU}} \
    --env-file {{.EnvFile}} \
{{- if .HealthCheckPath}}
    --health-cmd="curl -sf http://localhost:{{.ContainerPort}}{{.HealthCheckPath}} || exit 1" \
    --health-interval={{.HealthCheckInterval}}s \
    --health-timeout={{.HealthCheckTimeout}}s \
    --health-retries={{.HealthCheckRetries}} \
    --health-start-period=10s \
{{- end}}
    {{.Image}}{{if .Command}} \
    {{.Command}}{{end}}

# Stop container gracefully
ExecStop=/usr/bin/podman stop -t 10 pvdify-{{.App}}-{{.Process}}-%i

# Cleanup on failure
ExecStopPost=-/usr/bin/podman rm -f pvdify-{{.App}}-{{.Process}}-%i

# Environment
Environment=PODMAN_USERNS=keep-id

[Install]
WantedBy=multi-user.target
`
