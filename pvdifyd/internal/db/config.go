package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/philoveracity/pvdifyd/internal/models"
)

// CreateConfigVersion inserts a new config version
func (db *DB) CreateConfigVersion(appName string, data []byte) (*models.ConfigVar, error) {
	// Get next version
	var maxVersion sql.NullInt64
	err := db.QueryRow("SELECT MAX(version) FROM config_vars WHERE app_name = ?", appName).Scan(&maxVersion)
	if err != nil {
		return nil, fmt.Errorf("query max config version: %w", err)
	}

	version := 1
	if maxVersion.Valid {
		version = int(maxVersion.Int64) + 1
	}

	now := time.Now()
	result, err := db.Exec(`
		INSERT INTO config_vars (app_name, version, data, created_at)
		VALUES (?, ?, ?, ?)
	`, appName, version, data, now)
	if err != nil {
		return nil, fmt.Errorf("insert config: %w", err)
	}

	id, _ := result.LastInsertId()
	return &models.ConfigVar{
		ID:        id,
		AppName:   appName,
		Version:   version,
		Data:      data,
		CreatedAt: now,
	}, nil
}

// GetLatestConfig retrieves the latest config version for an app
func (db *DB) GetLatestConfig(appName string) (*models.ConfigVar, error) {
	cfg := &models.ConfigVar{}
	err := db.QueryRow(`
		SELECT id, app_name, version, data, created_at
		FROM config_vars WHERE app_name = ? ORDER BY version DESC LIMIT 1
	`, appName).Scan(&cfg.ID, &cfg.AppName, &cfg.Version, &cfg.Data, &cfg.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query config: %w", err)
	}
	return cfg, nil
}

// GetConfigVersion retrieves a specific config version
func (db *DB) GetConfigVersion(appName string, version int) (*models.ConfigVar, error) {
	cfg := &models.ConfigVar{}
	err := db.QueryRow(`
		SELECT id, app_name, version, data, created_at
		FROM config_vars WHERE app_name = ? AND version = ?
	`, appName, version).Scan(&cfg.ID, &cfg.AppName, &cfg.Version, &cfg.Data, &cfg.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query config version: %w", err)
	}
	return cfg, nil
}
