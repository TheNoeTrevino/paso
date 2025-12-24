package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// labelRepository handles pure data access for labels
// No business logic, no events, no validation - just database operations
type labelRepository struct {
	queries *generated.Queries
	db      *sql.DB
}

// newLabelRepository creates a new label repository
func newLabelRepository(queries *generated.Queries, db *sql.DB) *labelRepository {
	return &labelRepository{
		queries: queries,
		db:      db,
	}
}

// CreateLabel inserts a new label
func (r *labelRepository) CreateLabel(ctx context.Context, projectID int, name, color string) (*models.Label, error) {
	row, err := r.queries.CreateLabel(ctx, generated.CreateLabelParams{
		Name:      name,
		Color:     color,
		ProjectID: int64(projectID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}
	return toLabelModel(row), nil
}

// GetLabelsByProject retrieves all labels for a project
func (r *labelRepository) GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	rows, err := r.queries.GetLabelsByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get labels for project %d: %w", projectID, err)
	}

	labels := make([]*models.Label, len(rows))
	for i, row := range rows {
		labels[i] = toLabelModel(row)
	}
	return labels, nil
}

// GetLabelsForTask retrieves all labels for a task
func (r *labelRepository) GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	rows, err := r.queries.GetLabelsForTask(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get labels for task %d: %w", taskID, err)
	}

	labels := make([]*models.Label, len(rows))
	for i, row := range rows {
		labels[i] = toLabelModel(row)
	}
	return labels, nil
}

// UpdateLabel updates a label's name and color
func (r *labelRepository) UpdateLabel(ctx context.Context, labelID int, name, color string) error {
	err := r.queries.UpdateLabel(ctx, generated.UpdateLabelParams{
		Name:  name,
		Color: color,
		ID:    int64(labelID),
	})
	if err != nil {
		return fmt.Errorf("failed to update label %d: %w", labelID, err)
	}
	return nil
}

// DeleteLabel removes a label
func (r *labelRepository) DeleteLabel(ctx context.Context, labelID int) error {
	err := r.queries.DeleteLabel(ctx, int64(labelID))
	if err != nil {
		return fmt.Errorf("failed to delete label %d: %w", labelID, err)
	}
	return nil
}

// AddLabelToTask associates a label with a task
func (r *labelRepository) AddLabelToTask(ctx context.Context, taskID, labelID int) error {
	err := r.queries.AddLabelToTask(ctx, generated.AddLabelToTaskParams{
		TaskID:  int64(taskID),
		LabelID: int64(labelID),
	})
	if err != nil {
		return fmt.Errorf("failed to add label %d to task %d: %w", labelID, taskID, err)
	}
	return nil
}

// RemoveLabelFromTask removes the association between a label and a task
func (r *labelRepository) RemoveLabelFromTask(ctx context.Context, taskID, labelID int) error {
	err := r.queries.RemoveLabelFromTask(ctx, generated.RemoveLabelFromTaskParams{
		TaskID:  int64(taskID),
		LabelID: int64(labelID),
	})
	if err != nil {
		return fmt.Errorf("failed to remove label %d from task %d: %w", labelID, taskID, err)
	}
	return nil
}

// DeleteAllLabelsFromTask removes all labels from a task
func (r *labelRepository) DeleteAllLabelsFromTask(ctx context.Context, taskID int) error {
	err := r.queries.DeleteAllLabelsFromTask(ctx, int64(taskID))
	if err != nil {
		return fmt.Errorf("failed to delete all labels from task %d: %w", taskID, err)
	}
	return nil
}

// InsertTaskLabel inserts a task-label association
func (r *labelRepository) InsertTaskLabel(ctx context.Context, taskID, labelID int) error {
	err := r.queries.InsertTaskLabel(ctx, generated.InsertTaskLabelParams{
		TaskID:  int64(taskID),
		LabelID: int64(labelID),
	})
	if err != nil {
		return fmt.Errorf("failed to insert task-label association: %w", err)
	}
	return nil
}

// WithTx returns a new repository instance that uses the given transaction
func (r *labelRepository) WithTx(tx *sql.Tx) *labelRepository {
	return &labelRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

// BeginTx starts a new transaction
func (r *labelRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// ============================================================================
// MODEL CONVERSION HELPERS
// ============================================================================

func toLabelModel(row generated.Label) *models.Label {
	return &models.Label{
		ID:        int(row.ID),
		Name:      row.Name,
		Color:     row.Color,
		ProjectID: int(row.ProjectID),
	}
}
