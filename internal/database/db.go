// Package database handles the initialization and connection to the SQLite db
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB() (*sql.DB, error) {
	// adding paso database to the home dir
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	pasoDir := filepath.Join(home, ".paso")
	if err := os.MkdirAll(pasoDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	dbPath := filepath.Join(pasoDir, "tasks.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign key constraints (required for CASCADE deletions)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Printf("Failed to enable foreign keys: %v", err)
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("error closing db: %v", closeErr)
		}
		return nil, err
	}

	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("error closing db: %v", closeErr)
		}
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	if err := runMigrations(context.Background(), db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("error closing db: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
