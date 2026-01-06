package systemd

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Manager controls systemd services
type Manager struct{}

// NewManager creates a new systemd manager
func NewManager() *Manager {
	return &Manager{}
}

// DaemonReload runs systemctl daemon-reload
func (m *Manager) DaemonReload(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("daemon-reload: %s: %w", string(output), err)
	}
	return nil
}

// Start starts a service instance
func (m *Manager) Start(ctx context.Context, unit string, instance int) error {
	name := fmt.Sprintf("%s@%d", unit, instance)
	cmd := exec.CommandContext(ctx, "systemctl", "start", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("start %s: %s: %w", name, string(output), err)
	}
	return nil
}

// Stop stops a service instance
func (m *Manager) Stop(ctx context.Context, unit string, instance int) error {
	name := fmt.Sprintf("%s@%d", unit, instance)
	cmd := exec.CommandContext(ctx, "systemctl", "stop", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("stop %s: %s: %w", name, string(output), err)
	}
	return nil
}

// Restart restarts a service instance
func (m *Manager) Restart(ctx context.Context, unit string, instance int) error {
	name := fmt.Sprintf("%s@%d", unit, instance)
	cmd := exec.CommandContext(ctx, "systemctl", "restart", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("restart %s: %s: %w", name, string(output), err)
	}
	return nil
}

// Enable enables a service
func (m *Manager) Enable(ctx context.Context, unit string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "enable", unit)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("enable %s: %s: %w", unit, string(output), err)
	}
	return nil
}

// Disable disables a service
func (m *Manager) Disable(ctx context.Context, unit string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "disable", unit)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("disable %s: %s: %w", unit, string(output), err)
	}
	return nil
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Active    string
	SubState  string
	MainPID   int
	Memory    string
	LoadState string
}

// Status returns the status of a service instance
func (m *Manager) Status(ctx context.Context, unit string, instance int) (*ServiceStatus, error) {
	name := fmt.Sprintf("%s@%d", unit, instance)
	cmd := exec.CommandContext(ctx, "systemctl", "show", name,
		"--property=ActiveState,SubState,MainPID,MemoryCurrent,LoadState")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("show %s: %w", name, err)
	}

	status := &ServiceStatus{}
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		switch key {
		case "ActiveState":
			status.Active = value
		case "SubState":
			status.SubState = value
		case "MainPID":
			status.MainPID, _ = strconv.Atoi(value)
		case "MemoryCurrent":
			status.Memory = formatBytes(value)
		case "LoadState":
			status.LoadState = value
		}
	}

	return status, nil
}

// ListInstances returns running instance numbers for a unit
func (m *Manager) ListInstances(ctx context.Context, unit string) ([]int, error) {
	pattern := strings.TrimSuffix(unit, ".service") + "@*.service"
	cmd := exec.CommandContext(ctx, "systemctl", "list-units", pattern, "--no-legend", "--plain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list units: %w", err)
	}

	var instances []int
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		name := fields[0]
		// Extract instance number from name like pvdify-app-web@1.service
		atIdx := strings.Index(name, "@")
		dotIdx := strings.Index(name, ".service")
		if atIdx != -1 && dotIdx > atIdx {
			numStr := name[atIdx+1 : dotIdx]
			if num, err := strconv.Atoi(numStr); err == nil {
				instances = append(instances, num)
			}
		}
	}

	return instances, nil
}

func formatBytes(s string) string {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return s
	}
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.1fG", float64(n)/gb)
	case n >= mb:
		return fmt.Sprintf("%.1fM", float64(n)/mb)
	case n >= kb:
		return fmt.Sprintf("%.1fK", float64(n)/kb)
	default:
		return fmt.Sprintf("%dB", n)
	}
}
