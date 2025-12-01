package database

import (
	"context"
	"database/sql"

	"github.com/thenoetrevino/paso/internal/models"
)

// LabelRepo handles all label-related database operations.
type LabelRepo struct {
	db *sql.DB
}

// Create creates a new label in the database for a specific project
func (r *LabelRepo) Create(ctx context.Context, projectID int, name, color string) (*models.Label, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO labels (name, color, project_id) VALUES (?, ?, ?)`,
		name, color, projectID,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Label{
		ID:        int(id),
		Name:      name,
		Color:     color,
		ProjectID: projectID,
	}, nil
}

// GetByProject retrieves all labels for a specific project
func (r *LabelRepo) GetByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, color, project_id FROM labels WHERE project_id = ? ORDER BY name`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*models.Label
	for rows.Next() {
		label := &models.Label{}
		if err := rows.Scan(&label.ID, &label.Name, &label.Color, &label.ProjectID); err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	return labels, rows.Err()
}

// GetForTask retrieves all labels associated with a task
func (r *LabelRepo) GetForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.name, l.color, l.project_id
		FROM labels l
		INNER JOIN task_labels tl ON l.id = tl.label_id
		WHERE tl.task_id = ?
		ORDER BY l.name
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []*models.Label
	for rows.Next() {
		label := &models.Label{}
		if err := rows.Scan(&label.ID, &label.Name, &label.Color, &label.ProjectID); err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	return labels, rows.Err()
}

// Update updates an existing label's name and color
func (r *LabelRepo) Update(ctx context.Context, labelID int, name, color string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE labels SET name = ?, color = ? WHERE id = ?`,
		name, color, labelID,
	)
	return err
}

// Delete removes a label from the database (cascade removes task associations)
func (r *LabelRepo) Delete(ctx context.Context, labelID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM labels WHERE id = ?", labelID)
	return err
}

// AddToTask associates a label with a task
func (r *LabelRepo) AddToTask(ctx context.Context, taskID, labelID int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO task_labels (task_id, label_id) VALUES (?, ?)`,
		taskID, labelID,
	)
	return err
}

// RemoveFromTask removes the association between a label and a task
func (r *LabelRepo) RemoveFromTask(ctx context.Context, taskID, labelID int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM task_labels WHERE task_id = ? AND label_id = ?`,
		taskID, labelID,
	)
	return err
}

// SetForTask replaces all labels for a task with the given label IDs
func (r *LabelRepo) SetForTask(ctx context.Context, taskID int, labelIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove all existing labels
	_, err = tx.ExecContext(ctx, `DELETE FROM task_labels WHERE task_id = ?`, taskID)
	if err != nil {
		return err
	}

	// Add new labels
	for _, labelID := range labelIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)`,
			taskID, labelID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
