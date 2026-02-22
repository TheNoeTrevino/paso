package standuplog

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

type Service interface {
	Create(ctx context.Context, projectID int, content string) (*models.StandupLog, error)
	ListByProject(ctx context.Context, projectID int) ([]models.StandupLog, error)
	GetByID(ctx context.Context, id int) (*models.StandupLog, error)
	Delete(ctx context.Context, id int) error
}

type service struct {
	db      *sql.DB
	dbType  database.DatabaseType
	queries types.Querier
}

func NewService(db *sql.DB, dbType database.DatabaseType) (Service, error) {
	queries, err := database.NewQuerier(db, dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create standuplog service: %w", err)
	}
	return &service{
		db:      db,
		dbType:  dbType,
		queries: queries,
	}, nil
}

func (s *service) Create(ctx context.Context, projectID int, content string) (*models.StandupLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateCreateParams(projectID, content); err != nil {
		return nil, err
	}

	log, err := s.queries.CreateStandupLog(ctx, types.CreateStandupLogParams{
		ProjectID: int64(projectID),
		Content:   content,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create standup log: %w", err)
	}

	return converters.StandupLogToModel(log), nil
}

func (s *service) ListByProject(ctx context.Context, projectID int) ([]models.StandupLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}

	logs, err := s.queries.GetStandupLogsByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to list standup logs: %w", err)
	}

	return converters.StandupLogsToModels(logs), nil
}

func (s *service) GetByID(ctx context.Context, id int) (*models.StandupLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if id <= 0 {
		return nil, ErrInvalidLogID
	}

	log, err := s.queries.GetStandupLog(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get standup log: %w", err)
	}

	return converters.StandupLogToModel(log), nil
}

func (s *service) Delete(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if id <= 0 {
		return ErrInvalidLogID
	}

	if err := s.queries.DeleteStandupLog(ctx, int64(id)); err != nil {
		return fmt.Errorf("failed to delete standup log: %w", err)
	}

	return nil
}
