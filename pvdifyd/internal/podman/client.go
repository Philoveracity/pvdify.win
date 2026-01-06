package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Client interfaces with Podman
type Client struct {
	socket string
	http   *http.Client
}

// New creates a new Podman client
func New(socket string) *Client {
	// Create HTTP client with Unix socket transport
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			socketPath := strings.TrimPrefix(socket, "unix://")
			return net.Dial("unix", socketPath)
		},
	}

	return &Client{
		socket: socket,
		http: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// Ping checks if Podman is available
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.http.Get("http://d/v4.0.0/libpod/_ping")
	if err != nil {
		return fmt.Errorf("ping podman: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("podman ping returned %d", resp.StatusCode)
	}
	return nil
}

// ContainerInfo represents container information
type ContainerInfo struct {
	ID      string `json:"Id"`
	Name    string `json:"Name"`
	State   string `json:"State"`
	Image   string `json:"Image"`
	Created string `json:"Created"`
	Ports   []Port `json:"Ports"`
}

// Port represents a port mapping
type Port struct {
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

// ListContainers returns all containers matching a filter
func (c *Client) ListContainers(ctx context.Context, namePrefix string) ([]ContainerInfo, error) {
	filter := fmt.Sprintf(`{"name":["%s"]}`, namePrefix)
	url := fmt.Sprintf("http://d/v4.0.0/libpod/containers/json?all=true&filters=%s", filter)

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	defer resp.Body.Close()

	var containers []ContainerInfo
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("decode containers: %w", err)
	}

	return containers, nil
}

// PullImage pulls an image from a registry
func (c *Client) PullImage(ctx context.Context, image string) error {
	// Use podman CLI for pulling (more reliable with auth)
	cmd := exec.CommandContext(ctx, "podman", "pull", image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pull image: %s: %w", string(output), err)
	}
	return nil
}

// ImageExists checks if an image exists locally
func (c *Client) ImageExists(ctx context.Context, image string) (bool, error) {
	url := fmt.Sprintf("http://d/v4.0.0/libpod/images/%s/exists", image)
	resp, err := c.http.Get(url)
	if err != nil {
		return false, fmt.Errorf("check image: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusNoContent, nil
}

// StopContainer stops a container
func (c *Client) StopContainer(ctx context.Context, name string, timeout int) error {
	url := fmt.Sprintf("http://d/v4.0.0/libpod/containers/%s/stop?timeout=%d", name, timeout)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("stop container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotModified {
		return fmt.Errorf("stop container returned %d", resp.StatusCode)
	}
	return nil
}

// RemoveContainer removes a container
func (c *Client) RemoveContainer(ctx context.Context, name string, force bool) error {
	url := fmt.Sprintf("http://d/v4.0.0/libpod/containers/%s?force=%t", name, force)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("remove container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("remove container returned %d", resp.StatusCode)
	}
	return nil
}

// ContainerStats represents container resource usage
type ContainerStats struct {
	CPU    float64 `json:"cpu_percent"`
	Memory int64   `json:"memory"`
}

// GetStats returns resource stats for a container
func (c *Client) GetStats(ctx context.Context, name string) (*ContainerStats, error) {
	url := fmt.Sprintf("http://d/v4.0.0/libpod/containers/%s/stats?stream=false", name)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	defer resp.Body.Close()

	var stats ContainerStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("decode stats: %w", err)
	}

	return &stats, nil
}

// HealthCheck performs a health check on a container's exposed port
func (c *Client) HealthCheck(ctx context.Context, port int, path string) error {
	url := fmt.Sprintf("http://localhost:%d%s", port, path)
	client := &http.Client{Timeout: 5 * time.Second}

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("health check returned %d", resp.StatusCode)
	}
	return nil
}

// ExecResult represents command execution result
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Exec runs a command inside a container
func (c *Client) Exec(ctx context.Context, name string, cmd []string) (*ExecResult, error) {
	args := append([]string{"exec", name}, cmd...)
	command := exec.CommandContext(ctx, "podman", args...)

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return &ExecResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}
