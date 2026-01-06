package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/philoveracity/pvdifyd/internal/models"
)

// CreateRelease inserts a new release
func (db *DB) CreateRelease(release *models.Release) error {
	// Get next version
	var maxVersion sql.NullInt64
	err := db.QueryRow("SELECT MAX(version) FROM releases WHERE app_name = ?", release.AppName).Scan(&maxVersion)
	if err != nil {
		return fmt.Errorf("query max version: %w", err)
	}

	release.Version = 1
	if maxVersion.Valid {
		release.Version = int(maxVersion.Int64) + 1
	}
	release.CreatedAt = time.Now()
	if release.Status == "" {
		release.Status = models.ReleaseStatusPending
	}

	result, err := db.Exec(`
		INSERT INTO releases (app_name, version, image, config_version, status, created_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, release.AppName, release.Version, release.Image, release.ConfigVersion, release.Status, release.CreatedAt, release.CreatedBy)
	if err != nil {
		return fmt.Errorf("insert release: %w", err)
	}

	id, _ := result.LastInsertId()
	release.ID = id
	return nil
}

// GetRelease retrieves a release by app name and version
func (db *DB) GetRelease(appName string, version int) (*models.Release, error) {
	release := &models.Release{}
	var configVersion sql.NullInt64
	var createdBy sql.NullString

	err := db.QueryRow(`
		SELECT id, app_name, version, image, config_version, status, created_at, created_by
		FROM releases WHERE app_name = ? AND version = ?
	`, appName, version).Scan(&release.ID, &release.AppName, &release.Version, &release.Image,
		&configVersion, &release.Status, &release.CreatedAt, &createdBy)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query release: %w", err)
	}

	if configVersion.Valid {
		release.ConfigVersion = int(configVersion.Int64)
	}
	if createdBy.Valid {
		release.CreatedBy = createdBy.String
	}

	return release, nil
}

// GetLatestRelease retrieves the most recent release for an app
func (db *DB) GetLatestRelease(appName string) (*models.Release, error) {
	release := &models.Release{}
	var configVersion sql.NullInt64
	var createdBy sql.NullString

	err := db.QueryRow(`
		SELECT id, app_name, version, image, config_version, status, created_at, created_by
		FROM releases WHERE app_name = ? ORDER BY version DESC LIMIT 1
	`, appName).Scan(&release.ID, &release.AppName, &release.Version, &release.Image,
		&configVersion, &release.Status, &release.CreatedAt, &createdBy)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query latest release: %w", err)
	}

	if configVersion.Valid {
		release.ConfigVersion = int(configVersion.Int64)
	}
	if createdBy.Valid {
		release.CreatedBy = createdBy.String
	}

	return release, nil
}

// GetActiveRelease retrieves the currently active release
func (db *DB) GetActiveRelease(appName string) (*models.Release, error) {
	release := &models.Release{}
	var configVersion sql.NullInt64
	var createdBy sql.NullString

	err := db.QueryRow(`
		SELECT id, app_name, version, image, config_version, status, created_at, created_by
		FROM releases WHERE app_name = ? AND status = 'active' ORDER BY version DESC LIMIT 1
	`, appName).Scan(&release.ID, &release.AppName, &release.Version, &release.Image,
		&configVersion, &release.Status, &release.CreatedAt, &createdBy)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query active release: %w", err)
	}

	if configVersion.Valid {
		release.ConfigVersion = int(configVersion.Int64)
	}
	if createdBy.Valid {
		release.CreatedBy = createdBy.String
	}

	return release, nil
}

// ListReleases retrieves all releases for an app
func (db *DB) ListReleases(appName string, limit int) ([]*models.Release, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := db.Query(`
		SELECT id, app_name, version, image, config_version, status, created_at, created_by
		FROM releases WHERE app_name = ? ORDER BY version DESC LIMIT ?
	`, appName, limit)
	if err != nil {
		return nil, fmt.Errorf("query releases: %w", err)
	}
	defer rows.Close()

	var releases []*models.Release
	for rows.Next() {
		release := &models.Release{}
		var configVersion sql.NullInt64
		var createdBy sql.NullString

		if err := rows.Scan(&release.ID, &release.AppName, &release.Version, &release.Image,
			&configVersion, &release.Status, &release.CreatedAt, &createdBy); err != nil {
			return nil, fmt.Errorf("scan release: %w", err)
		}

		if configVersion.Valid {
			release.ConfigVersion = int(configVersion.Int64)
		}
		if createdBy.Valid {
			release.CreatedBy = createdBy.String
		}

		releases = append(releases, release)
	}

	return releases, nil
}

// UpdateReleaseStatus updates the status of a release
func (db *DB) UpdateReleaseStatus(appName string, version int, status models.ReleaseStatus) error {
	_, err := db.Exec("UPDATE releases SET status = ? WHERE app_name = ? AND version = ?",
		status, appName, version)
	if err != nil {
		return fmt.Errorf("update release status: %w", err)
	}
	return nil
}
