package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/philoveracity/pvdifyd/internal/models"
)

// CreateApp inserts a new app
func (db *DB) CreateApp(app *models.App) error {
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now
	if app.Environment == "" {
		app.Environment = "production"
	}
	if app.Status == "" {
		app.Status = models.AppStatusCreated
	}

	_, err := db.Exec(`
		INSERT INTO apps (name, environment, status, image, bind_port, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, app.Name, app.Environment, app.Status, app.Image, app.BindPort, app.CreatedAt, app.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert app: %w", err)
	}
	return nil
}

// GetApp retrieves an app by name
func (db *DB) GetApp(name string) (*models.App, error) {
	app := &models.App{}
	var image sql.NullString
	var bindPort sql.NullInt64

	err := db.QueryRow(`
		SELECT name, environment, status, image, bind_port, created_at, updated_at
		FROM apps WHERE name = ?
	`, name).Scan(&app.Name, &app.Environment, &app.Status, &image, &bindPort, &app.CreatedAt, &app.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query app: %w", err)
	}

	if image.Valid {
		app.Image = image.String
	}
	if bindPort.Valid {
		app.BindPort = int(bindPort.Int64)
	}

	return app, nil
}

// ListApps retrieves all apps
func (db *DB) ListApps() ([]*models.App, error) {
	rows, err := db.Query(`
		SELECT name, environment, status, image, bind_port, created_at, updated_at
		FROM apps ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("query apps: %w", err)
	}
	defer rows.Close()

	var apps []*models.App
	for rows.Next() {
		app := &models.App{}
		var image sql.NullString
		var bindPort sql.NullInt64

		if err := rows.Scan(&app.Name, &app.Environment, &app.Status, &image, &bindPort, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan app: %w", err)
		}

		if image.Valid {
			app.Image = image.String
		}
		if bindPort.Valid {
			app.BindPort = int(bindPort.Int64)
		}

		apps = append(apps, app)
	}

	return apps, nil
}

// UpdateApp updates an existing app
func (db *DB) UpdateApp(name string, image *string, status *models.AppStatus, bindPort *int) error {
	updates := "updated_at = ?"
	args := []interface{}{time.Now()}

	if image != nil {
		updates += ", image = ?"
		args = append(args, *image)
	}
	if status != nil {
		updates += ", status = ?"
		args = append(args, *status)
	}
	if bindPort != nil {
		updates += ", bind_port = ?"
		args = append(args, *bindPort)
	}

	args = append(args, name)
	_, err := db.Exec("UPDATE apps SET "+updates+" WHERE name = ?", args...)
	if err != nil {
		return fmt.Errorf("update app: %w", err)
	}
	return nil
}

// DeleteApp removes an app
func (db *DB) DeleteApp(name string) error {
	result, err := db.Exec("DELETE FROM apps WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete app: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("app not found: %s", name)
	}
	return nil
}

// AllocatePort finds the next available port
func (db *DB) AllocatePort(startPort, endPort int) (int, error) {
	var maxPort sql.NullInt64
	err := db.QueryRow("SELECT MAX(bind_port) FROM apps WHERE bind_port IS NOT NULL").Scan(&maxPort)
	if err != nil {
		return 0, fmt.Errorf("query max port: %w", err)
	}

	nextPort := startPort
	if maxPort.Valid && int(maxPort.Int64) >= startPort {
		nextPort = int(maxPort.Int64) + 1
	}

	if nextPort > endPort {
		return 0, fmt.Errorf("no ports available in range %d-%d", startPort, endPort)
	}

	return nextPort, nil
}
