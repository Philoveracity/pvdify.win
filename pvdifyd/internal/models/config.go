package models

import "time"

// ConfigVar represents a versioned configuration snapshot
type ConfigVar struct {
	ID        int64     `json:"id" db:"id"`
	AppName   string    `json:"app_name" db:"app_name"`
	Version   int       `json:"version" db:"version"`
	Data      []byte    `json:"-" db:"data"` // SOPS encrypted YAML
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ConfigData represents decrypted config key-value pairs
type ConfigData map[string]string

// SetConfigRequest is the payload for setting config vars
type SetConfigRequest struct {
	Vars map[string]string `json:"vars" validate:"required"`
}
