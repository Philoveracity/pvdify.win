package models

import "time"

// AppStatus represents the lifecycle state of an app
type AppStatus string

const (
	AppStatusCreated  AppStatus = "created"
	AppStatusRunning  AppStatus = "running"
	AppStatusStopped  AppStatus = "stopped"
	AppStatusFailed   AppStatus = "failed"
	AppStatusDeleting AppStatus = "deleting"
)

// App represents a deployable application slot
type App struct {
	Name        string            `json:"name" db:"name"`
	Environment string            `json:"environment" db:"environment"`
	Status      AppStatus         `json:"status" db:"status"`
	Image       string            `json:"image,omitempty" db:"image"`
	BindPort    int               `json:"bind_port,omitempty" db:"bind_port"`
	Resources   *ResourceLimits   `json:"resources,omitempty"`
	Healthcheck *HealthcheckConfig `json:"healthcheck,omitempty"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// ResourceLimits defines container resource constraints
type ResourceLimits struct {
	Memory string  `json:"memory" yaml:"memory"` // e.g., "512M"
	CPU    float64 `json:"cpu" yaml:"cpu"`       // e.g., 0.5
}

// HealthcheckConfig defines health check parameters
type HealthcheckConfig struct {
	Path     string `json:"path" yaml:"path"`         // e.g., "/health"
	Interval string `json:"interval" yaml:"interval"` // e.g., "30s"
	Timeout  string `json:"timeout" yaml:"timeout"`   // e.g., "5s"
}

// CreateAppRequest is the payload for creating a new app
type CreateAppRequest struct {
	Name        string `json:"name" validate:"required,alphanum"`
	Environment string `json:"environment,omitempty"`
}

// UpdateAppRequest is the payload for updating an app
type UpdateAppRequest struct {
	Image       *string         `json:"image,omitempty"`
	Resources   *ResourceLimits `json:"resources,omitempty"`
	Healthcheck *HealthcheckConfig `json:"healthcheck,omitempty"`
}
