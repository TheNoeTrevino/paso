package standuplog

import (
	"context"
	"database/sql"
	"errors"
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
	ListByProjectInRange(ctx context.Context, projectID int, since, until time.Time) ([]models.StandupLog, error)
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

func (s *service) ListByProjectInRange(ctx context.Context, projectID int, since, until time.Time) ([]models.StandupLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}

	if !since.Before(until) {
		return nil, ErrInvalidDateRange
	}

	logs, err := s.queries.GetStandupLogsByProjectAndDateRange(ctx, types.GetStandupLogsByProjectAndDateRangeParams{
		ProjectID: int64(projectID),
		Since:     since,
		Until:     until,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list standup logs in range: %w", err)
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

	return database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := database.MustNewQuerier(tx, s.dbType)

		if _, err := qtx.GetStandupLog(ctx, int64(id)); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrLogNotFound
			}
			return fmt.Errorf("failed to get standup log: %w", err)
		}

		if err := qtx.DeleteStandupLog(ctx, int64(id)); err != nil {
			return fmt.Errorf("failed to delete standup log: %w", err)
		}

		return nil
	})
}
