package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
	"github.com/thenoetrevino/paso/internal/database/generated"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// runMigrations runs goose migrations and seeds default data
func runMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	// Seed default data after migrations
	return seedDefaultData(ctx, db)
}

// seedDefaultData seeds default project, columns, and labels if needed
func seedDefaultData(ctx context.Context, db *sql.DB) error {
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
		Name:                "Todo",
		ProjectID:           projectID,
		PrevID:              nil,
		NextID:              nil,
		HoldsReadyTasks:     true,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create Todo column: %w", err)
	}

	// Create "In Progress" column (middle of list)
	inProgressCol, err := q.CreateColumn(ctx, generated.CreateColumnParams{
		Name:                "In Progress",
		ProjectID:           projectID,
		PrevID:              todoCol.ID,
		NextID:              nil,
		HoldsReadyTasks:     false,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create In Progress column: %w", err)
	}

	// Create "Done" column (tail of list, holds completed tasks)
	doneCol, err := q.CreateColumn(ctx, generated.CreateColumnParams{
		Name:                "Done",
		ProjectID:           projectID,
		PrevID:              inProgressCol.ID,
		NextID:              nil,
		HoldsReadyTasks:     false,
		HoldsCompletedTasks: true,
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
