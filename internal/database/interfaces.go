// Package database defines repository interfaces for data access
package database

import (
	"context"

	"github.com/thenoetrevino/paso/internal/models"
)

// ProjectRepository defines operations for project management
type ProjectRepository interface {
	Create(ctx context.Context, name, description string) (*models.Project, error)
	GetAll(ctx context.Context) ([]*models.Project, error)
	GetByID(ctx context.Context, id int) (*models.Project, error)
	Update(ctx context.Context, id int, name, description string) error
	Delete(ctx context.Context, id int) error
}

// ColumnRepository defines operations for column management
type ColumnRepository interface {
	Create(ctx context.Context, name string, projectID int, afterID *int) (*models.Column, error)
	GetByProject(ctx context.Context, projectID int) ([]*models.Column, error)
	GetByID(ctx context.Context, id int) (*models.Column, error)
	UpdateName(ctx context.Context, id int, name string) error
	UpdatePositions(ctx context.Context, prevID, currentID, nextID *int, newPrevID, newNextID *int) error
	Delete(ctx context.Context, id int) error
}

// TaskRepository defines operations for task management
type TaskRepository interface {
	Create(ctx context.Context, title, description string, columnID, position int) (*models.Task, error)
	GetByID(ctx context.Context, id int) (*models.Task, error)
	GetSummariesByColumn(ctx context.Context, columnID int) ([]*models.TaskSummary, error)
	GetDetail(ctx context.Context, id int) (*models.TaskDetail, error)
	GetCountByColumn(ctx context.Context, columnID int) (int, error)
	Update(ctx context.Context, id int, title, description string) error
	UpdatePosition(ctx context.Context, id, position int) error
	Move(ctx context.Context, id, newColumnID, newPosition int) error
	Delete(ctx context.Context, id int) error
}

// LabelRepository defines operations for label management
type LabelRepository interface {
	Create(ctx context.Context, projectID int, name, color string) (*models.Label, error)
	GetByProject(ctx context.Context, projectID int) ([]*models.Label, error)
	GetByID(ctx context.Context, id int) (*models.Label, error)
	GetForTask(ctx context.Context, taskID int) ([]*models.Label, error)
	Update(ctx context.Context, id int, name, color string) error
	Delete(ctx context.Context, id int) error
	AddToTask(ctx context.Context, taskID, labelID int) error
	RemoveFromTask(ctx context.Context, taskID, labelID int) error
	SetForTask(ctx context.Context, taskID int, labelIDs []int) error
}
