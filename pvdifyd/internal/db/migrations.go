package db

import (
	"fmt"
)

// Migrate runs database migrations
func (db *DB) Migrate() error {
	// Create migrations table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Get current version
	var currentVersion int
	row := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM migrations")
	if err := row.Scan(&currentVersion); err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	// Run pending migrations
	for i, migration := range migrations {
		version := i + 1
		if version <= currentVersion {
			continue
		}

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("run migration %d: %w", version, err)
		}

		if _, err := db.Exec("INSERT INTO migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("record migration %d: %w", version, err)
		}
	}

	return nil
}

var migrations = []string{
	// Migration 1: Initial schema
	`
	CREATE TABLE IF NOT EXISTS apps (
		name TEXT PRIMARY KEY,
		environment TEXT NOT NULL DEFAULT 'production',
		status TEXT NOT NULL DEFAULT 'created',
		image TEXT,
		bind_port INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS releases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_name TEXT NOT NULL,
		version INTEGER NOT NULL,
		image TEXT NOT NULL,
		config_version INTEGER,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_by TEXT,
		UNIQUE(app_name, version),
		FOREIGN KEY (app_name) REFERENCES apps(name) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS domains (
		domain TEXT PRIMARY KEY,
		app_name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		cf_record_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (app_name) REFERENCES apps(name) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS config_vars (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_name TEXT NOT NULL,
		version INTEGER NOT NULL,
		data BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(app_name, version),
		FOREIGN KEY (app_name) REFERENCES apps(name) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS processes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_name TEXT NOT NULL,
		name TEXT NOT NULL,
		command TEXT,
		count INTEGER NOT NULL DEFAULT 1,
		UNIQUE(app_name, name),
		FOREIGN KEY (app_name) REFERENCES apps(name) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_releases_app_name ON releases(app_name);
	CREATE INDEX IF NOT EXISTS idx_releases_status ON releases(status);
	CREATE INDEX IF NOT EXISTS idx_domains_app_name ON domains(app_name);
	CREATE INDEX IF NOT EXISTS idx_config_vars_app_name ON config_vars(app_name);
	CREATE INDEX IF NOT EXISTS idx_processes_app_name ON processes(app_name);
	`,
}
