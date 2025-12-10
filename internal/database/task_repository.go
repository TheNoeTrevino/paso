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

// CreateTask creates a new task in the database with an auto-assigned ticket number
func (r *TaskRepo) CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
	// Get project ID from column
	var projectID int
	err := r.db.QueryRowContext(ctx,
		`SELECT project_id FROM columns WHERE id = ?`,
		columnID,
	).Scan(&projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project for column %d: %w", columnID, err)
	}

	// Start transaction for ticket number allocation and task creation
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get next ticket number for this project
	var ticketNumber int
	err = tx.QueryRowContext(ctx,
		`SELECT next_ticket_number FROM project_counters WHERE project_id = ?`,
		projectID,
	).Scan(&ticketNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket number for project %d: %w", projectID, err)
	}

	// Increment ticket counter
	_, err = tx.ExecContext(ctx,
		`UPDATE project_counters SET next_ticket_number = next_ticket_number + 1 WHERE project_id = ?`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to increment ticket counter for project %d: %w", projectID, err)
	}

	// Insert task with ticket number
	result, err := tx.ExecContext(ctx,
		`INSERT INTO tasks (title, description, column_id, position, ticket_number)
		 VALUES (?, ?, ?, ?, ?)`,
		title, description, columnID, position, ticketNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID for task: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit task creation transaction: %w", err)
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

	tasks := make([]*models.Task, 0, 50)
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
	// Use ASCII Unit Separator (0x1F) as delimiter to avoid conflicts with user data
	query := `
		SELECT
			t.id,
			t.title,
			t.column_id,
			t.position,
			GROUP_CONCAT(l.id, CHAR(31)) as label_ids,
			GROUP_CONCAT(l.name, CHAR(31)) as label_names,
			GROUP_CONCAT(l.color, CHAR(31)) as label_colors
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

	summaries := make([]*models.TaskSummary, 0, 50)
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

// GetTaskSummariesByProject retrieves all task summaries for a project, grouped by column
// This prevents N+1 queries by fetching all tasks at once
func (r *TaskRepo) GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	// Use ASCII Unit Separator (0x1F) as delimiter to avoid conflicts with user data
	query := `
		SELECT
			t.id,
			t.title,
			t.column_id,
			t.position,
			GROUP_CONCAT(l.id, CHAR(31)) as label_ids,
			GROUP_CONCAT(l.name, CHAR(31)) as label_names,
			GROUP_CONCAT(l.color, CHAR(31)) as label_colors
		FROM tasks t
		INNER JOIN columns c ON t.column_id = c.id
		LEFT JOIN task_labels tl ON t.id = tl.task_id
		LEFT JOIN labels l ON tl.label_id = l.id
		WHERE c.project_id = ?
		GROUP BY t.id, t.title, t.column_id, t.position
		ORDER BY t.position`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query task summaries for project %d: %w", projectID, err)
	}
	defer rows.Close()

	// Group tasks by column_id
	tasksByColumn := make(map[int][]*models.TaskSummary)
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
			return nil, fmt.Errorf("failed to scan task summary for project %d: %w", projectID, err)
		}

		// Parse concatenated label data
		summary.Labels = parseLabelStrings(labelIDsStr, labelNamesStr, labelColorsStr)

		// Append to the appropriate column slice
		tasksByColumn[summary.ColumnID] = append(tasksByColumn[summary.ColumnID], summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task summaries for project %d: %w", projectID, err)
	}
	return tasksByColumn, nil
}

// GetTaskSummariesByProjectFiltered retrieves task summaries for a project that match the search query
// Uses case-insensitive LIKE search on task title
func (r *TaskRepo) GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error) {
	// Use ASCII Unit Separator (0x1F) as delimiter to avoid conflicts with user data
	query := `
		SELECT
			t.id,
			t.title,
			t.column_id,
			t.position,
			GROUP_CONCAT(l.id, CHAR(31)) as label_ids,
			GROUP_CONCAT(l.name, CHAR(31)) as label_names,
			GROUP_CONCAT(l.color, CHAR(31)) as label_colors
		FROM tasks t
		INNER JOIN columns c ON t.column_id = c.id
		LEFT JOIN task_labels tl ON t.id = tl.task_id
		LEFT JOIN labels l ON tl.label_id = l.id
		WHERE c.project_id = ? AND t.title LIKE ?
		GROUP BY t.id, t.title, t.column_id, t.position
		ORDER BY t.position`

	rows, err := r.db.QueryContext(ctx, query, projectID, "%"+searchQuery+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query filtered tasks for project %d: %w", projectID, err)
	}
	defer rows.Close()

	// Group tasks by column_id (same as GetTaskSummariesByProject)
	tasksByColumn := make(map[int][]*models.TaskSummary)
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
			return nil, fmt.Errorf("failed to scan task summary: %w", err)
		}

		// Parse concatenated label data
		summary.Labels = parseLabelStrings(labelIDsStr, labelNamesStr, labelColorsStr)

		// Append to the appropriate column slice
		tasksByColumn[summary.ColumnID] = append(tasksByColumn[summary.ColumnID], summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task summaries: %w", err)
	}
	return tasksByColumn, nil
}

// GetTaskDetail retrieves full task details including description, timestamps, labels, and subtasks
func (r *TaskRepo) GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error) {
	detail := &models.TaskDetail{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, column_id, position, ticket_number, created_at, updated_at
		 FROM tasks WHERE id = ?`,
		taskID,
	).Scan(
		&detail.ID, &detail.Title, &detail.Description,
		&detail.ColumnID, &detail.Position, &detail.TicketNumber, &detail.CreatedAt, &detail.UpdatedAt,
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

	labels := make([]*models.Label, 0, 10)
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

	// Get parent tasks (tasks that depend on this task)
	parentTasks, err := r.GetParentTasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent tasks for task %d: %w", taskID, err)
	}
	detail.ParentTasks = parentTasks

	// Get child tasks (tasks this task depends on)
	childTasks, err := r.GetChildTasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child tasks for task %d: %w", taskID, err)
	}
	detail.ChildTasks = childTasks

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
		return models.ErrAlreadyLastColumn
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
		return models.ErrAlreadyFirstColumn
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

