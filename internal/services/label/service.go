package label

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/thenoetrevino/paso/internal/converters"
	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// Re-export error variables from models package for backward compatibility
var (
	ErrEmptyName        = models.ErrEmptyName
	ErrNameTooLong      = models.ErrNameTooLong
	ErrInvalidColor     = models.ErrInvalidColor
	ErrInvalidLabelID   = models.ErrInvalidLabelID
	ErrInvalidProjectID = models.ErrInvalidProjectID
	ErrInvalidTaskID    = models.ErrInvalidTaskID
	ErrLabelNotFound    = models.ErrLabelNotFound
)

// Hex color regex pattern
var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// Service defines all label-related business operations
type Service interface {
	// Read operations
	GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error)
	GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error)

	// Write operations
	CreateLabel(ctx context.Context, req CreateLabelRequest) (*models.Label, error)
	UpdateLabel(ctx context.Context, req UpdateLabelRequest) error
	DeleteLabel(ctx context.Context, id int) error
}

// CreateLabelRequest encapsulates data for creating a label
type CreateLabelRequest struct {
	ProjectID int
	Name      string
	Color     string // Hex color like #FF5733
}

// UpdateLabelRequest encapsulates data for updating a label
type UpdateLabelRequest struct {
	ID    int
	Name  *string
	Color *string
}

// service implements Service interface using SQLC directly
type service struct {
	db          *sql.DB
	queries     generated.Querier
	eventClient events.EventPublisher
}

// NewService creates a new label service
func NewService(db *sql.DB, eventClient events.EventPublisher) Service {
	return &service{
		db:          db,
		queries:     generated.New(db),
		eventClient: eventClient,
	}
}

// GetLabelsByProject retrieves all labels for a project
func (s *service) GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	labels, err := s.queries.GetLabelsByProject(ctx, int64(projectID))
	if err != nil {
		return nil, err
	}
	return converters.LabelsToModels(labels), nil
}

// GetLabelsForTask retrieves all labels for a task
func (s *service) GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	if taskID <= 0 {
		return nil, ErrInvalidTaskID
	}
	labels, err := s.queries.GetLabelsForTask(ctx, int64(taskID))
	if err != nil {
		return nil, err
	}
	return converters.LabelsToModels(labels), nil
}

// CreateLabel creates a new label with validation
func (s *service) CreateLabel(ctx context.Context, req CreateLabelRequest) (*models.Label, error) {
	// Validate request
	if err := s.validateCreateLabel(req); err != nil {
		return nil, err
	}

	// Create label
	label, err := s.queries.CreateLabel(ctx, generated.CreateLabelParams{
		Name:      req.Name,
		Color:     req.Color,
		ProjectID: int64(req.ProjectID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	// Publish event
	s.publishLabelEvent(ctx, int(label.ID), int(label.ProjectID))

	return converters.LabelToModel(label), nil
}

// UpdateLabel updates an existing label
func (s *service) UpdateLabel(ctx context.Context, req UpdateLabelRequest) error {
	// Validate label ID
	if req.ID <= 0 {
		return ErrInvalidLabelID
	}

	// Validate fields if provided
	if req.Name != nil && *req.Name == "" {
		return ErrEmptyName
	}
	if req.Name != nil && len(*req.Name) > 50 {
		return ErrNameTooLong
	}
	if req.Color != nil && !hexColorRegex.MatchString(*req.Color) {
		return ErrInvalidColor
	}

	// Get existing label to fill in missing fields
	existing, err := s.queries.GetLabelByID(ctx, int64(req.ID))
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrLabelNotFound
		}
		return fmt.Errorf("failed to get label: %w", err)
	}

	// Determine final values
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	color := existing.Color
	if req.Color != nil {
		color = *req.Color
	}

	// Update label
	if err := s.queries.UpdateLabel(ctx, generated.UpdateLabelParams{
		ID:    int64(req.ID),
		Name:  name,
		Color: color,
	}); err != nil {
		return fmt.Errorf("failed to update label: %w", err)
	}

	// Publish event
	s.publishLabelEvent(ctx, req.ID, int(existing.ProjectID))

	return nil
}

// DeleteLabel deletes a label
func (s *service) DeleteLabel(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidLabelID
	}

	// Get label to find project ID for event
	existing, err := s.queries.GetLabelByID(ctx, int64(id))
	if err != nil {
		if err == sql.ErrNoRows {
			// Label doesn't exist, but that's okay for deletion
			return nil
		}
		return fmt.Errorf("failed to get label: %w", err)
	}

	projectID := int(existing.ProjectID)

	// Delete label
	if err := s.queries.DeleteLabel(ctx, int64(id)); err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	// Publish event
	s.publishLabelEvent(ctx, id, projectID)

	return nil
}

// validateCreateLabel validates a CreateLabelRequest
func (s *service) validateCreateLabel(req CreateLabelRequest) error {
	if req.ProjectID <= 0 {
		return ErrInvalidProjectID
	}
	if req.Name == "" {
		return ErrEmptyName
	}
	if len(req.Name) > 50 {
		return ErrNameTooLong
	}
	if !hexColorRegex.MatchString(req.Color) {
		return ErrInvalidColor
	}
	return nil
}

// publishLabelEvent publishes a label event with retry logic
func (s *service) publishLabelEvent(ctx context.Context, labelID, projectID int) {
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
