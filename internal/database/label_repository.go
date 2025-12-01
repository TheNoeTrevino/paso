package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thenoetrevino/paso/internal/models"
)

// LabelRepo handles all label-related database operations.
type LabelRepo struct {
	db *sql.DB
}

// CreateLabel creates a new label in the database for a specific project
func (r *LabelRepo) CreateLabel(ctx context.Context, projectID int, name, color string) (*models.Label, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO labels (name, color, project_id) VALUES (?, ?, ?)`,
		name, color, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create label '%s' for project %d: %w", name, projectID, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get label ID after insert: %w", err)
	}

	return &models.Label{
		ID:        int(id),
		Name:      name,
		Color:     color,
		ProjectID: projectID,
	}, nil
}

// GetLabelsByProject retrieves all labels for a specific project
func (r *LabelRepo) GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, color, project_id FROM labels WHERE project_id = ? ORDER BY name`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels for project %d: %w", projectID, err)
	}
	defer rows.Close()

	labels := make([]*models.Label, 0, 20)
	for rows.Next() {
		label := &models.Label{}
		if err := rows.Scan(&label.ID, &label.Name, &label.Color, &label.ProjectID); err != nil {
			return nil, fmt.Errorf("failed to scan label row for project %d: %w", projectID, err)
		}
		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating label rows for project %d: %w", projectID, err)
	}
	return labels, nil
}

// GetLabelsForTask retrieves all labels associated with a task
func (r *LabelRepo) GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.name, l.color, l.project_id
		FROM labels l
		INNER JOIN task_labels tl ON l.id = tl.label_id
		WHERE tl.task_id = ?
		ORDER BY l.name
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels for task %d: %w", taskID, err)
	}
	defer rows.Close()

	labels := make([]*models.Label, 0, 10)
	for rows.Next() {
		label := &models.Label{}
		if err := rows.Scan(&label.ID, &label.Name, &label.Color, &label.ProjectID); err != nil {
			return nil, fmt.Errorf("failed to scan label row for task %d: %w", taskID, err)
		}
		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating label rows for task %d: %w", taskID, err)
	}
	return labels, nil
}

// UpdateLabel updates an existing label's name and color
func (r *LabelRepo) UpdateLabel(ctx context.Context, labelID int, name, color string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE labels SET name = ?, color = ? WHERE id = ?`,
		name, color, labelID,
	)
	if err != nil {
		return fmt.Errorf("failed to update label %d: %w", labelID, err)
	}
	return nil
}

// DeleteLabel removes a label from the database (cascade removes task associations)
func (r *LabelRepo) DeleteLabel(ctx context.Context, labelID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM labels WHERE id = ?", labelID)
	if err != nil {
		return fmt.Errorf("failed to delete label %d: %w", labelID, err)
	}
	return nil
}

// AddLabelToTask associates a label with a task
func (r *LabelRepo) AddLabelToTask(ctx context.Context, taskID, labelID int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO task_labels (task_id, label_id) VALUES (?, ?)`,
		taskID, labelID,
	)
	if err != nil {
		return fmt.Errorf("failed to add label %d to task %d: %w", labelID, taskID, err)
	}
	return nil
}

// RemoveLabelFromTask removes the association between a label and a task
func (r *LabelRepo) RemoveLabelFromTask(ctx context.Context, taskID, labelID int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM task_labels WHERE task_id = ? AND label_id = ?`,
		taskID, labelID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove label %d from task %d: %w", labelID, taskID, err)
	}
	return nil
}

// SetTaskLabels replaces all labels for a task with the given label IDs
func (r *LabelRepo) SetTaskLabels(ctx context.Context, taskID int, labelIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for setting labels on task %d: %w", taskID, err)
	}
	defer tx.Rollback()

	// Remove all existing labels
	_, err = tx.ExecContext(ctx, `DELETE FROM task_labels WHERE task_id = ?`, taskID)
	if err != nil {
		return fmt.Errorf("failed to clear existing labels for task %d: %w", taskID, err)
	}

	// Add new labels
	for _, labelID := range labelIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)`,
			taskID, labelID,
		)
		if err != nil {
			return fmt.Errorf("failed to add label %d to task %d: %w", labelID, taskID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit label changes for task %d: %w", taskID, err)
	}
	return nil
}
