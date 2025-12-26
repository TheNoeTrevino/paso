package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/thenoetrevino/paso/internal/database/generated"
)

// runMigrations creates the database schema and seeds default data if needed
func runMigrations(ctx context.Context, db *sql.DB) error {
	// 1. Create and seed lookup tables (no dependencies)
	if err := createLookupTables(ctx, db); err != nil {
		return err
	}

	// 2. Create core tables
	if err := createCoreTables(ctx, db); err != nil {
		return err
	}

	// 3. Create tasks table (depends on columns, types, priorities)
	if err := createTasksTable(ctx, db); err != nil {
		return err
	}

	// 4. Create join tables
	if err := createJoinTables(ctx, db); err != nil {
		return err
	}

	// 5. Create indexes
	if err := createIndexes(ctx, db); err != nil {
		return err
	}

	// 6. Seed default data
	if err := seedDefaultProject(ctx, db); err != nil {
		return err
	}

	if err := seedDefaultColumns(ctx, db); err != nil {
		return err
	}

	if err := seedDefaultLabels(ctx, db); err != nil {
		return err
	}

	return nil
}

// createLookupTables creates and seeds the lookup tables (types, priorities, relation_types)
func createLookupTables(ctx context.Context, db *sql.DB) error {
	// Create types table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS types (
			id INTEGER PRIMARY KEY,
			description TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return err
	}

	// Seed types table
	_, err = db.ExecContext(ctx, `
		INSERT OR IGNORE INTO types (id, description) VALUES
			(1, 'task'),
			(2, 'feature')
	`)
	if err != nil {
		return err
	}

	// Create priorities table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS priorities (
			id INTEGER PRIMARY KEY,
			description TEXT NOT NULL UNIQUE,
			color TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Seed priorities table
	_, err = db.ExecContext(ctx, `
		INSERT OR IGNORE INTO priorities (id, description, color) VALUES
			(1, 'trivial', '#3B82F6'),
			(2, 'low', '#22C55E'),
			(3, 'medium', '#EAB308'),
			(4, 'high', '#F97316'),
			(5, 'critical', '#EF4444')
	`)
	if err != nil {
		return err
	}

	// Create relation_types table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS relation_types (
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

	// Seed relation_types table
	_, err = db.ExecContext(ctx, `
		INSERT OR IGNORE INTO relation_types (id, p_to_c_label, c_to_p_label, color, is_blocking) VALUES
			(1, 'Parent', 'Child', '#6B7280', 0),
			(2, 'Blocked By', 'Blocker', '#EF4444', 1),
			(3, 'Related To', 'Related To', '#3B82F6', 0)
	`)
	if err != nil {
		return err
	}

	return nil
}

// createCoreTables creates projects, project_counters, columns, and labels tables
func createCoreTables(ctx context.Context, db *sql.DB) error {
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

	// Create project_counters table
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

	// Create columns table (linked list structure)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS columns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			prev_id INTEGER NULL,
			next_id INTEGER NULL,
			project_id INTEGER NOT NULL,
			holds_ready_tasks BOOLEAN NOT NULL DEFAULT 0,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create labels table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			color TEXT NOT NULL DEFAULT '#7D56F4',
			project_id INTEGER NOT NULL,
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
			UNIQUE(name, project_id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// createTasksTable creates the tasks table with all foreign key constraints
func createTasksTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			column_id INTEGER NOT NULL,
			position INTEGER NOT NULL,
			ticket_number INTEGER,
			type_id INTEGER NOT NULL DEFAULT 1,
			priority_id INTEGER NOT NULL DEFAULT 3,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE,
			FOREIGN KEY (type_id) REFERENCES types(id),
			FOREIGN KEY (priority_id) REFERENCES priorities(id)
		)
	`)
	return err
}

// createJoinTables creates task_labels and task_subtasks join tables
func createJoinTables(ctx context.Context, db *sql.DB) error {
	// Create task_labels table
	_, err := db.ExecContext(ctx, `
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

	// Create task_subtasks table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS task_subtasks (
			parent_id INTEGER NOT NULL,
			child_id INTEGER NOT NULL,
			relation_type_id INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (parent_id, child_id),
			FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (child_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (relation_type_id) REFERENCES relation_types(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// createIndexes creates all necessary indexes for performance
func createIndexes(ctx context.Context, db *sql.DB) error {
	indexes := []string{
		// Index for efficient task queries by column
		`CREATE INDEX IF NOT EXISTS idx_tasks_column ON tasks(column_id, position)`,

		// Index for efficient project-based column queries
		`CREATE INDEX IF NOT EXISTS idx_columns_project ON columns(project_id)`,

		// Index for efficient project-based label queries
		`CREATE INDEX IF NOT EXISTS idx_labels_project ON labels(project_id)`,

		// Index for efficient label lookups in task_labels
		`CREATE INDEX IF NOT EXISTS idx_task_labels_label ON task_labels(label_id)`,

		// Index for efficient parent task lookups
		`CREATE INDEX IF NOT EXISTS idx_task_subtasks_parent ON task_subtasks(parent_id)`,

		// Index for efficient child task lookups
		`CREATE INDEX IF NOT EXISTS idx_task_subtasks_child ON task_subtasks(child_id)`,
	}

	for _, indexSQL := range indexes {
		_, err := db.ExecContext(ctx, indexSQL)
		if err != nil {
			return err
		}
	}

	return nil
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

// CreateDefaultColumns creates the standard three columns (Todo, In Progress, Done)
// for a given project using the provided querier (works with both db and tx)
func CreateDefaultColumns(ctx context.Context, q generated.Querier, projectID int64) error {
	// Create "Todo" column (head of list, holds ready tasks)
	todoCol, err := q.CreateColumn(ctx, generated.CreateColumnParams{
		Name:            "Todo",
		ProjectID:       projectID,
		PrevID:          nil,
		NextID:          nil,
		HoldsReadyTasks: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create Todo column: %w", err)
	}

	// Create "In Progress" column (middle of list)
	inProgressCol, err := q.CreateColumn(ctx, generated.CreateColumnParams{
		Name:      "In Progress",
		ProjectID: projectID,
		PrevID:    todoCol.ID,
		NextID:    nil,
	})
	if err != nil {
		return fmt.Errorf("failed to create In Progress column: %w", err)
	}

	// Create "Done" column (tail of list)
	doneCol, err := q.CreateColumn(ctx, generated.CreateColumnParams{
		Name:      "Done",
		ProjectID: projectID,
		PrevID:    inProgressCol.ID,
		NextID:    nil,
	})
	if err != nil {
		return fmt.Errorf("failed to create Done column: %w", err)
	}

	// Update next_id pointers to complete the linked list
	if err := q.UpdateColumnNextID(ctx, generated.UpdateColumnNextIDParams{
		ID:     todoCol.ID,
		NextID: inProgressCol.ID,
	}); err != nil {
		return fmt.Errorf("failed to update Todo next_id: %w", err)
	}

	if err := q.UpdateColumnNextID(ctx, generated.UpdateColumnNextIDParams{
		ID:     inProgressCol.ID,
		NextID: doneCol.ID,
	}); err != nil {
		return fmt.Errorf("failed to update In Progress next_id: %w", err)
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

	// Get the default project ID
	var defaultProjectID int
	err = db.QueryRowContext(ctx, `SELECT id FROM projects WHERE name = 'Default' LIMIT 1`).Scan(&defaultProjectID)
	if err != nil {
		return err
	}

	// Use the shared helper to create default columns
	q := generated.New(db)
	return CreateDefaultColumns(ctx, q, int64(defaultProjectID))
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
