package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// TaskRepo handles all task-related database operations.
type TaskRepo struct {
	db          *sql.DB
	eventClient *events.Client
}

const defaultTaskCapacity = 50

// CreateTask creates a new task in the database with an auto-assigned ticket number
func (r *TaskRepo) CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
	// Get project ID from column
	projectID, err := getProjectIDFromTable(ctx, r.db, "columns", columnID)
	if err != nil {
		return nil, err
	}

	// Start transaction for ticket number allocation and task creation
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

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

	// Send event notification before commit
	sendEvent(r.eventClient, projectID)

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
	defer closeRows(rows, "GetTasksByColumn")

	tasks := make([]*models.Task, 0, defaultTaskCapacity)
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

// taskSummaryFilter defines filtering options for querying task summaries
type taskSummaryFilter struct {
	filterType  string // "column", "project", or "project_filtered"
	columnID    int
	projectID   int
	searchQuery string
}

// getTaskSummaries is a consolidated helper that retrieves task summaries based on filter criteria.
// Returns a slice of summaries and optionally groups them by column_id for project-level queries.
func (r *TaskRepo) getTaskSummaries(ctx context.Context, filter taskSummaryFilter) ([]*models.TaskSummary, map[int][]*models.TaskSummary, error) {
	var query string
	var args []interface{}
	var groupByColumn bool

	baseSelect := `
		SELECT
			t.id,
			t.title,
			t.column_id,
			t.position,
			ty.description,
			p.description,
			p.color,
			GROUP_CONCAT(l.id, CHAR(31)) as label_ids,
			GROUP_CONCAT(l.name, CHAR(31)) as label_names,
			GROUP_CONCAT(l.color, CHAR(31)) as label_colors
		FROM tasks t`

	switch filter.filterType {
	case "column":
		query = baseSelect + `
		LEFT JOIN types ty ON t.type_id = ty.id
		LEFT JOIN priorities p ON t.priority_id = p.id
		LEFT JOIN task_labels tl ON t.id = tl.task_id
		LEFT JOIN labels l ON tl.label_id = l.id
		WHERE t.column_id = ?
		GROUP BY t.id, t.title, t.column_id, t.position, ty.description, p.description, p.color
		ORDER BY t.position`
		args = []interface{}{filter.columnID}
		groupByColumn = false
	case "project":
		query = baseSelect + `
		INNER JOIN columns c ON t.column_id = c.id
		LEFT JOIN types ty ON t.type_id = ty.id
		LEFT JOIN priorities p ON t.priority_id = p.id
		LEFT JOIN task_labels tl ON t.id = tl.task_id
		LEFT JOIN labels l ON tl.label_id = l.id
		WHERE c.project_id = ?
		GROUP BY t.id, t.title, t.column_id, t.position, ty.description, p.description, p.color
		ORDER BY t.position`
		args = []interface{}{filter.projectID}
		groupByColumn = true
	case "project_filtered":
		query = baseSelect + `
		INNER JOIN columns c ON t.column_id = c.id
		LEFT JOIN types ty ON t.type_id = ty.id
		LEFT JOIN priorities p ON t.priority_id = p.id
		LEFT JOIN task_labels tl ON t.id = tl.task_id
		LEFT JOIN labels l ON tl.label_id = l.id
		WHERE c.project_id = ? AND t.title LIKE ?
		GROUP BY t.id, t.title, t.column_id, t.position, ty.description, p.description, p.color
		ORDER BY t.position`
		args = []interface{}{filter.projectID, "%" + filter.searchQuery + "%"}
		groupByColumn = true
	default:
		return nil, nil, fmt.Errorf("invalid filter type: %s", filter.filterType)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query task summaries: %w", err)
	}
	defer closeRows(rows, "GetTaskSummaries")

	var summaries []*models.TaskSummary
	var tasksByColumn map[int][]*models.TaskSummary

	if groupByColumn {
		tasksByColumn = make(map[int][]*models.TaskSummary)
	} else {
		summaries = make([]*models.TaskSummary, 0, defaultTaskCapacity)
	}

	for rows.Next() {
		var labelIDsStr, labelNamesStr, labelColorsStr sql.NullString
		var typeDesc, priorityDesc, priorityColor sql.NullString
		summary := &models.TaskSummary{}

		if err := rows.Scan(
			&summary.ID,
			&summary.Title,
			&summary.ColumnID,
			&summary.Position,
			&typeDesc,
			&priorityDesc,
			&priorityColor,
			&labelIDsStr,
			&labelNamesStr,
			&labelColorsStr,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan task summary: %w", err)
		}

		// Set type description (default to empty string if null)
		if typeDesc.Valid {
			summary.TypeDescription = typeDesc.String
		}

		// Set priority description and color (default to empty string if null)
		if priorityDesc.Valid {
			summary.PriorityDescription = priorityDesc.String
		}
		if priorityColor.Valid {
			summary.PriorityColor = priorityColor.String
		}

		// Parse concatenated label data
		summary.Labels = parseLabelStrings(labelIDsStr, labelNamesStr, labelColorsStr)

		if groupByColumn {
			tasksByColumn[summary.ColumnID] = append(tasksByColumn[summary.ColumnID], summary)
		} else {
			summaries = append(summaries, summary)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating task summaries: %w", err)
	}

	return summaries, tasksByColumn, nil
}

// GetTaskSummariesByColumn retrieves task summaries for a column, including labels
// Uses a single query with LEFT JOIN to avoid N+1 query pattern
func (r *TaskRepo) GetTaskSummariesByColumn(ctx context.Context, columnID int) ([]*models.TaskSummary, error) {
	summaries, _, err := r.getTaskSummaries(ctx, taskSummaryFilter{
		filterType: "column",
		columnID:   columnID,
	})
	return summaries, err
}

// GetTaskSummariesByProject retrieves all task summaries for a project, grouped by column
// This prevents N+1 queries by fetching all tasks at once
func (r *TaskRepo) GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	_, tasksByColumn, err := r.getTaskSummaries(ctx, taskSummaryFilter{
		filterType: "project",
		projectID:  projectID,
	})
	return tasksByColumn, err
}

