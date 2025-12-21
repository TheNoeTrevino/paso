package database

import (
	"context"
	"database/sql"
	"log"
)

// runMigrations creates the database schema and seeds default data if needed
func runMigrations(ctx context.Context, db *sql.DB) error {
	// Create projects table
	_, err := db.ExecContext(ctx, `
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
	_, err = db.ExecContext(ctx, `
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
	_, err = db.ExecContext(ctx, `
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
	_, err = db.ExecContext(ctx, `
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

	// Create index for efficient task queries
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_tasks_column
		ON tasks(column_id, position)
	`)
	if err != nil {
		return err
	}

	// Create labels table (project-specific)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			color TEXT NOT NULL DEFAULT '#7D56F4',
			project_id INTEGER NOT NULL DEFAULT 1,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			UNIQUE(name, project_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create task_labels join table (many-to-many)
	_, err = db.ExecContext(ctx, `
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
	if err := seedDefaultProject(ctx, db); err != nil {
		return err
	}

	// Seed default columns if the table is empty
	if err := seedDefaultColumns(ctx, db); err != nil {
		return err
	}

	// Migrate to linked list structure
	if err := migrateToLinkedList(ctx, db); err != nil {
		return err
	}

	// Migrate columns to include project_id
	if err := migrateColumnsToProject(ctx, db); err != nil {
		return err
	}

	// Migrate tasks to include ticket_number
	if err := migrateTasksTicketNumber(ctx, db); err != nil {
		return err
	}

	// Migrate labels to include project_id and seed default labels
	if err := migrateLabelsToProject(ctx, db); err != nil {
		return err
	}

	// Migrate task_subtasks table for parent/child relationships
	if err := migrateTaskSubtasks(ctx, db); err != nil {
		return err
	}

	// Migrate types table and add type_id to tasks
	if err := migrateTaskTypes(ctx, db); err != nil {
		return err
	}

	// Migrate priorities table and add priority_id to tasks
	if err := migrateTaskPriorities(ctx, db); err != nil {
		return err
	}

	// Migrate relation_types table and add relation_type_id to task_subtasks
	if err := migrateRelationTypes(ctx, db); err != nil {
		return err
	}

	// Create indexes AFTER all table/column migrations are complete
	// Index on columns.project_id for efficient project-based queries
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_columns_project
		ON columns(project_id)
	`)
	if err != nil {
		return err
	}

	// Index on labels.project_id for efficient project-based label queries
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_labels_project
		ON labels(project_id)
	`)
	if err != nil {
		return err
	}

	// Index on task_labels.label_id for efficient label lookups
	_, err = db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_task_labels_label
		ON task_labels(label_id)
	`)
	if err != nil {
		return err
	}

	return nil
}

// seedDefaultColumns inserts default columns if the columns table is empty
func seedDefaultColumns(ctx context.Context, db *sql.DB) error {
	// Check if columns table is empty
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM columns").Scan(&count)
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
		_, err := db.ExecContext(ctx,
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
func migrateToLinkedList(ctx context.Context, db *sql.DB) error {
	// Check if prev_id column already exists
	var count int
	err := db.QueryRowContext(ctx, `
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
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// 1. Add new columns for linked list structure
	_, err = tx.ExecContext(ctx, `ALTER TABLE columns ADD COLUMN prev_id INTEGER NULL`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE columns ADD COLUMN next_id INTEGER NULL`)
	if err != nil {
		return err
	}

	// 2. Migrate existing data: query all columns ordered by position
	rows, err := tx.QueryContext(ctx, `SELECT id FROM columns ORDER BY position`)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var columnIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		columnIDs = append(columnIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

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
		_, err = tx.ExecContext(ctx, `
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
	_, err = tx.ExecContext(ctx, `
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

	_, err = tx.ExecContext(ctx, `
		INSERT INTO columns_new (id, name, prev_id, next_id)
		SELECT id, name, prev_id, next_id FROM columns
	`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DROP TABLE columns`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE columns_new RENAME TO columns`)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// seedDefaultProject creates a default project if no projects exist
func seedDefaultProject(ctx context.Context, db *sql.DB) error {
	// Check if projects table is empty
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects").Scan(&count)
	if err != nil {
		return err
	}

	// If projects exist, don't seed
	if count > 0 {
		return nil
	}

	// Insert default project
	result, err := db.ExecContext(ctx,
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

	_, err = db.ExecContext(ctx,
		`INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)`,
		projectID,
	)
	return err
}

// migrateColumnsToProject adds project_id column to columns table and associates existing columns with default project
func migrateColumnsToProject(ctx context.Context, db *sql.DB) error {
	// Check if project_id column already exists
	var count int
	err := db.QueryRowContext(ctx, `
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
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Get the default project ID
	var defaultProjectID int
	err = tx.QueryRowContext(ctx, `SELECT id FROM projects WHERE name = 'Default' LIMIT 1`).Scan(&defaultProjectID)
	if err != nil {
		return err
	}

	// Add project_id column
	_, err = tx.ExecContext(ctx, `ALTER TABLE columns ADD COLUMN project_id INTEGER NOT NULL DEFAULT 1`)
	if err != nil {
		return err
	}

	// Update all existing columns to belong to the default project
	_, err = tx.ExecContext(ctx, `UPDATE columns SET project_id = ?`, defaultProjectID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// migrateTasksTicketNumber adds ticket_number column to tasks table
func migrateTasksTicketNumber(ctx context.Context, db *sql.DB) error {
	// Check if ticket_number column already exists
	var count int
	err := db.QueryRowContext(ctx, `
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
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Add ticket_number column
	_, err = tx.ExecContext(ctx, `ALTER TABLE tasks ADD COLUMN ticket_number INTEGER`)
	if err != nil {
		return err
	}

	// Assign ticket numbers to existing tasks grouped by their project
	// First, get all tasks ordered by id (oldest first) with their project info
	rows, err := tx.QueryContext(ctx, `
		SELECT t.id, c.project_id
		FROM tasks t
		JOIN columns c ON t.column_id = c.id
		ORDER BY c.project_id, t.id
	`)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

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
			return err
		}

		// Get next ticket number for this project
		counter := projectCounters[projectID]
		counter++
		projectCounters[projectID] = counter

		updates = append(updates, taskUpdate{id: taskID, ticketNumber: counter})
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Apply updates
	for _, u := range updates {
		_, err = tx.ExecContext(ctx, `UPDATE tasks SET ticket_number = ? WHERE id = ?`, u.ticketNumber, u.id)
		if err != nil {
			return err
		}
	}

	// Update project counters to reflect the highest ticket number used
	for projectID, counter := range projectCounters {
		_, err = tx.ExecContext(ctx, `
			INSERT OR REPLACE INTO project_counters (project_id, next_ticket_number)
			VALUES (?, ?)
		`, projectID, counter+1)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// migrateLabelsToProject adds project_id column to labels table and seeds default labels
func migrateLabelsToProject(ctx context.Context, db *sql.DB) error {
	// Check if project_id column already exists
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pragma_table_info('labels')
		WHERE name = 'project_id'
	`).Scan(&count)
	if err != nil {
		return err
	}

	// If column doesn't exist, we need to migrate
	if count == 0 {
		// Start transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}()

		// Get the default project ID
		var defaultProjectID int
		err = tx.QueryRowContext(ctx, `SELECT id FROM projects WHERE name = 'Default' LIMIT 1`).Scan(&defaultProjectID)
		if err != nil {
			// If no default project, try to get the first project
			err = tx.QueryRowContext(ctx, `SELECT id FROM projects ORDER BY id LIMIT 1`).Scan(&defaultProjectID)
			if err != nil {
				// No projects exist yet, this will be handled when projects are created
				return nil
			}
		}

		// Add project_id column
		_, err = tx.ExecContext(ctx, `ALTER TABLE labels ADD COLUMN project_id INTEGER NOT NULL DEFAULT 1`)
		if err != nil {
			return err
		}

		// Update all existing labels to belong to the default project
		_, err = tx.ExecContext(ctx, `UPDATE labels SET project_id = ?`, defaultProjectID)
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	// Seed default labels for all projects that don't have labels yet
	return seedDefaultLabels(ctx, db)
}

// seedDefaultLabels seeds default GitHub-style labels for projects that don't have any labels
func seedDefaultLabels(ctx context.Context, db *sql.DB) error {
	// Default labels (GitHub-style)
	defaultLabels := []struct {
		name  string
		color string
	}{
		{"bug", "#EF4444"},         // Red
		{"duplicate", "#6B7280"},   // Gray
		{"enhancement", "#3B82F6"}, // Blue
		{"help wanted", "#22C55E"}, // Green
		{"invalid", "#6B7280"},     // Gray
		{"question", "#EC4899"},    // Pink/Magenta
	}

	// Get all projects
	rows, err := db.QueryContext(ctx, `SELECT id FROM projects`)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var projectIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		projectIDs = append(projectIDs, id)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// For each project, check if it has labels and seed if not
	for _, projectID := range projectIDs {
		var labelCount int
		err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM labels WHERE project_id = ?`, projectID).Scan(&labelCount)
		if err != nil {
			return err
		}

		// Only seed if project has no labels
		if labelCount == 0 {
			for _, label := range defaultLabels {
				_, err := db.ExecContext(ctx,
					`INSERT OR IGNORE INTO labels (name, color, project_id) VALUES (?, ?, ?)`,
					label.name, label.color, projectID,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// migrateTaskSubtasks creates the task_subtasks join table for parent/child task relationships
func migrateTaskSubtasks(ctx context.Context, db *sql.DB) error {
	// Check if task_subtasks table already exists
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='task_subtasks'
	`).Scan(&count)
	if err != nil {
		return err
	}

	// If table exists, skip migration
	if count > 0 {
		return nil
	}

	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Create task_subtasks join table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS task_subtasks (
			parent_id INTEGER NOT NULL,
			child_id INTEGER NOT NULL,
			PRIMARY KEY (parent_id, child_id),
			FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (child_id) REFERENCES tasks(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes for efficient lookups
	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_task_subtasks_parent
		ON task_subtasks(parent_id)
	`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_task_subtasks_child
		ON task_subtasks(child_id)
	`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// migrateTaskTypes creates the types table and adds type_id column to tasks
func migrateTaskTypes(ctx context.Context, db *sql.DB) error {
	// Check if types table already exists
	var tableCount int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='types'
	`).Scan(&tableCount)
	if err != nil {
		return err
	}

	// If table doesn't exist, create it
	if tableCount == 0 {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}()

		// Create types table
		_, err = tx.ExecContext(ctx, `
			CREATE TABLE types (
				id INTEGER PRIMARY KEY,
				description TEXT NOT NULL UNIQUE
			)
		`)
		if err != nil {
			return err
		}

		// Seed types table with task and feature
		_, err = tx.ExecContext(ctx, `
			INSERT INTO types (id, description) VALUES
				(1, 'task'),
				(2, 'feature')
		`)
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	// Check if type_id column already exists in tasks table
	var columnCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pragma_table_info('tasks')
		WHERE name = 'type_id'
	`).Scan(&columnCount)
	if err != nil {
		return err
	}

	// If column exists, skip migration
	if columnCount > 0 {
		return nil
	}

	// Start transaction to add type_id column
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Add type_id column with default value of 1 (task)
	_, err = tx.ExecContext(ctx, `ALTER TABLE tasks ADD COLUMN type_id INTEGER NOT NULL DEFAULT 1`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// migrateTaskPriorities creates the priorities table and adds priority_id column to tasks
func migrateTaskPriorities(ctx context.Context, db *sql.DB) error {
	// Check if priorities table already exists
	var tableCount int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='priorities'
	`).Scan(&tableCount)
	if err != nil {
		return err
	}

	// If table doesn't exist, create it
	if tableCount == 0 {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}()

		// Create priorities table with color mapping
		_, err = tx.ExecContext(ctx, `
			CREATE TABLE priorities (
				id INTEGER PRIMARY KEY,
				description TEXT NOT NULL UNIQUE,
				color TEXT NOT NULL
			)
		`)
		if err != nil {
			return err
		}

		// Seed priorities table with values and color mappings
		// Colors go from blue -> yellow -> orange -> red
		_, err = tx.ExecContext(ctx, `
			INSERT INTO priorities (id, description, color) VALUES
				(1, 'trivial', '#3B82F6'),
				(2, 'low', '#22C55E'),
				(3, 'medium', '#EAB308'),
				(4, 'high', '#F97316'),
				(5, 'critical', '#EF4444')
		`)
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	// Check if priority_id column already exists in tasks table
	var columnCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pragma_table_info('tasks')
		WHERE name = 'priority_id'
	`).Scan(&columnCount)
	if err != nil {
		return err
	}

	// If column exists, skip migration
	if columnCount > 0 {
		return nil
	}

	// Start transaction to add priority_id column
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Add priority_id column with default value of 3 (medium)
	_, err = tx.ExecContext(ctx, `ALTER TABLE tasks ADD COLUMN priority_id INTEGER NOT NULL DEFAULT 3`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// migrateRelationTypes creates the relation_types table and adds relation_type_id column to task_subtasks
func migrateRelationTypes(ctx context.Context, db *sql.DB) error {
	// Check if relation_types table already exists
	var tableCount int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='relation_types'
	`).Scan(&tableCount)
	if err != nil {
		return err
	}

	// If table doesn't exist, create it
	if tableCount == 0 {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}()

		// Create relation_types table
		_, err = tx.ExecContext(ctx, `
			CREATE TABLE relation_types (
				id INTEGER PRIMARY KEY,
				p_to_c_label TEXT NOT NULL,
				c_to_p_label TEXT NOT NULL,
				color TEXT NOT NULL,
				is_blocking BOOLEAN NOT NULL DEFAULT 0
			)
		`)
		if err != nil {
			return err
		}

		// Seed relation_types table with default values
		_, err = tx.ExecContext(ctx, `
			INSERT INTO relation_types (id, p_to_c_label, c_to_p_label, color, is_blocking) VALUES
				(1, 'Parent', 'Child', '#6B7280', 0),
				(2, 'Blocked By', 'Blocker', '#EF4444', 1),
				(3, 'Related To', 'Related To', '#3B82F6', 0)
		`)
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	// Check if relation_type_id column already exists in task_subtasks table
	var columnCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pragma_table_info('task_subtasks')
		WHERE name = 'relation_type_id'
	`).Scan(&columnCount)
	if err != nil {
		return err
	}

	// If column exists, skip migration
	if columnCount > 0 {
		return nil
	}

	// Start transaction to add relation_type_id column
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Add relation_type_id column with default value of 1 (Parent/Child)
	_, err = tx.ExecContext(ctx, `ALTER TABLE task_subtasks ADD COLUMN relation_type_id INTEGER NOT NULL DEFAULT 1`)
	if err != nil {
		return err
	}

	return tx.Commit()
}
