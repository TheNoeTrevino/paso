package column

import (
	"context"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// Service defines all column-related business operations
type Service interface {
	// Read operations
	GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error)
	GetColumnByID(ctx context.Context, id int) (*models.Column, error)

	// Write operations
	CreateColumn(ctx context.Context, req CreateColumnRequest) (*models.Column, error)
	UpdateColumnName(ctx context.Context, id int, name string) error
	DeleteColumn(ctx context.Context, id int) error
}

// CreateColumnRequest encapsulates data for creating a column
type CreateColumnRequest struct {
	Name      string
	ProjectID int
	AfterID   *int // Optional: ID of column to insert after (nil = append to end)
}

// service implements Service interface
type service struct {
	repo        database.DataStore
	eventClient events.EventPublisher
}

// NewService creates a new column service
func NewService(repo database.DataStore, eventClient events.EventPublisher) Service {
	return &service{
		repo:        repo,
		eventClient: eventClient,
	}
}

// GetColumnsByProject retrieves all columns for a project
func (s *service) GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	return s.repo.GetColumnsByProject(ctx, projectID)
}

// GetColumnByID retrieves a specific column
func (s *service) GetColumnByID(ctx context.Context, id int) (*models.Column, error) {
	if id <= 0 {
		return nil, ErrInvalidColumnID
	}
	return s.repo.GetColumnByID(ctx, id)
}

// CreateColumn creates a new column with validation
func (s *service) CreateColumn(ctx context.Context, req CreateColumnRequest) (*models.Column, error) {
	// Validate request
	if err := s.validateCreateColumn(req); err != nil {
		return nil, err
	}

	// Create column in repository
	column, err := s.repo.CreateColumn(ctx, req.Name, req.ProjectID, req.AfterID)
	if err != nil {
		return nil, fmt.Errorf("failed to create column: %w", err)
	}

	// Publish event
	s.publishColumnEvent(column.ID, column.ProjectID)

	return column, nil
}

// UpdateColumnName updates a column's name
func (s *service) UpdateColumnName(ctx context.Context, id int, name string) error {
	// Validate
	if id <= 0 {
		return ErrInvalidColumnID
	}
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 50 {
		return ErrNameTooLong
	}

	// Get column to find project ID for event
	column, err := s.repo.GetColumnByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get column: %w", err)
	}

	// Update column
	if err := s.repo.UpdateColumnName(ctx, id, name); err != nil {
		return fmt.Errorf("failed to update column: %w", err)
	}

	// Publish event
	s.publishColumnEvent(id, column.ProjectID)

	return nil
}

// DeleteColumn deletes a column
func (s *service) DeleteColumn(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidColumnID
	}

	// Check if column has tasks (business rule)
	tasks, err := s.repo.GetTasksByColumn(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check column tasks: %w", err)
	}
	if len(tasks) > 0 {
		return ErrColumnHasTasks
	}

	// Get column to find project ID for event
	column, err := s.repo.GetColumnByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get column: %w", err)
	}

	// Delete column
	if err := s.repo.DeleteColumn(ctx, id); err != nil {
		return fmt.Errorf("failed to delete column: %w", err)
	}

	// Publish event
	s.publishColumnEvent(id, column.ProjectID)

	return nil
}

// validateCreateColumn validates a CreateColumnRequest
func (s *service) validateCreateColumn(req CreateColumnRequest) error {
	if req.Name == "" {
		return ErrEmptyName
	}
	if len(req.Name) > 50 {
		return ErrNameTooLong
	}
	if req.ProjectID <= 0 {
		return ErrInvalidProjectID
	}
	if req.AfterID != nil && *req.AfterID <= 0 {
		return ErrInvalidColumnID
	}
	return nil
}

// publishColumnEvent publishes a column event if event client exists
func (s *service) publishColumnEvent(columnID, projectID int) {
	if s.eventClient == nil {
		return
	}

	// Publish database changed event
	_ = columnID  // Used for future enhancement
	_ = projectID // Used for future enhancement
}
