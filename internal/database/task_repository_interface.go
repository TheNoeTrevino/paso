package database

import (
	"context"

	"github.com/thenoetrevino/paso/internal/models"
)

// TaskReader defines read operations for tasks.
type TaskReader interface {
	GetTaskSummariesByColumn(ctx context.Context, columnID int) ([]*models.TaskSummary, error)
	GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error)
	GetTasksByColumn(ctx context.Context, columnID int) ([]*models.Task, error)
	GetTaskDetail(ctx context.Context, id int) (*models.TaskDetail, error)
	GetTaskCountByColumn(ctx context.Context, columnID int) (int, error)
}

// TaskWriter defines write operations for tasks.
type TaskWriter interface {
	CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error)
	UpdateTask(ctx context.Context, id int, title, description string) error
	DeleteTask(ctx context.Context, id int) error
}

// TaskMover defines operations for moving tasks between columns and within columns.
type TaskMover interface {
	MoveTaskToNextColumn(ctx context.Context, taskID int) error
	MoveTaskToPrevColumn(ctx context.Context, taskID int) error
	SwapTaskUp(ctx context.Context, taskID int) error
	SwapTaskDown(ctx context.Context, taskID int) error
}

// TaskRepository combines all task-related operations.
type TaskRepository interface {
	TaskReader
	TaskWriter
	TaskMover
}
