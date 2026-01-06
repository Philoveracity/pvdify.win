package models

import "time"

// DomainStatus represents the state of a domain mapping
type DomainStatus string

const (
	DomainStatusPending  DomainStatus = "pending"
	DomainStatusActive   DomainStatus = "active"
	DomainStatusFailed   DomainStatus = "failed"
	DomainStatusDeleting DomainStatus = "deleting"
)

// Domain represents a custom domain mapped to an app
type Domain struct {
	Domain     string       `json:"domain" db:"domain"`
	AppName    string       `json:"app_name" db:"app_name"`
	Status     DomainStatus `json:"status" db:"status"`
	CFRecordID string       `json:"cf_record_id,omitempty" db:"cf_record_id"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
}

// AddDomainRequest is the payload for adding a domain to an app
type AddDomainRequest struct {
	Domain string `json:"domain" validate:"required,fqdn"`
}
