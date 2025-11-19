package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// InitDB initializes the database connection and creates the database file if it doesn't exist.
// It returns a database connection or an error if initialization fails.
func InitDB() (*sql.DB, error) {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory.")
		return nil, err
	}

	// Create .paso directory if it doesn't exist
	pasoDir := filepath.Join(home, ".paso")
	if err := os.MkdirAll(pasoDir, 0o755); err != nil {
		log.Fatalf("Failed to create directory.")
		return nil, err
	}

	// Open SQLite database
	dbPath := filepath.Join(pasoDir, "tasks.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database.")
		return nil, err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Connection verification failed")
		db.Close()
		return nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
		db.Close()
		return nil, err
	}

	return db, nil
}
