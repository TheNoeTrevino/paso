package label

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/thenoetrevino/paso/internal/converters"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/types"
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
	CountTasksByLabel(ctx context.Context, labelID int) (int, error)

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

// service implements Service interface using database.Querier abstraction
type service struct {
	db          *sql.DB
	dbType      database.DatabaseType
	queries     database.Querier
	eventClient events.EventPublisher
}

// NewService creates a new label service with database-agnostic queries.
func NewService(db *sql.DB, dbType database.DatabaseType, eventClient events.EventPublisher) (Service, error) {
	queries, err := database.NewQuerier(db, dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create label service: %w", err)
	}
	return &service{
		db:          db,
		dbType:      dbType,
		queries:     queries,
		eventClient: eventClient,
	}, nil
}

// GetLabelsByProject retrieves all labels for a project
func (s *service) GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateProjectID(projectID); err != nil {
		return nil, err
	}
	labels, err := s.queries.GetLabelsByProject(ctx, int64(projectID))
	if err != nil {
		return nil, err
	}
	return converters.LabelsToModels(labels), nil
}

// GetLabelsForTask retrieves all labels for a task
func (s *service) GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	if err := validateTaskID(taskID); err != nil {
		return nil, err
	}
	labels, err := s.queries.GetLabelsForTask(ctx, int64(taskID))
	if err != nil {
		return nil, err
	}
	return converters.LabelsToModels(labels), nil
}

// CountTasksByLabel returns the count of tasks that have this label attached
func (s *service) CountTasksByLabel(ctx context.Context, labelID int) (int, error) {
	if err := validateLabelID(labelID); err != nil {
		return 0, err
	}
	count, err := s.queries.CountTasksByLabel(ctx, int64(labelID))
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks by label: %w", err)
	}
	return int(count), nil
}

// CreateLabel creates a new label with validation
func (s *service) CreateLabel(ctx context.Context, req CreateLabelRequest) (*models.Label, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateCreateLabelRequest(req); err != nil {
		return nil, err
	}

	// Create label
	label, err := s.queries.CreateLabel(ctx, types.CreateLabelParams{
		Name:      req.Name,
		Color:     req.Color,
		ProjectID: int64(req.ProjectID),
	})
	if err != nil {
		// Check for unique constraint violation
		if isUniqueConstraintError(err) {
			return nil, fmt.Errorf("failed to create label: label with name '%s' already exists in this project", req.Name)
		}
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	// Publish event
	s.publishLabelEvent(ctx, int(label.ID), int(label.ProjectID))

	return converters.LabelToModel(label), nil
}

// UpdateLabel updates an existing label
func (s *service) UpdateLabel(ctx context.Context, req UpdateLabelRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateUpdateLabelRequest(req); err != nil {
		return err
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
	if err := s.queries.UpdateLabel(ctx, types.UpdateLabelParams{
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
	if err := validateLabelID(id); err != nil {
		return err
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

// isUniqueConstraintError checks if an error is a SQLite unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite returns "UNIQUE constraint failed" in the error message
	errStr := err.Error()
	return strings.Contains(errStr, "UNIQUE constraint failed") || strings.Contains(errStr, "constraint failed")
}