// GetTaskSummariesByProjectFiltered retrieves task summaries for a project that match the search query
// Uses case-insensitive LIKE search on task title
func (r *TaskRepo) GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error) {
	_, tasksByColumn, err := r.getTaskSummaries(ctx, taskSummaryFilter{
		filterType:  "project_filtered",
		projectID:   projectID,
		searchQuery: searchQuery,
	})
	return tasksByColumn, err
}

// GetTaskDetail retrieves full task details including description, timestamps, labels, and subtasks
func (r *TaskRepo) GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error) {
	detail := &models.TaskDetail{}
	var typeDesc, priorityDesc, priorityColor sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT t.id, t.title, t.description, t.column_id, t.position, t.ticket_number, t.created_at, t.updated_at, ty.description, p.description, p.color
		 FROM tasks t
		 LEFT JOIN types ty ON t.type_id = ty.id
		 LEFT JOIN priorities p ON t.priority_id = p.id
		 WHERE t.id = ?`,
		taskID,
	).Scan(
		&detail.ID, &detail.Title, &detail.Description,
		&detail.ColumnID, &detail.Position, &detail.TicketNumber, &detail.CreatedAt, &detail.UpdatedAt, &typeDesc, &priorityDesc, &priorityColor,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task detail for task %d: %w", taskID, err)
	}

	// Set type description (default to empty string if null)
	if typeDesc.Valid {
		detail.TypeDescription = typeDesc.String
	}

	// Set priority description and color (default to empty string if null)
	if priorityDesc.Valid {
		detail.PriorityDescription = priorityDesc.String
	}
	if priorityColor.Valid {
		detail.PriorityColor = priorityColor.String
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
	defer closeRows(rows, "GetTaskDetail labels")

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
	// Get projectID for event notification
	var projectID int
	err := r.db.QueryRowContext(ctx,
		`SELECT c.project_id FROM tasks t
		 INNER JOIN columns c ON t.column_id = c.id
		 WHERE t.id = ?`,
		taskID,
	).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("failed to get project for task %d: %w", taskID, err)
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE tasks
		 SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		title, description, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task %d: %w", taskID, err)
	}

	// Send event notification
	sendEvent(r.eventClient, projectID)

	return nil
}

// UpdateTaskPriority updates a task's priority
func (r *TaskRepo) UpdateTaskPriority(ctx context.Context, taskID, priorityID int) error {
	// Get projectID for event notification
	var projectID int
	err := r.db.QueryRowContext(ctx,
		`SELECT c.project_id FROM tasks t
		 INNER JOIN columns c ON t.column_id = c.id
		 WHERE t.id = ?`,
		taskID,
	).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("failed to get project for task %d: %w", taskID, err)
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE tasks
		 SET priority_id = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		priorityID, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task %d priority: %w", taskID, err)
	}

	// Send event notification
	sendEvent(r.eventClient, projectID)

	return nil
}

// UpdateTaskType updates a task's type
func (r *TaskRepo) UpdateTaskType(ctx context.Context, taskID, typeID int) error {
	// Get projectID for event notification
	var projectID int
	err := r.db.QueryRowContext(ctx,
		`SELECT c.project_id FROM tasks t
		 INNER JOIN columns c ON t.column_id = c.id
		 WHERE t.id = ?`,
		taskID,
	).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("failed to get project for task %d: %w", taskID, err)
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE tasks
		 SET type_id = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		typeID, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task %d type: %w", taskID, err)
	}

	// Send event notification
	sendEvent(r.eventClient, projectID)

	return nil
}

// moveTaskToColumn is a consolidated helper that moves a task to a column.
// For moveType "next" or "prev", it uses the linked list to find the target column.
// For moveType "direct", targetColumnID is used directly (pass 0 for targetColumnID when using "next"/"prev").
func (r *TaskRepo) moveTaskToColumn(ctx context.Context, taskID int, moveType string, targetColumnID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for moving task %d: %w", taskID, err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var finalTargetColumnID int
	var currentColumnID int

	switch moveType {
	case "next", "prev":
		// Get the task's current column_id
		err = tx.QueryRowContext(ctx,
			`SELECT column_id FROM tasks WHERE id = ?`,
			taskID,
		).Scan(&currentColumnID)
		if err != nil {
			return fmt.Errorf("failed to get current column for task %d: %w", taskID, err)
		}

		// Get the adjacent column's ID
		var adjacentColumnID sql.NullInt64
		var columnField string
		switch moveType {
		case "next":
			columnField = "next_id"
		case "prev":
			columnField = "prev_id"
		}

		err = tx.QueryRowContext(ctx,
			fmt.Sprintf(`SELECT %s FROM columns WHERE id = ?`, columnField),
			currentColumnID,
		).Scan(&adjacentColumnID)
		if err != nil {
			return fmt.Errorf("failed to get %s column for column %d: %w", moveType, currentColumnID, err)
		}

		// Check if there's an adjacent column
		if !adjacentColumnID.Valid {
			switch moveType {
			case "next":
				return models.ErrAlreadyLastColumn
			case "prev":
				return models.ErrAlreadyFirstColumn
			}
		}

		finalTargetColumnID = int(adjacentColumnID.Int64)
	case "direct":
		// Verify the target column exists
		var exists int
		err = tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM columns WHERE id = ?`,
			targetColumnID,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to verify target column %d: %w", targetColumnID, err)
		}
		if exists == 0 {
			return fmt.Errorf("target column %d does not exist", targetColumnID)
		}
		finalTargetColumnID = targetColumnID
	default:
		return fmt.Errorf("invalid move type: %s", moveType)
	}

	// Get the number of tasks in the target column to determine position
	var taskCount int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tasks WHERE column_id = ?`,
		finalTargetColumnID,
	).Scan(&taskCount)
	if err != nil {
		return fmt.Errorf("failed to count tasks in target column %d: %w", finalTargetColumnID, err)
	}

	// Move the task to the target column (append to end)
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks
		 SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		finalTargetColumnID, taskCount, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to move task %d to column %d: %w", taskID, finalTargetColumnID, err)
	}

	// Get projectID for event notification
	var projectID int
	var columnIDForProject int
	if moveType == "direct" {
		columnIDForProject = targetColumnID
	} else {
		columnIDForProject = currentColumnID
	}
	err = tx.QueryRowContext(ctx,
		`SELECT project_id FROM columns WHERE id = ?`,
		columnIDForProject,
	).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("failed to get project for column %d: %w", columnIDForProject, err)
	}

	// Send event notification before commit
	sendEvent(r.eventClient, projectID)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for moving task %d: %w", taskID, err)
	}
	return nil
}

// MoveTaskToNextColumn moves a task from its current column to the next column in the linked list
// Returns an error if the task doesn't exist or if there's no next column
func (r *TaskRepo) MoveTaskToNextColumn(ctx context.Context, taskID int) error {
	return r.moveTaskToColumn(ctx, taskID, "next", 0)
}

// MoveTaskToPrevColumn moves a task from its current column to the previous column in the linked list
// Returns an error if the task doesn't exist or if there's no previous column
func (r *TaskRepo) MoveTaskToPrevColumn(ctx context.Context, taskID int) error {
	return r.moveTaskToColumn(ctx, taskID, "prev", 0)
}

// MoveTaskToColumn moves a task to a specific target column.
// The task is appended to the end of the target column.
func (r *TaskRepo) MoveTaskToColumn(ctx context.Context, taskID int, targetColumnID int) error {
	return r.moveTaskToColumn(ctx, taskID, "direct", targetColumnID)
}

// swapTask is a consolidated helper that swaps a task's position with an adjacent task.
// direction should be "up" or "down".
func (r *TaskRepo) swapTask(ctx context.Context, taskID int, direction string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for swapping task %d %s: %w", taskID, direction, err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// Get the task's current column_id and position
	var columnID, currentPos int
	err = tx.QueryRowContext(ctx,
		`SELECT column_id, position FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&columnID, &currentPos)
	if err != nil {
		return fmt.Errorf("failed to get task %d position: %w", taskID, err)
	}

	var adjacentTaskID, newPos, adjacentPos int

	switch direction {
	case "up":
		// Find the task above (next smaller position)
		err = tx.QueryRowContext(ctx,
			`SELECT id, position FROM tasks
			 WHERE column_id = ? AND position < ?
			 ORDER BY position DESC LIMIT 1`,
			columnID, currentPos,
		).Scan(&adjacentTaskID, &adjacentPos)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.ErrAlreadyFirstTask
			}
			return fmt.Errorf("failed to find task above position %d: %w", currentPos, err)
		}
		newPos = adjacentPos
	case "down":
		// Find the task below (next larger position)
		err = tx.QueryRowContext(ctx,
			`SELECT id, position FROM tasks
			 WHERE column_id = ? AND position > ?
			 ORDER BY position ASC LIMIT 1`,
			columnID, currentPos,
		).Scan(&adjacentTaskID, &adjacentPos)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.ErrAlreadyLastTask
			}
			return fmt.Errorf("failed to find task below position %d: %w", currentPos, err)
		}
		newPos = adjacentPos
	default:
		return fmt.Errorf("invalid direction: %s", direction)
	}

	// Swap positions using a temporary value to avoid unique constraint violations
	// Set current task to -1 temporarily
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = -1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to set temporary position for task %d: %w", taskID, err)
	}

	// Move the adjacent task to current position
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		currentPos, adjacentTaskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update position for task %d: %w", adjacentTaskID, err)
	}

	// Move current task to new position
	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET position = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		newPos, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update position for task %d: %w", taskID, err)
	}

	// Get projectID for event notification
	var projectID int
	err = tx.QueryRowContext(ctx,
		`SELECT project_id FROM columns WHERE id = ?`,
		columnID,
	).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("failed to get project for column %d: %w", columnID, err)
	}

	// Send event notification before commit
	sendEvent(r.eventClient, projectID)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for swapping task %d %s: %w", taskID, direction, err)
	}
	return nil
}

