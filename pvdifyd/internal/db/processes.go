package db

import (
	"database/sql"
	"fmt"

	"github.com/philoveracity/pvdifyd/internal/models"
)

// UpsertProcess creates or updates a process definition
func (db *DB) UpsertProcess(process *models.Process) error {
	_, err := db.Exec(`
		INSERT INTO processes (app_name, name, command, count)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(app_name, name) DO UPDATE SET
			command = excluded.command,
			count = excluded.count
	`, process.AppName, process.Name, process.Command, process.Count)
	if err != nil {
		return fmt.Errorf("upsert process: %w", err)
	}
	return nil
}

// GetProcess retrieves a process by app name and process name
func (db *DB) GetProcess(appName, name string) (*models.Process, error) {
	p := &models.Process{}
	var command sql.NullString

	err := db.QueryRow(`
		SELECT id, app_name, name, command, count
		FROM processes WHERE app_name = ? AND name = ?
	`, appName, name).Scan(&p.ID, &p.AppName, &p.Name, &command, &p.Count)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query process: %w", err)
	}

	if command.Valid {
		p.Command = command.String
	}

	return p, nil
}

// ListProcesses retrieves all processes for an app
func (db *DB) ListProcesses(appName string) ([]*models.Process, error) {
	rows, err := db.Query(`
		SELECT id, app_name, name, command, count
		FROM processes WHERE app_name = ? ORDER BY name
	`, appName)
	if err != nil {
		return nil, fmt.Errorf("query processes: %w", err)
	}
	defer rows.Close()

	var processes []*models.Process
	for rows.Next() {
		p := &models.Process{}
		var command sql.NullString

		if err := rows.Scan(&p.ID, &p.AppName, &p.Name, &command, &p.Count); err != nil {
			return nil, fmt.Errorf("scan process: %w", err)
		}

		if command.Valid {
			p.Command = command.String
		}

		processes = append(processes, p)
	}

	return processes, nil
}

// ScaleProcess updates the count for a process
func (db *DB) ScaleProcess(appName, name string, count int) error {
	result, err := db.Exec("UPDATE processes SET count = ? WHERE app_name = ? AND name = ?",
		count, appName, name)
	if err != nil {
		return fmt.Errorf("scale process: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		// Create new process with default command
		return db.UpsertProcess(&models.Process{
			AppName: appName,
			Name:    name,
			Count:   count,
		})
	}
	return nil
}

// DeleteProcess removes a process
func (db *DB) DeleteProcess(appName, name string) error {
	_, err := db.Exec("DELETE FROM processes WHERE app_name = ? AND name = ?", appName, name)
	return err
}
