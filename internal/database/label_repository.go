package database

import (
	"context"
	"database/sql"

	"github.com/thenoetrevino/paso/internal/models"
)

// ============================================================================
// Label Operations
// ============================================================================

// CreateLabel creates a new label in the database for a specific project
func CreateLabel(ctx context.Context, db *sql.DB, projectID int, name, color string) (*models.Label, error) {
	result, err := db.ExecContext(ctx, 
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

// GetAllLabels retrieves all labels from the database (for backward compatibility)
// Prefer GetLabelsByProject for project-specific label retrieval
func GetAllLabels(ctx context.Context, db *sql.DB) ([]*models.Label, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name, color, project_id FROM labels ORDER BY name`)
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

// GetLabelsByProject retrieves all labels for a specific project
func GetLabelsByProject(ctx context.Context, db *sql.DB, projectID int) ([]*models.Label, error) {
	rows, err := db.QueryContext(ctx, 
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

// UpdateLabel updates an existing label's name and color
func UpdateLabel(ctx context.Context, db *sql.DB, labelID int, name, color string) error {
	_, err := db.ExecContext(ctx, 
		`UPDATE labels SET name = ?, color = ? WHERE id = ?`,
		name, color, labelID,
	)
	return err
}

// DeleteLabel removes a label from the database (cascade removes task associations)
func DeleteLabel(ctx context.Context, db *sql.DB, labelID int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM labels WHERE id = ?", labelID)
	return err
}

// AddLabelToTask associates a label with a task
func AddLabelToTask(ctx context.Context, db *sql.DB, taskID, labelID int) error {
	_, err := db.ExecContext(ctx, 
		`INSERT OR IGNORE INTO task_labels (task_id, label_id) VALUES (?, ?)`,
		taskID, labelID,
	)
	return err
}

// RemoveLabelFromTask removes the association between a label and a task
func RemoveLabelFromTask(ctx context.Context, db *sql.DB, taskID, labelID int) error {
	_, err := db.ExecContext(ctx, 
		`DELETE FROM task_labels WHERE task_id = ? AND label_id = ?`,
		taskID, labelID,
	)
	return err
}

// GetLabelsForTask retrieves all labels associated with a task
func GetLabelsForTask(ctx context.Context, db *sql.DB, taskID int) ([]*models.Label, error) {
	rows, err := db.QueryContext(ctx, `
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

// SetTaskLabels replaces all labels for a task with the given label IDs
func SetTaskLabels(ctx context.Context, db *sql.DB, taskID int, labelIDs []int) error {
	tx, err := db.BeginTx(ctx, nil)
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