// SwapTaskUp moves a task up in its column by swapping positions with the task above it.
// Returns ErrAlreadyFirstTask if the task is already at position 0
func (r *TaskRepo) SwapTaskUp(ctx context.Context, taskID int) error {
	return r.swapTask(ctx, taskID, "up")
}

// SwapTaskDown moves a task down in its column by swapping positions with the task below it.
// Returns ErrAlreadyLastTask if the task is already at the last position.
func (r *TaskRepo) SwapTaskDown(ctx context.Context, taskID int) error {
	return r.swapTask(ctx, taskID, "down")
}

// DeleteTask removes a task from the database
func (r *TaskRepo) DeleteTask(ctx context.Context, taskID int) error {
	// Get projectID before deleting
	var projectID int
	err := r.db.QueryRowContext(ctx,
		`SELECT c.project_id FROM tasks t
		 INNER JOIN columns c ON t.column_id = c.id
		 WHERE t.id = ?`,
		taskID,
	).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("failed to get project for task %d: %w", taskID, err)
	}

	_, err = r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task %d: %w", taskID, err)
	}

	// Send event notification
	sendEvent(r.eventClient, projectID)

	return nil
}

// getTaskReferences is a consolidated helper that retrieves task references based on the reference type.
// refType should be "parent", "child", or "project".
func (r *TaskRepo) getTaskReferences(ctx context.Context, refType string, id int) ([]*models.TaskReference, error) {
	var query string
	var errorContext string

	switch refType {
	case "parent":
		query = `
			SELECT t.id, t.ticket_number, t.title, p.name
			FROM tasks t
			INNER JOIN task_subtasks ts ON t.id = ts.parent_id
			INNER JOIN columns c ON t.column_id = c.id
			INNER JOIN projects p ON c.project_id = p.id
			WHERE ts.child_id = ?
			ORDER BY p.name, t.ticket_number`
		errorContext = "parent tasks for task"
	case "child":
		query = `
			SELECT t.id, t.ticket_number, t.title, p.name
			FROM tasks t
			INNER JOIN task_subtasks ts ON t.id = ts.child_id
			INNER JOIN columns c ON t.column_id = c.id
			INNER JOIN projects p ON c.project_id = p.id
			WHERE ts.parent_id = ?
			ORDER BY p.name, t.ticket_number`
		errorContext = "child tasks for task"
	case "project":
		query = `
			SELECT t.id, t.ticket_number, t.title, p.name
			FROM tasks t
			INNER JOIN columns c ON t.column_id = c.id
			INNER JOIN projects p ON c.project_id = p.id
			WHERE p.id = ?
			ORDER BY p.name, t.ticket_number`
		errorContext = "task references for project"
	default:
		return nil, fmt.Errorf("invalid reference type: %s", refType)
	}

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s %d: %w", errorContext, id, err)
	}
	defer closeRows(rows, fmt.Sprintf("GetTaskReferences: %s", refType))

	references := make([]*models.TaskReference, 0, 10)
	for rows.Next() {
		ref := &models.TaskReference{}
		if err := rows.Scan(&ref.ID, &ref.TicketNumber, &ref.Title, &ref.ProjectName); err != nil {
			return nil, fmt.Errorf("failed to scan %s %d: %w", errorContext, id, err)
		}
		references = append(references, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating %s %d: %w", errorContext, id, err)
	}
	return references, nil
}

