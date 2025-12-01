package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/thenoetrevino/paso/internal/models"
)

// TaskRepo handles all task-related database operations.
type TaskRepo struct {
	db *sql.DB
}

// CreateTask creates a new task in the database
func (r *TaskRepo) CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (title, description, column_id, position)
		 VALUES (?, ?, ?, ?)`,
		title, description, columnID, position,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID for task: %w", err)
	}

	// Retrieve the created task to get timestamps
	task := &models.Task{}
	err = r.db.QueryRowContext(ctx,
		`SELECT id, title, description, column_id, position, created_at, updated_at
		 FROM tasks WHERE id = ?`,
		id,
	).Scan(
		&task.ID, &task.Title, &task.Description,
		&task.ColumnID, &task.Position, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created task %d: %w", id, err)
	}

	return task, nil
}

// GetTasksByColumn retrieves all tasks for a column with full details
// This is primarily used for testing and admin purposes
func (r *TaskRepo) GetTasksByColumn(ctx context.Context, columnID int) ([]*models.Task, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, description, column_id, position, created_at, updated_at
		 FROM tasks WHERE column_id = ? ORDER BY position`,
		columnID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks for column %d: %w", columnID, err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(
			&task.ID, &task.Title, &task.Description,
			&task.ColumnID, &task.Position, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan task row for column %d: %w", columnID, err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks for column %d: %w", columnID, err)
	}
	return tasks, nil
}

// GetTaskSummariesByColumn retrieves task summaries for a column, including labels
// Uses a single query with LEFT JOIN to avoid N+1 query pattern
func (r *TaskRepo) GetTaskSummariesByColumn(ctx context.Context, columnID int) ([]*models.TaskSummary, error) {
	// Single query with LEFT JOIN to fetch tasks and their labels
	// GROUP_CONCAT aggregates labels for each task
	query := `
		SELECT
			t.id,
			t.title,
			t.column_id,
			t.position,
			GROUP_CONCAT(l.id, '|') as label_ids,
			GROUP_CONCAT(l.name, '|') as label_names,
			GROUP_CONCAT(l.color, '|') as label_colors
		FROM tasks t
		LEFT JOIN task_labels tl ON t.id = tl.task_id
		LEFT JOIN labels l ON tl.label_id = l.id
		WHERE t.column_id = ?
		GROUP BY t.id, t.title, t.column_id, t.position
		ORDER BY t.position`

	rows, err := r.db.QueryContext(ctx, query, columnID)
	if err != nil {
		return nil, fmt.Errorf("failed to query task summaries for column %d: %w", columnID, err)
	}
	defer rows.Close()

	var summaries []*models.TaskSummary
	for rows.Next() {
		var labelIDsStr, labelNamesStr, labelColorsStr sql.NullString
		summary := &models.TaskSummary{}

		if err := rows.Scan(
			&summary.ID,
			&summary.Title,
			&summary.ColumnID,
			&summary.Position,
			&labelIDsStr,
			&labelNamesStr,
			&labelColorsStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan task summary for column %d: %w", columnID, err)
		}

		// Parse concatenated label data
		summary.Labels = parseLabelStrings(labelIDsStr, labelNamesStr, labelColorsStr)
		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task summaries for column %d: %w", columnID, err)
	}
	return summaries, nil
}

