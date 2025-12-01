// Package database defines repository interfaces for data access
package database

import (
	"context"

	"github.com/thenoetrevino/paso/internal/models"
)

// DataStore defines the unified interface for all data operations needed by the TUI.
// This interface enables mocking with testify for unit testing.
type DataStore interface {
	// Projects
	CreateProject(ctx context.Context, name, description string) (*models.Project, error)
	GetAllProjects(ctx context.Context) ([]*models.Project, error)
	GetProjectByID(ctx context.Context, id int) (*models.Project, error)
	UpdateProject(ctx context.Context, id int, name, description string) error
	DeleteProject(ctx context.Context, id int) error

	// Columns
	CreateColumn(ctx context.Context, name string, projectID int, afterID *int) (*models.Column, error)
	GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error)
	GetColumnByID(ctx context.Context, id int) (*models.Column, error)
	UpdateColumnName(ctx context.Context, id int, name string) error
	DeleteColumn(ctx context.Context, id int) error

	// Tasks
	CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error)
	GetTaskSummariesByColumn(ctx context.Context, columnID int) ([]*models.TaskSummary, error)
	GetTasksByColumn(ctx context.Context, columnID int) ([]*models.Task, error)
	GetTaskDetail(ctx context.Context, id int) (*models.TaskDetail, error)
	GetTaskCountByColumn(ctx context.Context, columnID int) (int, error)
	UpdateTask(ctx context.Context, id int, title, description string) error
	MoveTaskToNextColumn(ctx context.Context, taskID int) error
	MoveTaskToPrevColumn(ctx context.Context, taskID int) error
	DeleteTask(ctx context.Context, id int) error

	// Labels
	CreateLabel(ctx context.Context, projectID int, name, color string) (*models.Label, error)
	GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error)
	GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error)
	UpdateLabel(ctx context.Context, id int, name, color string) error
	DeleteLabel(ctx context.Context, id int) error
	AddLabelToTask(ctx context.Context, taskID, labelID int) error
	RemoveLabelFromTask(ctx context.Context, taskID, labelID int) error
	SetTaskLabels(ctx context.Context, taskID int, labelIDs []int) error
}