// GetParentTasks retrieves tasks that depend on this task (parent issues)
// Returns lightweight TaskReference structs with project name for display
func (r *TaskRepo) GetParentTasks(ctx context.Context, taskID int) ([]*models.TaskReference, error) {
	return r.getTaskReferences(ctx, "parent", taskID)
}

// GetChildTasks retrieves tasks that must be done before this one (child issues)
// Returns lightweight TaskReference structs with project name for display
func (r *TaskRepo) GetChildTasks(ctx context.Context, taskID int) ([]*models.TaskReference, error) {
	return r.getTaskReferences(ctx, "child", taskID)
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
	return r.getTaskReferences(ctx, "project", projectID)
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
	// Get projectID for event notification
	var projectID int
	err := r.db.QueryRowContext(ctx,
		`SELECT c.project_id FROM tasks t
		 INNER JOIN columns c ON t.column_id = c.id
		 WHERE t.id = ?`,
		parentID,
	).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("failed to get project for task %d: %w", parentID, err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO task_subtasks (parent_id, child_id) VALUES (?, ?)`,
		parentID, childID,
	)
	if err != nil {
		return fmt.Errorf("failed to add subtask relationship (parent: %d, child: %d): %w", parentID, childID, err)
	}

	// Send event notification
	sendEvent(r.eventClient, projectID)

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
	// Get projectID before removing
	var projectID int
	err := r.db.QueryRowContext(ctx,
		`SELECT c.project_id FROM tasks t
		 INNER JOIN columns c ON t.column_id = c.id
		 WHERE t.id = ?`,
		parentID,
	).Scan(&projectID)
	if err != nil {
		return fmt.Errorf("failed to get project for task %d: %w", parentID, err)
	}

	_, err = r.db.ExecContext(ctx,
		`DELETE FROM task_subtasks WHERE parent_id = ? AND child_id = ?`,
		parentID, childID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove subtask relationship (parent: %d, child: %d): %w", parentID, childID, err)
	}

	// Send event notification
	sendEvent(r.eventClient, projectID)

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