// GetTaskDetail retrieves full task details including description, timestamps, and labels
func (r *TaskRepo) GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error) {
	detail := &models.TaskDetail{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, column_id, position, created_at, updated_at
		 FROM tasks WHERE id = ?`,
		taskID,
	).Scan(
		&detail.ID, &detail.Title, &detail.Description,
		&detail.ColumnID, &detail.Position, &detail.CreatedAt, &detail.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task detail for task %d: %w", taskID, err)
	}

	// Get labels for this task
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

	var labels []*models.Label
	for rows.Next() {
		label := &models.Label{}
		if err := rows.Scan(&label.ID, &label.Name, &label.Color, &label.ProjectID); err != nil {
			return nil, fmt.Errorf("failed to scan label for task %d: %w", taskID, err)
		}
		labels = append(labels, label)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating labels for task %d: %w", taskID, err)
	}

	detail.Labels = labels
	return detail, nil
}

// GetTaskCountByColumn returns the number of tasks in a specific column
func (r *TaskRepo) GetTaskCountByColumn(ctx context.Context, columnID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE column_id = ?", columnID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get task count for column %d: %w", columnID, err)
	}
	return count, nil
}

// UpdateTask updates a task's title and description
func (r *TaskRepo) UpdateTask(ctx context.Context, taskID int, title, description string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks
		 SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		title, description, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task %d: %w", taskID, err)
	}
	return nil
}

// MoveTaskToNextColumn moves a task from its current column to the next column in the linked list
// Returns an error if the task doesn't exist or if there's no next column
func (r *TaskRepo) MoveTaskToNextColumn(ctx context.Context, taskID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for moving task %d to next column: %w", taskID, err)
	}
	defer tx.Rollback()

	// 1. Get the task's current column_id
	var currentColumnID int
	err = tx.QueryRowContext(ctx,
		`SELECT column_id FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&currentColumnID)
	if err != nil {
		return fmt.Errorf("failed to get current column for task %d: %w", taskID, err)
	}

	// 2. Get the next column's ID
	var nextColumnID sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT next_id FROM columns WHERE id = ?`,
		currentColumnID,
	).Scan(&nextColumnID)
	if err != nil {
		return fmt.Errorf("failed to get next column for column %d: %w", currentColumnID, err)
	}

	// 3. Check if there's a next column
	if !nextColumnID.Valid {
		return sql.ErrNoRows // Already at last column
	}

	// 4. Get the number of tasks in the next column to determine position
	var taskCount int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tasks WHERE column_id = ?`,
		nextColumnID.Int64,
	).Scan(&taskCount)
	if err != nil {
		return fmt.Errorf("failed to count tasks in target column %d: %w", nextColumnID.Int64, err)
	}

	// 5. Move the task to the next column (append to end)
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks
		 SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		nextColumnID.Int64, taskCount, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to move task %d to next column: %w", taskID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for moving task %d: %w", taskID, err)
	}
	return nil
}

// MoveTaskToPrevColumn moves a task from its current column to the previous column in the linked list
// Returns an error if the task doesn't exist or if there's no previous column
func (r *TaskRepo) MoveTaskToPrevColumn(ctx context.Context, taskID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for moving task %d to prev column: %w", taskID, err)
	}
	defer tx.Rollback()

	// 1. Get the task's current column_id
	var currentColumnID int
	err = tx.QueryRowContext(ctx,
		`SELECT column_id FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&currentColumnID)
	if err != nil {
		return fmt.Errorf("failed to get current column for task %d: %w", taskID, err)
	}

	// 2. Get the previous column's ID
	var prevColumnID sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT prev_id FROM columns WHERE id = ?`,
		currentColumnID,
	).Scan(&prevColumnID)
	if err != nil {
		return fmt.Errorf("failed to get prev column for column %d: %w", currentColumnID, err)
	}

	// 3. Check if there's a previous column
	if !prevColumnID.Valid {
		return sql.ErrNoRows // Already at first column
	}

	// 4. Get the number of tasks in the previous column to determine position
	var taskCount int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tasks WHERE column_id = ?`,
		prevColumnID.Int64,
	).Scan(&taskCount)
	if err != nil {
		return fmt.Errorf("failed to count tasks in target column %d: %w", prevColumnID.Int64, err)
	}

	// 5. Move the task to the previous column (append to end)
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks
		 SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		prevColumnID.Int64, taskCount, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to move task %d to prev column: %w", taskID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for moving task %d: %w", taskID, err)
	}
	return nil
}

// DeleteTask removes a task from the database
func (r *TaskRepo) DeleteTask(ctx context.Context, taskID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task %d: %w", taskID, err)
	}
	return nil
}

// parseLabelStrings parses pipe-delimited label data from GROUP_CONCAT
func parseLabelStrings(labelIDsStr, labelNamesStr, labelColorsStr sql.NullString) []*models.Label {
	// If no labels exist for this task, return empty slice
	if !labelIDsStr.Valid || !labelNamesStr.Valid || !labelColorsStr.Valid {
		return []*models.Label{}
	}

	// Split the concatenated strings
	ids := strings.Split(labelIDsStr.String, "|")
	names := strings.Split(labelNamesStr.String, "|")
	colors := strings.Split(labelColorsStr.String, "|")

	// Ensure all arrays have the same length
	if len(ids) != len(names) || len(ids) != len(colors) {
		return []*models.Label{}
	}

	// Build label objects
	labels := make([]*models.Label, 0, len(ids))
	for i := range ids {
		id, err := strconv.Atoi(ids[i])
		if err != nil {
			continue // Skip malformed data
		}
		labels = append(labels, &models.Label{
			ID:    id,
			Name:  names[i],
			Color: colors[i],
		})
	}

	return labels
}
