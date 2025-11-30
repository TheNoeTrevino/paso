package database

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/thenoetrevino/paso/internal/models"
)

// ============================================================================
// Task Operations
// ============================================================================

// CreateTask creates a new task in the database
func CreateTask(ctx context.Context, db *sql.DB, title, description string, columnID, position int) (*models.Task, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO tasks (title, description, column_id, position)
		 VALUES (?, ?, ?, ?)`,
		title, description, columnID, position,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Retrieve the created task to get timestamps
	task := &models.Task{}
	err = db.QueryRowContext(ctx,
		`SELECT id, title, description, column_id, position, created_at, updated_at
		 FROM tasks WHERE id = ?`,
		id,
	).Scan(
		&task.ID, &task.Title, &task.Description,
		&task.ColumnID, &task.Position, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTasksByColumn retrieves all tasks for a specific column, ordered by position
func GetTasksByColumn(ctx context.Context, db *sql.DB, columnID int) ([]*models.Task, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, title, description, column_id, position, created_at, updated_at
		 FROM tasks
		 WHERE column_id = ?
		 ORDER BY position`,
		columnID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(
			&task.ID, &task.Title, &task.Description,
			&task.ColumnID, &task.Position, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// UpdateTaskColumn moves a task to a different column and/or position
func UpdateTaskColumn(ctx context.Context, db *sql.DB, taskID, newColumnID, newPosition int) error {
	_, err := db.ExecContext(ctx,
		`UPDATE tasks
		 SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		newColumnID, newPosition, taskID,
	)
	return err
}

// DeleteTask removes a task from the database
func DeleteTask(ctx context.Context, db *sql.DB, taskID int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", taskID)
	return err
}

// UpdateTaskTitle updates the title of an existing task
func UpdateTaskTitle(ctx context.Context, db *sql.DB, taskID int, title string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE tasks
		 SET title = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		title, taskID,
	)
	return err
}

// UpdateTask updates a task's title and description
func UpdateTask(ctx context.Context, db *sql.DB, taskID int, title, description string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE tasks
		 SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		title, description, taskID,
	)
	return err
}

// GetTaskCountByColumn returns the number of tasks in a specific column
func GetTaskCountByColumn(ctx context.Context, db *sql.DB, columnID int) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE column_id = ?", columnID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// MoveTask moves a task to a different column and position using a transaction
// This is a wrapper around UpdateTaskColumn that provides transactional guarantees
func MoveTask(ctx context.Context, db *sql.DB, taskID, newColumnID, newPosition int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update task's column and position
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks
		 SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		newColumnID, newPosition, taskID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// MoveTaskToNextColumn moves a task from its current column to the next column in the linked list
// Returns an error if the task doesn't exist or if there's no next column
func MoveTaskToNextColumn(ctx context.Context, db *sql.DB, taskID int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get the task's current column_id
	var currentColumnID int
	err = tx.QueryRowContext(ctx,
		`SELECT column_id FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&currentColumnID)
	if err != nil {
		return err
	}

	// 2. Get the next column's ID
	var nextColumnID sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT next_id FROM columns WHERE id = ?`,
		currentColumnID,
	).Scan(&nextColumnID)
	if err != nil {
		return err
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
		return err
	}

	// 5. Move the task to the next column (append to end)
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks
		 SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		nextColumnID.Int64, taskCount, taskID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// MoveTaskToPrevColumn moves a task from its current column to the previous column in the linked list
// Returns an error if the task doesn't exist or if there's no previous column
func MoveTaskToPrevColumn(ctx context.Context, db *sql.DB, taskID int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get the task's current column_id
	var currentColumnID int
	err = tx.QueryRowContext(ctx,
		`SELECT column_id FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&currentColumnID)
	if err != nil {
		return err
	}

	// 2. Get the previous column's ID
	var prevColumnID sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT prev_id FROM columns WHERE id = ?`,
		currentColumnID,
	).Scan(&prevColumnID)
	if err != nil {
		return err
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
		return err
	}

	// 5. Move the task to the previous column (append to end)
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks
		 SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		prevColumnID.Int64, taskCount, taskID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ============================================================================
// Task Summary/Detail Operations (DTOs with Labels)
// ============================================================================

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

// GetTaskSummariesByColumn retrieves task summaries for a column, including labels
// Uses a single query with LEFT JOIN to avoid N+1 query pattern
func GetTaskSummariesByColumn(ctx context.Context, db *sql.DB, columnID int) ([]*models.TaskSummary, error) {
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

	rows, err := db.QueryContext(ctx, query, columnID)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		// Parse concatenated label data
		summary.Labels = parseLabelStrings(labelIDsStr, labelNamesStr, labelColorsStr)
		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

// GetTaskDetail retrieves full task details including description, timestamps, and labels
func GetTaskDetail(ctx context.Context, db *sql.DB, taskID int) (*models.TaskDetail, error) {
	detail := &models.TaskDetail{}
	err := db.QueryRowContext(ctx,
		`SELECT id, title, description, column_id, position, created_at, updated_at
		 FROM tasks WHERE id = ?`,
		taskID,
	).Scan(
		&detail.ID, &detail.Title, &detail.Description,
		&detail.ColumnID, &detail.Position, &detail.CreatedAt, &detail.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Get labels
	labels, err := GetLabelsForTask(ctx, db, taskID)
	if err != nil {
		return nil, err
	}
	detail.Labels = labels

	return detail, nil
}