// SwapTaskUp moves a task up in its column by swapping positions with the task above it.
// Returns ErrAlreadyFirstTask if the task is already at position 0.
func (r *TaskRepo) SwapTaskUp(ctx context.Context, taskID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for swapping task %d up: %w", taskID, err)
	}
	defer tx.Rollback()

	// 1. Get the task's current column_id and position
	var columnID, currentPos int
	err = tx.QueryRowContext(ctx,
		`SELECT column_id, position FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&columnID, &currentPos)
	if err != nil {
		return fmt.Errorf("failed to get task %d position: %w", taskID, err)
	}

	// 2. Check if already at top (position 0)
	if currentPos == 0 {
		return models.ErrAlreadyFirstTask
	}

	// 3. Find the task above (position - 1)
	var aboveTaskID int
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM tasks WHERE column_id = ? AND position = ?`,
		columnID, currentPos-1,
	).Scan(&aboveTaskID)
	if err != nil {
		return fmt.Errorf("failed to find task above position %d: %w", currentPos, err)
	}

	// 4. Swap positions using a temporary value to avoid unique constraint violations
	// Set current task to -1 temporarily
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = -1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to set temporary position for task %d: %w", taskID, err)
	}

	// Move the above task down
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		currentPos, aboveTaskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update position for task %d: %w", aboveTaskID, err)
	}

	// Move current task to the above position
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		currentPos-1, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update position for task %d: %w", taskID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for swapping task %d up: %w", taskID, err)
	}
	return nil
}

