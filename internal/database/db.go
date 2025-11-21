// Package database handles the initialization and connection to the SQLite db
package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB() (*sql.DB, error) {
	// adding paso database to the home dir
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory.")
		return nil, err
	}

	pasoDir := filepath.Join(home, ".paso")
	if err := os.MkdirAll(pasoDir, 0o755); err != nil {
		log.Fatalf("Failed to create directory.")
		return nil, err
	}

	dbPath := filepath.Join(pasoDir, "tasks.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database.")
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping failed failed")
		db.Close()
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("Migrations: %v", err)
		db.Close()
		return nil, err
	}

	return db, nil
}
