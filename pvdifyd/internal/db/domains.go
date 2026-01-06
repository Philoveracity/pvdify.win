package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/philoveracity/pvdifyd/internal/models"
)

// CreateDomain inserts a new domain
func (db *DB) CreateDomain(domain *models.Domain) error {
	domain.CreatedAt = time.Now()
	if domain.Status == "" {
		domain.Status = models.DomainStatusPending
	}

	_, err := db.Exec(`
		INSERT INTO domains (domain, app_name, status, cf_record_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, domain.Domain, domain.AppName, domain.Status, domain.CFRecordID, domain.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert domain: %w", err)
	}
	return nil
}

// GetDomain retrieves a domain by name
func (db *DB) GetDomain(domain string) (*models.Domain, error) {
	d := &models.Domain{}
	var cfRecordID sql.NullString

	err := db.QueryRow(`
		SELECT domain, app_name, status, cf_record_id, created_at
		FROM domains WHERE domain = ?
	`, domain).Scan(&d.Domain, &d.AppName, &d.Status, &cfRecordID, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query domain: %w", err)
	}

	if cfRecordID.Valid {
		d.CFRecordID = cfRecordID.String
	}

	return d, nil
}

// ListDomains retrieves all domains for an app
func (db *DB) ListDomains(appName string) ([]*models.Domain, error) {
	rows, err := db.Query(`
		SELECT domain, app_name, status, cf_record_id, created_at
		FROM domains WHERE app_name = ? ORDER BY domain
	`, appName)
	if err != nil {
		return nil, fmt.Errorf("query domains: %w", err)
	}
	defer rows.Close()

	var domains []*models.Domain
	for rows.Next() {
		d := &models.Domain{}
		var cfRecordID sql.NullString

		if err := rows.Scan(&d.Domain, &d.AppName, &d.Status, &cfRecordID, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan domain: %w", err)
		}

		if cfRecordID.Valid {
			d.CFRecordID = cfRecordID.String
		}

		domains = append(domains, d)
	}

	return domains, nil
}

// UpdateDomainStatus updates the status and optional CF record ID
func (db *DB) UpdateDomainStatus(domain string, status models.DomainStatus, cfRecordID *string) error {
	if cfRecordID != nil {
		_, err := db.Exec("UPDATE domains SET status = ?, cf_record_id = ? WHERE domain = ?",
			status, *cfRecordID, domain)
		return err
	}
	_, err := db.Exec("UPDATE domains SET status = ? WHERE domain = ?", status, domain)
	return err
}

// DeleteDomain removes a domain
func (db *DB) DeleteDomain(domain string) error {
	result, err := db.Exec("DELETE FROM domains WHERE domain = ?", domain)
	if err != nil {
		return fmt.Errorf("delete domain: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("domain not found: %s", domain)
	}
	return nil
}
