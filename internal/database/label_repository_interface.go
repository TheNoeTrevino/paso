package database

import (
	"context"

	"github.com/thenoetrevino/paso/internal/models"
)

// LabelReader defines read operations for labels.
type LabelReader interface {
	GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error)
	GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error)
}

// LabelWriter defines write operations for labels.
type LabelWriter interface {
	CreateLabel(ctx context.Context, projectID int, name, color string) (*models.Label, error)
	UpdateLabel(ctx context.Context, id int, name, color string) error
	DeleteLabel(ctx context.Context, id int) error
}

// TaskLabelManager defines operations for managing task-label associations.
type TaskLabelManager interface {
	AddLabelToTask(ctx context.Context, taskID, labelID int) error
	RemoveLabelFromTask(ctx context.Context, taskID, labelID int) error
	SetTaskLabels(ctx context.Context, taskID int, labelIDs []int) error
}

// LabelRepository combines all label-related operations.
type LabelRepository interface {
	LabelReader
	LabelWriter
	TaskLabelManager
}