// SwapTaskDown moves a task down in its column by swapping positions with the task below it.
// Returns ErrAlreadyLastTask if the task is already at the last position.
func (r *TaskRepo) SwapTaskDown(ctx context.Context, taskID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for swapping task %d down: %w", taskID, err)
	}
	defer tx.Rollback()

	// 1. Get the task's current column_id and position
	var columnID, currentPos int
	err = tx.QueryRowContext(ctx,
		`SELECT column_id, position FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&columnID, &currentPos)
	if err != nil {
		return fmt.Errorf("failed to get task %d position: %w", taskID, err)
	}

	// 2. Get max position in the column
	var maxPos int
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(position), 0) FROM tasks WHERE column_id = ?`,
		columnID,
	).Scan(&maxPos)
	if err != nil {
		return fmt.Errorf("failed to get max position in column %d: %w", columnID, err)
	}

	// 3. Check if already at bottom
	if currentPos >= maxPos {
		return models.ErrAlreadyLastTask
	}

	// 4. Find the task below (position + 1)
	var belowTaskID int
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM tasks WHERE column_id = ? AND position = ?`,
		columnID, currentPos+1,
	).Scan(&belowTaskID)
	if err != nil {
		return fmt.Errorf("failed to find task below position %d: %w", currentPos, err)
	}

	// 5. Swap positions using a temporary value
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = -1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to set temporary position for task %d: %w", taskID, err)
	}

	// Move the below task up
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		currentPos, belowTaskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update position for task %d: %w", belowTaskID, err)
	}

	// Move current task to the below position
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		currentPos+1, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update position for task %d: %w", taskID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for swapping task %d down: %w", taskID, err)
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

// GetParentTasks retrieves tasks that depend on this task (parent issues)
// Returns lightweight TaskReference structs with project name for display
func (r *TaskRepo) GetParentTasks(ctx context.Context, taskID int) ([]*models.TaskReference, error) {
	query := `
		SELECT t.id, t.ticket_number, t.title, p.name
		FROM tasks t
		INNER JOIN task_subtasks ts ON t.id = ts.parent_id
		INNER JOIN columns c ON t.column_id = c.id
		INNER JOIN projects p ON c.project_id = p.id
		WHERE ts.child_id = ?
		ORDER BY p.name, t.ticket_number`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query parent tasks for task %d: %w", taskID, err)
	}
	defer rows.Close()

	references := make([]*models.TaskReference, 0, 10)
	for rows.Next() {
		ref := &models.TaskReference{}
		if err := rows.Scan(&ref.ID, &ref.TicketNumber, &ref.Title, &ref.ProjectName); err != nil {
			return nil, fmt.Errorf("failed to scan parent task for task %d: %w", taskID, err)
		}
		references = append(references, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating parent tasks for task %d: %w", taskID, err)
	}
	return references, nil
}

// GetChildTasks retrieves tasks that must be done before this one (child issues)
// Returns lightweight TaskReference structs with project name for display
func (r *TaskRepo) GetChildTasks(ctx context.Context, taskID int) ([]*models.TaskReference, error) {
	query := `
		SELECT t.id, t.ticket_number, t.title, p.name
		FROM tasks t
		INNER JOIN task_subtasks ts ON t.id = ts.child_id
		INNER JOIN columns c ON t.column_id = c.id
		INNER JOIN projects p ON c.project_id = p.id
		WHERE ts.parent_id = ?
		ORDER BY p.name, t.ticket_number`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query child tasks for task %d: %w", taskID, err)
	}
	defer rows.Close()

	references := make([]*models.TaskReference, 0, 10)
	for rows.Next() {
		ref := &models.TaskReference{}
		if err := rows.Scan(&ref.ID, &ref.TicketNumber, &ref.Title, &ref.ProjectName); err != nil {
			return nil, fmt.Errorf("failed to scan child task for task %d: %w", taskID, err)
		}
		references = append(references, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating child tasks for task %d: %w", taskID, err)
	}
	return references, nil
}

// GetTaskReferencesForProject retrieves all task references for a project.
// This is used by the parent/child task pickers to display all available tasks.
// Returns lightweight TaskReference structs with project name for display in PROJ-123 format.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - projectID: The ID of the project to query
//
// Returns:
//   - A slice of TaskReference objects ordered by project name and ticket number
//   - An error if the query fails
func (r *TaskRepo) GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error) {
	query := `
		SELECT t.id, t.ticket_number, t.title, p.name
		FROM tasks t
		INNER JOIN columns c ON t.column_id = c.id
		INNER JOIN projects p ON c.project_id = p.id
		WHERE p.id = ?
		ORDER BY p.name, t.ticket_number`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query task references for project %d: %w", projectID, err)
	}
	defer rows.Close()

	references := make([]*models.TaskReference, 0, 10)
	for rows.Next() {
		ref := &models.TaskReference{}
		if err := rows.Scan(&ref.ID, &ref.TicketNumber, &ref.Title, &ref.ProjectName); err != nil {
			return nil, fmt.Errorf("failed to scan task reference for project %d: %w", projectID, err)
		}
		references = append(references, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task references for project %d: %w", projectID, err)
	}
	return references, nil
}

// AddSubtask creates a parent-child relationship between tasks.
// This establishes a dependency where the parent task blocks on (depends on) the child task.
// In other words: the parent cannot be completed until the child is completed.
//
// CRITICAL - Parameter Ordering:
//   - parentID: The task that will block on the child (the dependent task)
//   - childID: The task that must be completed first (the dependency)
//
// Example: AddSubtask(taskA, taskB) means taskA depends on taskB being completed first.
//
// The function uses INSERT OR IGNORE to prevent duplicate relationships.
func (r *TaskRepo) AddSubtask(ctx context.Context, parentID, childID int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO task_subtasks (parent_id, child_id) VALUES (?, ?)`,
		parentID, childID,
	)
	if err != nil {
		return fmt.Errorf("failed to add subtask relationship (parent: %d, child: %d): %w", parentID, childID, err)
	}
	return nil
}

// RemoveSubtask removes a parent-child relationship between tasks.
// This removes the dependency where the parent task blocks on the child task.
//
// CRITICAL - Parameter Ordering:
//   - parentID: The task that currently blocks on the child (the dependent task)
//   - childID: The task that the parent depends on (the dependency)
//
// Example: RemoveSubtask(taskA, taskB) removes the dependency of taskA on taskB.
func (r *TaskRepo) RemoveSubtask(ctx context.Context, parentID, childID int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM task_subtasks WHERE parent_id = ? AND child_id = ?`,
		parentID, childID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove subtask relationship (parent: %d, child: %d): %w", parentID, childID, err)
	}
	return nil
}

// parseLabelStrings parses label data from GROUP_CONCAT (delimited by ASCII Unit Separator)
func parseLabelStrings(labelIDsStr, labelNamesStr, labelColorsStr sql.NullString) []*models.Label {
	// If no labels exist for this task, return empty slice
	if !labelIDsStr.Valid || !labelNamesStr.Valid || !labelColorsStr.Valid {
		return []*models.Label{}
	}

	// Split the concatenated strings using ASCII Unit Separator (0x1F)
	const delimiter = "\x1F"
	ids := strings.Split(labelIDsStr.String, delimiter)
	names := strings.Split(labelNamesStr.String, delimiter)
	colors := strings.Split(labelColorsStr.String, delimiter)

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
