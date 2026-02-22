package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sync"

	"github.com/pressly/goose/v3"
	"github.com/thenoetrevino/paso/internal/database/types"
)

// Type aliases for convenience in CreateDefaultColumns
type (
	CreateColumnParams       = types.CreateColumnParams
	UpdateColumnNextIDParams = types.UpdateColumnNextIDParams
	NullInt64                = types.NullInt64
)

//go:embed migrations_sqlite/*.sql migrations_postgres/*.sql
var embedMigrations embed.FS

// gooseMu serializes access to goose's package-level global state (SetBaseFS,
// SetDialect). This is only relevant when parallel tests each call
// applyMigrations on their own in-memory DB; production code calls it once
// during startup.
var gooseMu sync.Mutex

// RunMigrationsOnly applies goose schema migrations without seeding default data.
// This is intended for test setup where tests need a clean schema without seed data.
func RunMigrationsOnly(db *sql.DB, dbType DatabaseType) error {
	return applyMigrations(db, dbType)
}

// applyMigrations runs goose migrations for the appropriate database type
func applyMigrations(db *sql.DB, dbType DatabaseType) error {
	gooseMu.Lock()
	defer gooseMu.Unlock()

	goose.SetBaseFS(embedMigrations)

	switch dbType {
	case SQLite:
		if err := goose.SetDialect("sqlite3"); err != nil {
			return fmt.Errorf("failed to set goose dialect: %w", err)
		}
		if err := goose.Up(db, "migrations_sqlite"); err != nil {
			return fmt.Errorf("failed to run goose migrations: %w", err)
		}
	case PostgreSQL:
		if err := goose.SetDialect("postgres"); err != nil {
			return fmt.Errorf("failed to set goose dialect: %w", err)
		}
		if err := goose.Up(db, "migrations_postgres"); err != nil {
			return fmt.Errorf("failed to run goose migrations: %w", err)
		}
	default:
		return fmt.Errorf("unknown database type: %s", dbType)
	}

	return nil
}

// runMigrations runs goose migrations for the appropriate database type and seeds default data
func runMigrations(ctx context.Context, db *sql.DB, dbType DatabaseType) error {
	slog.Info("running database migrations", "type", dbType)
	if err := applyMigrations(db, dbType); err != nil {
		slog.Error("migrations failed", "error", err)
		return err
	}
	slog.Info("migrations completed successfully")

	slog.Info("seeding default data")
	if err := seedDefaultData(ctx, db, dbType); err != nil {
		slog.Error("failed to seed default data", "error", err)
		return err
	}
	slog.Info("default data seeded successfully")
	return nil
}

// seedDefaultData seeds default project, columns, and labels if needed
func seedDefaultData(ctx context.Context, db *sql.DB, dbType DatabaseType) error {
	slog.Debug("seeding default project")
	if err := seedDefaultProject(ctx, db, dbType); err != nil {
		slog.Error("failed to seed default project", "error", err)
		return err
	}

	slog.Debug("seeding default columns")
	if err := seedDefaultColumns(ctx, db, dbType); err != nil {
		slog.Error("failed to seed default columns", "error", err)
		return err
	}

	slog.Debug("seeding default labels")
	if err := seedDefaultLabels(ctx, db, dbType); err != nil {
		slog.Error("failed to seed default labels", "error", err)
		return err
	}

	return nil
}

// seedDefaultProject creates a default project if no projects exist
func seedDefaultProject(ctx context.Context, db *sql.DB, dbType DatabaseType) error {
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

	// Insert default project with database-specific SQL
	var projectID int64

	switch dbType {
	case PostgreSQL:
		// PostgreSQL: use RETURNING to get the ID
		insertQuery := `INSERT INTO projects (name, description) VALUES ($1, $2) RETURNING id`
		err = db.QueryRowContext(ctx, insertQuery, "Default", "Default project").Scan(&projectID)
	case SQLite:
		// SQLite: use LastInsertId
		insertQuery := `INSERT INTO projects (name, description) VALUES (?, ?)`
		result, err := db.ExecContext(ctx, insertQuery, "Default", "Default project")
		if err != nil {
			return err
		}
		projectID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	// Use generated query for project counter
	queries, err := NewQuerier(db, dbType)
	if err != nil {
		return err
	}
	return queries.InitializeProjectCounter(ctx, projectID)
}

// CreateDefaultColumns creates the standard three columns (Todo, In Progress, Done)
// for a given project using the provided database-agnostic querier
func CreateDefaultColumns(ctx context.Context, q Querier, projectID int64) error {
	// Create "Todo" column (head of list, holds ready tasks)
	todoCol, err := q.CreateColumn(ctx, CreateColumnParams{
		Name:                "Todo",
		ProjectID:           projectID,
		PrevID:              NullInt64{Valid: false},
		NextID:              NullInt64{Valid: false},
		HoldsReadyTasks:     true,
		HoldsCompletedTasks: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create Todo column: %w", err)
	}

	// Create "In Progress" column (middle of list, holds in-progress tasks)
	inProgressCol, err := q.CreateColumn(ctx, CreateColumnParams{
		Name:                 "In Progress",
		ProjectID:            projectID,
		PrevID:               NullInt64{Int64: todoCol.ID, Valid: true},
		NextID:               NullInt64{Valid: false},
		HoldsReadyTasks:      false,
		HoldsCompletedTasks:  false,
		HoldsInProgressTasks: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create In Progress column: %w", err)
	}

	// Create "Done" column (tail of list, holds completed tasks)
	doneCol, err := q.CreateColumn(ctx, CreateColumnParams{
		Name:                "Done",
		ProjectID:           projectID,
		PrevID:              NullInt64{Int64: inProgressCol.ID, Valid: true},
		NextID:              NullInt64{Valid: false},
		HoldsReadyTasks:     false,
		HoldsCompletedTasks: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create Done column: %w", err)
	}

	// Update next_id pointers to complete the linked list
	if err := q.UpdateColumnNextID(ctx, UpdateColumnNextIDParams{
		ID:     todoCol.ID,
		NextID: NullInt64{Int64: inProgressCol.ID, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to update Todo next_id: %w", err)
	}

	if err := q.UpdateColumnNextID(ctx, UpdateColumnNextIDParams{
		ID:     inProgressCol.ID,
		NextID: NullInt64{Int64: doneCol.ID, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to update In Progress next_id: %w", err)
	}

	return nil
}

// seedDefaultColumns inserts default columns if the columns table is empty
func seedDefaultColumns(ctx context.Context, db *sql.DB, dbType DatabaseType) error {
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

	// Use the database-agnostic querier
	q, err := NewQuerier(db, dbType)
	if err != nil {
		return fmt.Errorf("failed to create querier for seeding columns: %w", err)
	}
	return CreateDefaultColumns(ctx, q, int64(defaultProjectID))
}

// seedDefaultLabels seeds default GitHub-style labels for projects that don't have any labels
func seedDefaultLabels(ctx context.Context, db *sql.DB, dbType DatabaseType) error {
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
			slog.Error("failed to close rows", "error", err)
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

	// Use generated queries for labels
	queries, err := NewQuerier(db, dbType)
	if err != nil {
		return err
	}

	// For each project, check if it has labels and seed if not
	for _, projectID := range projectIDs {
		labelCount, err := queries.GetLabelCountByProject(ctx, int64(projectID))
		if err != nil {
			return err
		}

		// Only seed if project has no labels
		if labelCount == 0 {
			for _, label := range defaultLabels {
				if err := queries.UpsertLabel(ctx, types.UpsertLabelParams{
					Name:      label.name,
					Color:     label.color,
					ProjectID: int64(projectID),
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
