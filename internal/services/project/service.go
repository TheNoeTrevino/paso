package project

import (
	"context"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// Service defines all project-related business operations
type Service interface {
	// Read operations
	GetAllProjects(ctx context.Context) ([]*models.Project, error)
	GetProjectByID(ctx context.Context, id int) (*models.Project, error)

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

// service implements Service interface
type service struct {
	repo        database.DataStore
	eventClient events.EventPublisher
}

// NewService creates a new project service
func NewService(repo database.DataStore, eventClient events.EventPublisher) Service {
	return &service{
		repo:        repo,
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

// CreateProject creates a new project with validation
func (s *service) CreateProject(ctx context.Context, req CreateProjectRequest) (*models.Project, error) {
	// Validate request
	if err := s.validateCreateProject(req); err != nil {
		return nil, err
	}

	// Create project in repository
	project, err := s.repo.CreateProject(ctx, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Publish event
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

	// Publish event
	s.publishProjectEvent(req.ID)

	return nil
}

// DeleteProject deletes a project
func (s *service) DeleteProject(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidProjectID
	}

	// Check if project has columns (business rule)
	columns, err := s.repo.GetColumnsByProject(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check project columns: %w", err)
	}
	if len(columns) > 0 {
		return ErrProjectHasColumns
	}

	// Delete project
	if err := s.repo.DeleteProject(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Publish event
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

// publishProjectEvent publishes a project event if event client exists
func (s *service) publishProjectEvent(projectID int) {
	if s.eventClient == nil {
		return
	}

	// Publish database changed event
	_ = projectID // Used for future enhancement
}
