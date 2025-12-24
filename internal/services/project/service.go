package project

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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
	DeleteProject(ctx context.Context, id int) error
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

// repository defines the data access methods needed by the project service
// This interface is private to the service layer
type repository interface {
	// Project operations
	CreateProjectRecord(ctx context.Context, name, description string) (*models.Project, error)
	GetProjectByID(ctx context.Context, id int) (*models.Project, error)
	GetAllProjects(ctx context.Context) ([]*models.Project, error)
	UpdateProject(ctx context.Context, id int, name, description string) error
	DeleteProject(ctx context.Context, id int) error

	// Project-related operations
	InitializeProjectCounter(ctx context.Context, projectID int) error
	DeleteProjectCounter(ctx context.Context, projectID int) error
	DeleteTasksByProject(ctx context.Context, projectID int) error
	DeleteColumnsByProject(ctx context.Context, projectID int) error
	GetProjectTaskCount(ctx context.Context, projectID int) (int, error)

	// Transaction support
	BeginTx(ctx context.Context) (*sql.Tx, error)
	WithTx(tx *sql.Tx) repository
}

// columnRepository is needed to check if project has columns (business rule)
type columnRepository interface {
	GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error)
}

// service implements Service interface with private repository
type service struct {
	repo        repository
	columnRepo  columnRepository
	eventClient events.EventPublisher
}

// NewService creates a new project service with private repository
func NewService(repo repository, columnRepo columnRepository, eventClient events.EventPublisher) Service {
	return &service{
		repo:        repo,
		columnRepo:  columnRepo,
		eventClient: eventClient,
	}
}

// GetAllProjects retrieves all projects
func (s *service) GetAllProjects(ctx context.Context) ([]*models.Project, error) {
	return s.repo.GetAllProjects(ctx)
}

// GetProjectByID retrieves a specific project
func (s *service) GetProjectByID(ctx context.Context, id int) (*models.Project, error) {
	if id <= 0 {
		return nil, ErrInvalidProjectID
	}
	return s.repo.GetProjectByID(ctx, id)
}

// GetTaskCount returns the number of tasks in a project
func (s *service) GetTaskCount(ctx context.Context, projectID int) (int, error) {
	if projectID <= 0 {
		return 0, ErrInvalidProjectID
	}
	return s.repo.GetProjectTaskCount(ctx, projectID)
}

// CreateProject creates a new project with validation
func (s *service) CreateProject(ctx context.Context, req CreateProjectRequest) (*models.Project, error) {
	// Validate request
	if err := s.validateCreateProject(req); err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	repoTx := s.repo.WithTx(tx)

	// Create project record
	project, err := repoTx.CreateProjectRecord(ctx, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Initialize project counter (for task ticket numbers)
	if err := repoTx.InitializeProjectCounter(ctx, project.ID); err != nil {
		return nil, fmt.Errorf("failed to initialize project counter: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event after successful commit
	s.publishProjectEvent(project.ID)

	return project, nil
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
	existing, err := s.repo.GetProjectByID(ctx, req.ID)
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
		description = *req.Description
	}

	// Update project
	if err := s.repo.UpdateProject(ctx, req.ID, name, description); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	// Publish event after successful update
	s.publishProjectEvent(req.ID)

	return nil
}

// DeleteProject deletes a project (business rule: must not have columns)
func (s *service) DeleteProject(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidProjectID
	}

	// Business rule: Check if project has columns
	columns, err := s.columnRepo.GetColumnsByProject(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check project columns: %w", err)
	}
	if len(columns) > 0 {
		return ErrProjectHasColumns
	}

	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	repoTx := s.repo.WithTx(tx)

	// Delete tasks first
	if err := repoTx.DeleteTasksByProject(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tasks: %w", err)
	}

	// Delete columns
	if err := repoTx.DeleteColumnsByProject(ctx, id); err != nil {
		return fmt.Errorf("failed to delete columns: %w", err)
	}

	// Delete project counter
	if err := repoTx.DeleteProjectCounter(ctx, id); err != nil {
		return fmt.Errorf("failed to delete counter: %w", err)
	}

	// Delete project
	if err := repoTx.DeleteProject(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event after successful deletion
	s.publishProjectEvent(id)

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

// publishProjectEvent publishes a project event
func (s *service) publishProjectEvent(projectID int) {
	if s.eventClient == nil {
		return
	}

	if err := s.eventClient.SendEvent(events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: projectID,
	}); err != nil {
		log.Printf("failed to send event for project %d: %v", projectID, err)
	}
}
