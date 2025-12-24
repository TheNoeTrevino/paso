package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// projectRepository handles pure data access for projects
// No business logic, no events, no validation - just database operations
type projectRepository struct {
	queries *generated.Queries
	db      *sql.DB
}

// newProjectRepository creates a new project repository
func newProjectRepository(queries *generated.Queries, db *sql.DB) *projectRepository {
	return &projectRepository{
		queries: queries,
		db:      db,
	}
}

// CreateProjectRecord inserts a project record (without columns/counters)
func (r *projectRepository) CreateProjectRecord(ctx context.Context, name, description string) (*models.Project, error) {
	row, err := r.queries.CreateProjectRecord(ctx, generated.CreateProjectRecordParams{
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return toProjectModel(row), nil
}

// GetProjectByID retrieves a project by ID
func (r *projectRepository) GetProjectByID(ctx context.Context, id int) (*models.Project, error) {
	row, err := r.queries.GetProjectByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d: %w", id, err)
	}
	return toProjectModel(row), nil
}

// GetAllProjects retrieves all projects
func (r *projectRepository) GetAllProjects(ctx context.Context) ([]*models.Project, error) {
	rows, err := r.queries.GetAllProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all projects: %w", err)
	}

	projects := make([]*models.Project, len(rows))
	for i, row := range rows {
		projects[i] = toProjectModel(row)
	}
	return projects, nil
}

// UpdateProject updates a project's name and description
func (r *projectRepository) UpdateProject(ctx context.Context, id int, name, description string) error {
	err := r.queries.UpdateProject(ctx, generated.UpdateProjectParams{
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		ID:          int64(id),
	})
	if err != nil {
		return fmt.Errorf("failed to update project %d: %w", id, err)
	}
	return nil
}

// DeleteProject removes a project
func (r *projectRepository) DeleteProject(ctx context.Context, id int) error {
	err := r.queries.DeleteProject(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("failed to delete project %d: %w", id, err)
	}
	return nil
}

// InitializeProjectCounter creates the counter for a project
func (r *projectRepository) InitializeProjectCounter(ctx context.Context, projectID int) error {
	err := r.queries.InitializeProjectCounter(ctx, int64(projectID))
	if err != nil {
		return fmt.Errorf("failed to initialize counter for project %d: %w", projectID, err)
	}
	return nil
}

// DeleteProjectCounter removes the counter for a project
func (r *projectRepository) DeleteProjectCounter(ctx context.Context, projectID int) error {
	err := r.queries.DeleteProjectCounter(ctx, int64(projectID))
	if err != nil {
		return fmt.Errorf("failed to delete counter for project %d: %w", projectID, err)
	}
	return nil
}

// DeleteTasksByProject removes all tasks for a project
func (r *projectRepository) DeleteTasksByProject(ctx context.Context, projectID int) error {
	err := r.queries.DeleteTasksByProject(ctx, int64(projectID))
	if err != nil {
		return fmt.Errorf("failed to delete tasks for project %d: %w", projectID, err)
	}
	return nil
}

// DeleteColumnsByProject removes all columns for a project
func (r *projectRepository) DeleteColumnsByProject(ctx context.Context, projectID int) error {
	err := r.queries.DeleteColumnsByProject(ctx, int64(projectID))
	if err != nil {
		return fmt.Errorf("failed to delete columns for project %d: %w", projectID, err)
	}
	return nil
}

// GetProjectTaskCount returns the number of tasks in a project
func (r *projectRepository) GetProjectTaskCount(ctx context.Context, projectID int) (int, error) {
	count, err := r.queries.GetProjectTaskCount(ctx, int64(projectID))
	if err != nil {
		return 0, fmt.Errorf("failed to get task count for project %d: %w", projectID, err)
	}
	return int(count), nil
}

// WithTx returns a new repository instance that uses the given transaction
func (r *projectRepository) WithTx(tx *sql.Tx) *projectRepository {
	return &projectRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

// BeginTx starts a new transaction
func (r *projectRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// ============================================================================
// MODEL CONVERSION HELPERS
// ============================================================================

func toProjectModel(row generated.Project) *models.Project {
	return &models.Project{
		ID:          int(row.ID),
		Name:        row.Name,
		Description: nullStringToString(row.Description),
		CreatedAt:   nullTimeToTime(row.CreatedAt),
		UpdatedAt:   nullTimeToTime(row.UpdatedAt),
	}
}
