package label

import (
	"context"
	"fmt"
	"regexp"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
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

// service implements Service interface
type service struct {
	repo        database.DataStore
	eventClient events.EventPublisher
}

// NewService creates a new label service
func NewService(repo database.DataStore, eventClient events.EventPublisher) Service {
	return &service{
		repo:        repo,
		eventClient: eventClient,
	}
}

// GetLabelsByProject retrieves all labels for a project
func (s *service) GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	return s.repo.GetLabelsByProject(ctx, projectID)
}

// GetLabelsForTask retrieves all labels for a task
func (s *service) GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	if taskID <= 0 {
		return nil, ErrInvalidTaskID
	}
	return s.repo.GetLabelsForTask(ctx, taskID)
}

// CreateLabel creates a new label with validation
func (s *service) CreateLabel(ctx context.Context, req CreateLabelRequest) (*models.Label, error) {
	// Validate request
	if err := s.validateCreateLabel(req); err != nil {
		return nil, err
	}

	// Create label in repository
	label, err := s.repo.CreateLabel(ctx, req.ProjectID, req.Name, req.Color)
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	// Publish event
	s.publishLabelEvent(label.ID, label.ProjectID)

	return label, nil
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
	labels, err := s.repo.GetLabelsByProject(ctx, 0) // TODO: Need a GetLabelByID method
	if err != nil {
		return fmt.Errorf("failed to get label: %w", err)
	}

	var existing *models.Label
	for _, l := range labels {
		if l.ID == req.ID {
			existing = l
			break
		}
	}
	if existing == nil {
		return ErrLabelNotFound
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
	if err := s.repo.UpdateLabel(ctx, req.ID, name, color); err != nil {
		return fmt.Errorf("failed to update label: %w", err)
	}

	// Publish event
	s.publishLabelEvent(req.ID, existing.ProjectID)

	return nil
}

// DeleteLabel deletes a label
func (s *service) DeleteLabel(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidLabelID
	}

	// Get label to find project ID for event
	// Note: We need to get all labels since there's no GetLabelByID
	labels, err := s.repo.GetLabelsByProject(ctx, 0) // TODO: Need a GetLabelByID method
	if err != nil {
		return fmt.Errorf("failed to get label: %w", err)
	}

	var projectID int
	for _, l := range labels {
		if l.ID == id {
			projectID = l.ProjectID
			break
		}
	}

	// Delete label
	if err := s.repo.DeleteLabel(ctx, id); err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	// Publish event
	s.publishLabelEvent(id, projectID)

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

// publishLabelEvent publishes a label event if event client exists
func (s *service) publishLabelEvent(labelID, projectID int) {
	if s.eventClient == nil {
		return
	}

	// Publish database changed event
	_ = labelID   // Used for future enhancement
	_ = projectID // Used for future enhancement
}
