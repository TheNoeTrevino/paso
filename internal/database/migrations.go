package database

import "database/sql"

// runMigrations creates the database schema and seeds default data if needed
func runMigrations(db *sql.DB) error {
	// Create columns table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS columns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			position INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create tasks table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			column_id INTEGER NOT NULL,
			position INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create index for efficient queries
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_column
		ON tasks(column_id, position)
	`)
	if err != nil {
		return err
	}

	// Seed default columns if the table is empty
	if err := seedDefaultColumns(db); err != nil {
		return err
	}

	return nil
}

// seedDefaultColumns inserts default columns if the columns table is empty
func seedDefaultColumns(db *sql.DB) error {
	// Check if columns table is empty
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM columns").Scan(&count)
	if err != nil {
		return err
	}

	// If columns exist, don't seed
	if count > 0 {
		return nil
	}

	// Insert default columns
	defaultColumns := []struct {
		name     string
		position int
	}{
		{"Todo", 0},
		{"In Progress", 1},
		{"Done", 2},
	}

	for _, col := range defaultColumns {
		_, err := db.Exec(
			"INSERT INTO columns (name, position) VALUES (?, ?)",
			col.name, col.position,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
