package models

import "time"

// ReleaseStatus represents the state of a release
type ReleaseStatus string

const (
	ReleaseStatusPending  ReleaseStatus = "pending"
	ReleaseStatusDeploying ReleaseStatus = "deploying"
	ReleaseStatusActive   ReleaseStatus = "active"
	ReleaseStatusRolledBack ReleaseStatus = "rolled_back"
	ReleaseStatusFailed   ReleaseStatus = "failed"
)

// Release represents an immutable deployment version
type Release struct {
	ID            int64         `json:"id" db:"id"`
	AppName       string        `json:"app_name" db:"app_name"`
	Version       int           `json:"version" db:"version"`
	Image         string        `json:"image" db:"image"`
	ConfigVersion int           `json:"config_version,omitempty" db:"config_version"`
	Status        ReleaseStatus `json:"status" db:"status"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	CreatedBy     string        `json:"created_by,omitempty" db:"created_by"`
}

// CreateReleaseRequest is the payload for creating a new release (deploy)
type CreateReleaseRequest struct {
	Image     string `json:"image" validate:"required"`
	CreatedBy string `json:"created_by,omitempty"`
}

// RollbackRequest is the payload for rolling back to a previous release
type RollbackRequest struct {
	Version int `json:"version,omitempty"` // If omitted, rollback to previous
}
