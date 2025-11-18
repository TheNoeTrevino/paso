package database

import (
	"database/sql"
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
		return nil, err
	}

	// Create .paso directory if it doesn't exist
	pasoDir := filepath.Join(home, ".paso")
	if err := os.MkdirAll(pasoDir, 0755); err != nil {
		return nil, err
	}

	// Open SQLite database
	dbPath := filepath.Join(pasoDir, "tasks.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
