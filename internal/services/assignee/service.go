package assignee

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/thenoetrevino/paso/internal/converters"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/models"
)

// Service defines all assignee-related business operations
type Service interface {
	// Read operations
	List(ctx context.Context) ([]*models.Assignee, error)
	GetByName(ctx context.Context, name string) (*models.Assignee, error)
	GetByID(ctx context.Context, id int) (*models.Assignee, error)

	// Write operations
	Create(ctx context.Context, name string) (*models.Assignee, error)
	GetOrCreate(ctx context.Context, name string) (*models.Assignee, error)
	Delete(ctx context.Context, id int) error
}

// service implements Service interface using database.Querier abstraction
type service struct {
	db      *sql.DB
	dbType  database.DatabaseType
	queries database.Querier
}

// NewService creates a new assignee service with database-agnostic queries.
func NewService(db *sql.DB, dbType database.DatabaseType) (Service, error) {
	queries, err := database.NewQuerier(db, dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create assignee service: %w", err)
	}
	return &service{
		db:      db,
		dbType:  dbType,
		queries: queries,
	}, nil
}

// List retrieves all assignees ordered by name
func (s *service) List(ctx context.Context) ([]*models.Assignee, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	assignees, err := s.queries.ListAssignees(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list assignees: %w", err)
	}
	return converters.AssigneesToModels(assignees), nil
}

// GetByName retrieves an assignee by name (case-insensitive)
func (s *service) GetByName(ctx context.Context, name string) (*models.Assignee, error) {
	name = strings.TrimSpace(name)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateName(name); err != nil {
		return nil, err
	}

	assignee, err := s.queries.GetAssigneeByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAssigneeNotFound
		}
		return nil, fmt.Errorf("failed to get assignee: %w", err)
	}

	return converters.AssigneeToModel(assignee), nil
}

// GetByID retrieves an assignee by ID
func (s *service) GetByID(ctx context.Context, id int) (*models.Assignee, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateAssigneeID(id); err != nil {
		return nil, err
	}

	assignee, err := s.queries.GetAssigneeByID(ctx, int64(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAssigneeNotFound
		}
		return nil, fmt.Errorf("failed to get assignee: %w", err)
	}

	return converters.AssigneeToModel(assignee), nil
}

// Create creates a new assignee with validation
func (s *service) Create(ctx context.Context, name string) (*models.Assignee, error) {
	name = strings.TrimSpace(name)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateName(name); err != nil {
		return nil, err
	}

	assignee, err := s.queries.CreateAssignee(ctx, name)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, fmt.Errorf("assignee with name '%s' already exists", name)
		}
		return nil, fmt.Errorf("failed to create assignee: %w", err)
	}

	return converters.AssigneeToModel(assignee), nil
}

// GetOrCreate retrieves an assignee by name or creates it if it doesn't exist
func (s *service) GetOrCreate(ctx context.Context, name string) (*models.Assignee, error) {
	name = strings.TrimSpace(name)

	assignee, err := s.GetByName(ctx, name)
	if err == nil {
		return assignee, nil
	}

	if err != ErrAssigneeNotFound {
		return nil, err
	}

	created, createErr := s.Create(ctx, name)
	if createErr == nil {
		return created, nil
	}

	if isUniqueConstraintError(createErr) || strings.Contains(createErr.Error(), "already exists") {
		return s.GetByName(ctx, name)
	}

	return nil, createErr
}

// Delete deletes an assignee after checking task count
// Tasks with this assignee will have assignee_id set to NULL via on delete set null
func (s *service) Delete(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateAssigneeID(id); err != nil {
		return err
	}

	rowsAffected, err := s.queries.DeleteAssignee(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("failed to delete assignee: %w", err)
	}

	if rowsAffected == 0 {
		return ErrAssigneeNotFound
	}

	return nil
}

// isUniqueConstraintError checks if an error is a unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "UNIQUE constraint failed") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "duplicate key")
}
