package database

import "database/sql"

// runMigrations creates the database schema and seeds default data if needed
func runMigrations(db *sql.DB) error {
	// Create projects table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create project_counters table for ticket numbering per project
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS project_counters (
			project_id INTEGER PRIMARY KEY,
			next_ticket_number INTEGER DEFAULT 1,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create columns table
	_, err = db.Exec(`
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

	// Create labels table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			color TEXT NOT NULL DEFAULT '#7D56F4'
		)
	`)
	if err != nil {
		return err
	}

	// Create task_labels join table (many-to-many)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS task_labels (
			task_id INTEGER NOT NULL,
			label_id INTEGER NOT NULL,
			PRIMARY KEY (task_id, label_id),
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Seed default project if no projects exist
	if err := seedDefaultProject(db); err != nil {
		return err
	}

	// Seed default columns if the table is empty
	if err := seedDefaultColumns(db); err != nil {
		return err
	}

	// Migrate to linked list structure
	if err := migrateToLinkedList(db); err != nil {
		return err
	}

	// Migrate columns to include project_id
	if err := migrateColumnsToProject(db); err != nil {
		return err
	}

	// Migrate tasks to include ticket_number
	if err := migrateTasksTicketNumber(db); err != nil {
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

// migrateToLinkedList converts the position-based column ordering to a linked list structure
// This migration is idempotent and can be run multiple times safely
func migrateToLinkedList(db *sql.DB) error {
	// Check if prev_id column already exists
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('columns')
		WHERE name IN ('prev_id', 'next_id')
	`).Scan(&count)
	if err != nil {
		return err
	}

	hasLinkedListColumns := count == 2

	// If migration already complete, skip
	if hasLinkedListColumns {
		return nil
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Add new columns for linked list structure
	_, err = tx.Exec(`ALTER TABLE columns ADD COLUMN prev_id INTEGER NULL`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`ALTER TABLE columns ADD COLUMN next_id INTEGER NULL`)
	if err != nil {
		return err
	}

	// 2. Migrate existing data: query all columns ordered by position
	rows, err := tx.Query(`SELECT id FROM columns ORDER BY position`)
	if err != nil {
		return err
	}

	var columnIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		columnIDs = append(columnIDs, id)
	}
	rows.Close()

	// 3. Build linked list by setting prev_id and next_id
	for i, id := range columnIDs {
		var prevID *int
		var nextID *int

		// Set prev_id (NULL for first column)
		if i > 0 {
			prevID = &columnIDs[i-1]
		}

		// Set next_id (NULL for last column)
		if i < len(columnIDs)-1 {
			nextID = &columnIDs[i+1]
		}

		// Update the column with linked list pointers
		_, err = tx.Exec(`
			UPDATE columns
			SET prev_id = ?, next_id = ?
			WHERE id = ?
		`, prevID, nextID, id)
		if err != nil {
			return err
		}
	}

	// 4. Drop the old position column
	// SQLite doesn't support DROP COLUMN directly before version 3.35.0
	// We'll create a new table and copy data
	_, err = tx.Exec(`
		CREATE TABLE columns_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			prev_id INTEGER NULL,
			next_id INTEGER NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO columns_new (id, name, prev_id, next_id)
		SELECT id, name, prev_id, next_id FROM columns
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DROP TABLE columns`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`ALTER TABLE columns_new RENAME TO columns`)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// seedDefaultProject creates a default project if no projects exist
func seedDefaultProject(db *sql.DB) error {
	// Check if projects table is empty
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count)
	if err != nil {
		return err
	}

	// If projects exist, don't seed
	if count > 0 {
		return nil
	}

	// Insert default project
	result, err := db.Exec(
		`INSERT INTO projects (name, description) VALUES (?, ?)`,
		"Default", "Default project",
	)
	if err != nil {
		return err
	}

	// Initialize the counter for the default project
	projectID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)`,
		projectID,
	)
	return err
}

// migrateColumnsToProject adds project_id column to columns table and associates existing columns with default project
func migrateColumnsToProject(db *sql.DB) error {
	// Check if project_id column already exists
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('columns')
		WHERE name = 'project_id'
	`).Scan(&count)
	if err != nil {
		return err
	}

	// If column exists, skip migration
	if count > 0 {
		return nil
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the default project ID
	var defaultProjectID int
	err = tx.QueryRow(`SELECT id FROM projects WHERE name = 'Default' LIMIT 1`).Scan(&defaultProjectID)
	if err != nil {
		return err
	}

	// Add project_id column
	_, err = tx.Exec(`ALTER TABLE columns ADD COLUMN project_id INTEGER NOT NULL DEFAULT 1`)
	if err != nil {
		return err
	}

	// Update all existing columns to belong to the default project
	_, err = tx.Exec(`UPDATE columns SET project_id = ?`, defaultProjectID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// migrateTasksTicketNumber adds ticket_number column to tasks table
func migrateTasksTicketNumber(db *sql.DB) error {
	// Check if ticket_number column already exists
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks')
		WHERE name = 'ticket_number'
	`).Scan(&count)
	if err != nil {
		return err
	}

	// If column exists, skip migration
	if count > 0 {
		return nil
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Add ticket_number column
	_, err = tx.Exec(`ALTER TABLE tasks ADD COLUMN ticket_number INTEGER`)
	if err != nil {
		return err
	}

	// Assign ticket numbers to existing tasks grouped by their project
	// First, get all tasks ordered by id (oldest first) with their project info
	rows, err := tx.Query(`
		SELECT t.id, c.project_id
		FROM tasks t
		JOIN columns c ON t.column_id = c.id
		ORDER BY c.project_id, t.id
	`)
	if err != nil {
		return err
	}

	// Track counters per project
	projectCounters := make(map[int]int)
	type taskUpdate struct {
		id           int
		ticketNumber int
	}
	var updates []taskUpdate

	for rows.Next() {
		var taskID, projectID int
		if err := rows.Scan(&taskID, &projectID); err != nil {
			rows.Close()
			return err
		}

		// Get next ticket number for this project
		counter := projectCounters[projectID]
		counter++
		projectCounters[projectID] = counter

		updates = append(updates, taskUpdate{id: taskID, ticketNumber: counter})
	}
	rows.Close()

	// Apply updates
	for _, u := range updates {
		_, err = tx.Exec(`UPDATE tasks SET ticket_number = ? WHERE id = ?`, u.ticketNumber, u.id)
		if err != nil {
			return err
		}
	}

	// Update project counters to reflect the highest ticket number used
	for projectID, counter := range projectCounters {
		_, err = tx.Exec(`
			INSERT OR REPLACE INTO project_counters (project_id, next_ticket_number)
			VALUES (?, ?)
		`, projectID, counter+1)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
