package taskevent

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thenoetrevino/paso/internal/converters"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/models"
)

// Service defines task event operations
type Service interface {
	// Transaction-aware methods (accept a Querier to work within existing transactions)
	CreateTaskCreatedEvent(ctx context.Context, qtx types.Querier, taskID int, title, author string) error
	CreateTaskMovedEvent(ctx context.Context, qtx types.Querier, taskID int, fromColumn, toColumn, author string) error
	CreateTaskAssociatedEvent(ctx context.Context, qtx types.Querier, taskID, relatedTaskID int, relatedTitle, relationLabel, author string) error
	CreateTaskDisassociatedEvent(ctx context.Context, qtx types.Querier, taskID, relatedTaskID int, author string) error
	CreateLabelAddedEvent(ctx context.Context, qtx types.Querier, taskID int, labelName, author string) error
	CreateLabelRemovedEvent(ctx context.Context, qtx types.Querier, taskID int, labelName, author string) error
	CreatePriorityChangedEvent(ctx context.Context, qtx types.Querier, taskID int, oldPriority, newPriority, author string) error
	CreateTypeChangedEvent(ctx context.Context, qtx types.Querier, taskID int, oldType, newType, author string) error

	// Read operations
	GetEventsByTask(ctx context.Context, taskID int) ([]models.TaskEvent, error)
}

type service struct {
	db      *sql.DB
	dbType  database.DatabaseType
	queries types.Querier
}

// NewService creates a new task event service
func NewService(db *sql.DB, dbType database.DatabaseType) (Service, error) {
	queries, err := database.NewQuerier(db, dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create taskevent service: %w", err)
	}
	return &service{
		db:      db,
		dbType:  dbType,
		queries: queries,
	}, nil
}

// createEvent is a helper that creates an event with the given content
func (s *service) createEvent(ctx context.Context, qtx types.Querier, taskID int, content, author string) error {
	if err := validateEventParams(taskID, content); err != nil {
		return err
	}

	_, err := qtx.CreateTaskEvent(ctx, types.CreateTaskEventParams{
		TaskID:  int64(taskID),
		Content: content,
		Author:  author,
	})
	if err != nil {
		return fmt.Errorf("failed to create task event: %w", err)
	}
	return nil
}

func (s *service) CreateTaskCreatedEvent(ctx context.Context, qtx types.Querier, taskID int, title, author string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	content := fmt.Sprintf("Task created: %s", title)
	return s.createEvent(ctx, qtx, taskID, content, author)
}

func (s *service) CreateTaskMovedEvent(ctx context.Context, qtx types.Querier, taskID int, fromColumn, toColumn, author string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	content := fmt.Sprintf("Moved from '%s' to '%s'", fromColumn, toColumn)
	return s.createEvent(ctx, qtx, taskID, content, author)
}

func (s *service) CreateTaskAssociatedEvent(ctx context.Context, qtx types.Querier, taskID, relatedTaskID int, relatedTitle, relationLabel, author string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	content := fmt.Sprintf("Associated to task #%d (%s) as: %s", relatedTaskID, relatedTitle, relationLabel)
	return s.createEvent(ctx, qtx, taskID, content, author)
}

func (s *service) CreateTaskDisassociatedEvent(ctx context.Context, qtx types.Querier, taskID, relatedTaskID int, author string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	content := fmt.Sprintf("Removed association to task #%d", relatedTaskID)
	return s.createEvent(ctx, qtx, taskID, content, author)
}

func (s *service) CreateLabelAddedEvent(ctx context.Context, qtx types.Querier, taskID int, labelName, author string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	content := fmt.Sprintf("Added label: %s", labelName)
	return s.createEvent(ctx, qtx, taskID, content, author)
}

func (s *service) CreateLabelRemovedEvent(ctx context.Context, qtx types.Querier, taskID int, labelName, author string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	content := fmt.Sprintf("Removed label: %s", labelName)
	return s.createEvent(ctx, qtx, taskID, content, author)
}

func (s *service) CreatePriorityChangedEvent(ctx context.Context, qtx types.Querier, taskID int, oldPriority, newPriority, author string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	content := fmt.Sprintf("Changed priority from '%s' to '%s'", oldPriority, newPriority)
	return s.createEvent(ctx, qtx, taskID, content, author)
}

func (s *service) CreateTypeChangedEvent(ctx context.Context, qtx types.Querier, taskID int, oldType, newType, author string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	content := fmt.Sprintf("Changed type from '%s' to '%s'", oldType, newType)
	return s.createEvent(ctx, qtx, taskID, content, author)
}

func (s *service) GetEventsByTask(ctx context.Context, taskID int) ([]models.TaskEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if taskID <= 0 {
		return nil, ErrInvalidTaskID
	}

	events, err := s.queries.GetEventsByTask(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	return converters.TaskEventsToModels(events), nil
}
