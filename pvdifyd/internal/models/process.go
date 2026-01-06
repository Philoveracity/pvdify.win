package models

// Process represents a process type definition for an app
type Process struct {
	ID      int64  `json:"id" db:"id"`
	AppName string `json:"app_name" db:"app_name"`
	Name    string `json:"name" db:"name"` // e.g., "web", "worker"
	Command string `json:"command,omitempty" db:"command"`
	Count   int    `json:"count" db:"count"`
}

// ProcessStatus represents runtime status of a process instance
type ProcessStatus struct {
	Name      string `json:"name"`
	Instance  int    `json:"instance"`
	State     string `json:"state"` // running, stopped, failed
	Container string `json:"container,omitempty"`
	Uptime    string `json:"uptime,omitempty"`
}

// ScaleRequest is the payload for scaling processes
type ScaleRequest struct {
	Processes map[string]int `json:"processes" validate:"required"` // e.g., {"web": 2, "worker": 1}
}
