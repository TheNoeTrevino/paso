package project

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// Service defines all project-related business operations
type Service interface {
	// Read operations
	GetAllProjects(ctx context.Context) ([]*models.Project, error)
	GetProjectByID(ctx context.Context, id int) (*models.Project, error)
	GetTaskCount(ctx context.Context, projectID int) (int, error)

	// Write operations
	CreateProject(ctx context.Context, req CreateProjectRequest) (*models.Project, error)
	UpdateProject(ctx context.Context, req UpdateProjectRequest) error
	DeleteProject(ctx context.Context, id int, force bool) error
}

// CreateProjectRequest encapsulates data for creating a project
type CreateProjectRequest struct {
	Name        string
	Description string
}

// UpdateProjectRequest encapsulates data for updating a project
type UpdateProjectRequest struct {
	ID          int
	Name        *string
	Description *string
}

// service implements Service interface using SQLC directly
type service struct {
	db          *sql.DB
	queries     generated.Querier
	eventClient events.EventPublisher
}

// NewService creates a new project service
func NewService(db *sql.DB, eventClient events.EventPublisher) Service {
	return &service{
		db:          db,
		queries:     generated.New(db),
		eventClient: eventClient,
	}
}

// GetAllProjects retrieves all projects
func (s *service) GetAllProjects(ctx context.Context) ([]*models.Project, error) {
	projects, err := s.queries.GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}
	return toProjectModels(projects), nil
}

// GetProjectByID retrieves a specific project
func (s *service) GetProjectByID(ctx context.Context, id int) (*models.Project, error) {
	if id <= 0 {
		return nil, ErrInvalidProjectID
	}
	project, err := s.queries.GetProjectByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return toProjectModel(project), nil
}

// GetTaskCount returns the number of tasks in a project
func (s *service) GetTaskCount(ctx context.Context, projectID int) (int, error) {
	if projectID <= 0 {
		return 0, ErrInvalidProjectID
	}
	count, err := s.queries.GetProjectTaskCount(ctx, int64(projectID))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// CreateProject creates a new project with validation
func (s *service) CreateProject(ctx context.Context, req CreateProjectRequest) (*models.Project, error) {
	// Validate request
	if err := s.validateCreateProject(req); err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	qtx := generated.New(tx)

	// Create project record
	project, err := qtx.CreateProjectRecord(ctx, generated.CreateProjectRecordParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Initialize project counter (for task ticket numbers)
	if err := qtx.InitializeProjectCounter(ctx, project.ID); err != nil {
		return nil, fmt.Errorf("failed to initialize project counter: %w", err)
	}

	// Create default columns (Todo, In Progress, Done)
	if err := database.CreateDefaultColumns(ctx, qtx, project.ID); err != nil {
		return nil, fmt.Errorf("failed to create default columns: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event after successful commit
	s.publishProjectEvent(ctx, int(project.ID))

	return toProjectModel(project), nil
}

// UpdateProject updates an existing project
func (s *service) UpdateProject(ctx context.Context, req UpdateProjectRequest) error {
	// Validate project ID
	if req.ID <= 0 {
		return ErrInvalidProjectID
	}

	// Validate fields if provided
	if req.Name != nil && *req.Name == "" {
		return ErrEmptyName
	}
	if req.Name != nil && len(*req.Name) > 100 {
		return ErrNameTooLong
	}

	// Get existing project to fill in missing fields
	existing, err := s.queries.GetProjectByID(ctx, int64(req.ID))
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Determine final values
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	description := existing.Description
	if req.Description != nil {
		description = sql.NullString{String: *req.Description, Valid: *req.Description != ""}
	}

	// Update project
	if err := s.queries.UpdateProject(ctx, generated.UpdateProjectParams{
		ID:          int64(req.ID),
		Name:        name,
		Description: description,
	}); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Publish event
	s.publishProjectEvent(ctx, req.ID)

	return nil
}

// DeleteProject deletes a project (business rule: must not have tasks unless force=true)
func (s *service) DeleteProject(ctx context.Context, id int, force bool) error {
	if id <= 0 {
		return ErrInvalidProjectID
	}

	// Business rule: Check if project has tasks (unless force is enabled)
	if !force {
		taskCount, err := s.queries.GetProjectTaskCount(ctx, int64(id))
		if err != nil {
			return fmt.Errorf("failed to check project tasks: %w", err)
		}
		if taskCount > 0 {
			return ErrProjectHasTasks
		}
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	qtx := generated.New(tx)

	// Delete tasks first
	if err := qtx.DeleteTasksByProject(ctx, int64(id)); err != nil {
		return fmt.Errorf("failed to delete tasks: %w", err)
	}

	// Delete columns
	if err := qtx.DeleteColumnsByProject(ctx, int64(id)); err != nil {
		return fmt.Errorf("failed to delete columns: %w", err)
	}

	// Delete project counter
	if err := qtx.DeleteProjectCounter(ctx, int64(id)); err != nil {
		return fmt.Errorf("failed to delete counter: %w", err)
	}

	// Delete project
	if err := qtx.DeleteProject(ctx, int64(id)); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event after successful deletion
	s.publishProjectEvent(ctx, id)

	return nil
}

// validateCreateProject validates a CreateProjectRequest
func (s *service) validateCreateProject(req CreateProjectRequest) error {
	if req.Name == "" {
		return ErrEmptyName
	}
	if len(req.Name) > 100 {
		return ErrNameTooLong
	}
	return nil
}

// publishProjectEvent publishes a project event with retry logic
func (s *service) publishProjectEvent(ctx context.Context, projectID int) {
	if s.eventClient == nil {
		return
	}

	// Publish with retry (3 attempts with exponential backoff)
	// Non-blocking: errors are logged but don't affect the operation
	_ = events.PublishWithRetry(s.eventClient, events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: projectID,
	}, 3)
}

// Model conversion helpers

func toProjectModel(p generated.Project) *models.Project {
	return &models.Project{
		ID:          int(p.ID),
		Name:        p.Name,
		Description: database.NullStringToString(p.Description),
		CreatedAt:   database.NullTimeToTime(p.CreatedAt),
		UpdatedAt:   database.NullTimeToTime(p.UpdatedAt),
	}
}

func toProjectModels(projects []generated.Project) []*models.Project {
	result := make([]*models.Project, len(projects))
	for i, p := range projects {
		result[i] = toProjectModel(p)
	}
	return result
}
