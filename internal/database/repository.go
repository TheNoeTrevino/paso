package database

import (
	"database/sql"

	"github.com/thenoetrevino/paso/internal/models"
)

// ============================================================================
// Project Operations
// ============================================================================

// CreateProject creates a new project with default columns (Todo, In Progress, Done)
func CreateProject(db *sql.DB, name, description string) (*models.Project, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert the project
	result, err := tx.Exec(
		`INSERT INTO projects (name, description) VALUES (?, ?)`,
		name, description,
	)
	if err != nil {
		return nil, err
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Initialize the project counter
	_, err = tx.Exec(
		`INSERT INTO project_counters (project_id, next_ticket_number) VALUES (?, 1)`,
		projectID,
	)
	if err != nil {
		return nil, err
	}

	// Create default columns for the project (as a linked list)
	defaultColumns := []string{"Todo", "In Progress", "Done"}
	var prevColID *int

	for _, colName := range defaultColumns {
		var colResult sql.Result
		if prevColID == nil {
			// First column: no prev_id
			colResult, err = tx.Exec(
				`INSERT INTO columns (name, project_id, prev_id, next_id) VALUES (?, ?, NULL, NULL)`,
				colName, projectID,
			)
		} else {
			// Subsequent columns: set prev_id
			colResult, err = tx.Exec(
				`INSERT INTO columns (name, project_id, prev_id, next_id) VALUES (?, ?, ?, NULL)`,
				colName, projectID, *prevColID,
			)
		}
		if err != nil {
			return nil, err
		}

		colID, err := colResult.LastInsertId()
		if err != nil {
			return nil, err
		}
		colIDInt := int(colID)

		// Update the previous column's next_id
		if prevColID != nil {
			_, err = tx.Exec(`UPDATE columns SET next_id = ? WHERE id = ?`, colIDInt, *prevColID)
			if err != nil {
				return nil, err
			}
		}

		prevColID = &colIDInt
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Retrieve the created project
	return GetProjectByID(db, int(projectID))
}

// GetProjectByID retrieves a project by its ID
func GetProjectByID(db *sql.DB, id int) (*models.Project, error) {
	project := &models.Project{}
	err := db.QueryRow(
		`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`,
		id,
	).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// GetAllProjects retrieves all projects ordered by name
func GetAllProjects(db *sql.DB) ([]*models.Project, error) {
	rows, err := db.Query(`SELECT id, name, description, created_at, updated_at FROM projects ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// UpdateProject updates a project's name and description
func UpdateProject(db *sql.DB, id int, name, description string) error {
	_, err := db.Exec(
		`UPDATE projects SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, description, id,
	)
	return err
}

// DeleteProject removes a project and all its columns and tasks (cascade)
func DeleteProject(db *sql.DB, id int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get all columns for this project
	rows, err := tx.Query(`SELECT id FROM columns WHERE project_id = ?`, id)
	if err != nil {
		return err
	}

	var columnIDs []int
	for rows.Next() {
		var colID int
		if err := rows.Scan(&colID); err != nil {
			rows.Close()
			return err
		}
		columnIDs = append(columnIDs, colID)
	}
	rows.Close()

	// Delete all tasks in those columns
	for _, colID := range columnIDs {
		_, err = tx.Exec(`DELETE FROM tasks WHERE column_id = ?`, colID)
		if err != nil {
			return err
		}
	}

	// Delete all columns for this project
	_, err = tx.Exec(`DELETE FROM columns WHERE project_id = ?`, id)
	if err != nil {
		return err
	}

	// Delete the project counter
	_, err = tx.Exec(`DELETE FROM project_counters WHERE project_id = ?`, id)
	if err != nil {
		return err
	}

	// Delete the project
	_, err = tx.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetProjectTaskCount returns the total number of tasks in a project
func GetProjectTaskCount(db *sql.DB, projectID int) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM tasks t
		JOIN columns c ON t.column_id = c.id
		WHERE c.project_id = ?
	`, projectID).Scan(&count)
	return count, err
}

// GetNextTicketNumber returns and increments the next ticket number for a project
func GetNextTicketNumber(db *sql.DB, projectID int) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Get current counter
	var ticketNumber int
	err = tx.QueryRow(
		`SELECT next_ticket_number FROM project_counters WHERE project_id = ?`,
		projectID,
	).Scan(&ticketNumber)
	if err != nil {
		return 0, err
	}

	// Increment counter
	_, err = tx.Exec(
		`UPDATE project_counters SET next_ticket_number = next_ticket_number + 1 WHERE project_id = ?`,
		projectID,
	)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return ticketNumber, nil
}

// CreateColumn creates a new column in the database for a specific project
// If afterColumnID is nil, the column is appended to the end of the project's list
// Otherwise, it's inserted after the specified column
func CreateColumn(db *sql.DB, name string, projectID int, afterColumnID *int) (*models.Column, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var prevID *int
	var nextID *int

	if afterColumnID == nil {
		// Append to end: find tail (column where next_id IS NULL) for this project
		var tailID sql.NullInt64
		err = tx.QueryRow(`SELECT id FROM columns WHERE next_id IS NULL AND project_id = ? LIMIT 1`, projectID).Scan(&tailID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if tailID.Valid {
			// There's a tail, insert after it
			tailIDInt := int(tailID.Int64)
			prevID = &tailIDInt
			nextID = nil
		} else {
			// No columns exist for this project, this will be the head
			prevID = nil
			nextID = nil
		}
	} else {
		// Insert after specified column
		prevID = afterColumnID

		// Get the next_id of the column we're inserting after
		var currentNextID sql.NullInt64
		err = tx.QueryRow(`SELECT next_id FROM columns WHERE id = ?`, *afterColumnID).Scan(&currentNextID)
		if err != nil {
			return nil, err
		}

		if currentNextID.Valid {
			nextIDInt := int(currentNextID.Int64)
			nextID = &nextIDInt
		}
	}

	// Create the new column
	result, err := tx.Exec(
		`INSERT INTO columns (name, project_id, prev_id, next_id) VALUES (?, ?, ?, ?)`,
		name, projectID, prevID, nextID,
	)
	if err != nil {
		return nil, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	newIDInt := int(newID)

	// Update the previous column's next_id to point to the new column
	if prevID != nil {
		_, err = tx.Exec(`UPDATE columns SET next_id = ? WHERE id = ?`, newIDInt, *prevID)
		if err != nil {
			return nil, err
		}
	}

	// Update the next column's prev_id to point to the new column
	if nextID != nil {
		_, err = tx.Exec(`UPDATE columns SET prev_id = ? WHERE id = ?`, newIDInt, *nextID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Column{
		ID:        newIDInt,
		Name:      name,
		ProjectID: projectID,
		PrevID:    prevID,
		NextID:    nextID,
	}, nil
}

// GetColumnsByProject retrieves all columns for a specific project by traversing the linked list
// Returns columns in order from head to tail
func GetColumnsByProject(db *sql.DB, projectID int) ([]*models.Column, error) {
	// 1. Find the head (column where prev_id IS NULL) for this project
	var headID sql.NullInt64
	err := db.QueryRow(`SELECT id FROM columns WHERE prev_id IS NULL AND project_id = ? LIMIT 1`, projectID).Scan(&headID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No columns exist for this project
			return []*models.Column{}, nil
		}
		return nil, err
	}

	if !headID.Valid {
		// No head found, return empty list
		return []*models.Column{}, nil
	}

	// 2. Traverse the linked list using next_id
	var columns []*models.Column
	currentID := int(headID.Int64)

	for {
		// Get the current column
		col := &models.Column{}
		var prevID, nextID sql.NullInt64

		err := db.QueryRow(
			`SELECT id, name, project_id, prev_id, next_id FROM columns WHERE id = ?`,
			currentID,
		).Scan(&col.ID, &col.Name, &col.ProjectID, &prevID, &nextID)

		if err != nil {
			return nil, err
		}

		// Set pointer fields
		if prevID.Valid {
			prevIDInt := int(prevID.Int64)
			col.PrevID = &prevIDInt
		}
		if nextID.Valid {
			nextIDInt := int(nextID.Int64)
			col.NextID = &nextIDInt
		}

		// Add column to result
		columns = append(columns, col)

		// Move to next column
		if !nextID.Valid {
			// We've reached the tail
			break
		}
		currentID = int(nextID.Int64)
	}

	return columns, nil
}

// GetAllColumns retrieves all columns from the database by traversing the linked list
// This is a convenience function that gets columns for all projects (used in tests)
// Returns columns in order from head to tail
func GetAllColumns(db *sql.DB) ([]*models.Column, error) {
	// Get the first project and return its columns
	// For backward compatibility with tests
	var projectID int
	err := db.QueryRow(`SELECT id FROM projects ORDER BY id LIMIT 1`).Scan(&projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.Column{}, nil
		}
		return nil, err
	}
	return GetColumnsByProject(db, projectID)
}

// CreateTask creates a new task in the database
func CreateTask(db *sql.DB, title, description string, columnID, position int) (*models.Task, error) {
	result, err := db.Exec(
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
	err = db.QueryRow(
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
func GetTasksByColumn(db *sql.DB, columnID int) ([]*models.Task, error) {
	rows, err := db.Query(
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
func UpdateTaskColumn(db *sql.DB, taskID, newColumnID, newPosition int) error {
	_, err := db.Exec(
		`UPDATE tasks
		 SET column_id = ?, position = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		newColumnID, newPosition, taskID,
	)
	return err
}

// DeleteTask removes a task from the database
func DeleteTask(db *sql.DB, taskID int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	return err
}

// UpdateTaskTitle updates the title of an existing task
func UpdateTaskTitle(db *sql.DB, taskID int, title string) error {
	_, err := db.Exec(
		`UPDATE tasks
		 SET title = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		title, taskID,
	)
	return err
}

// UpdateColumnName updates the name of an existing column
func UpdateColumnName(db *sql.DB, columnID int, name string) error {
	_, err := db.Exec(
		`UPDATE columns SET name = ? WHERE id = ?`,
		name, columnID,
	)
	return err
}

// DeleteColumn removes a column and all its tasks from the database
// This operation maintains the linked list structure by updating adjacent columns
func DeleteColumn(db *sql.DB, columnID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get the column's prev_id and next_id
	var prevID, nextID sql.NullInt64
	err = tx.QueryRow(
		`SELECT prev_id, next_id FROM columns WHERE id = ?`,
		columnID,
	).Scan(&prevID, &nextID)
	if err != nil {
		return err
	}

	// 2. Update adjacent columns to maintain linked list
	// If there's a previous column, update its next_id
	if prevID.Valid {
		var newNextID interface{}
		if nextID.Valid {
			newNextID = nextID.Int64
		} else {
			newNextID = nil
		}
		_, err = tx.Exec(`UPDATE columns SET next_id = ? WHERE id = ?`, newNextID, prevID.Int64)
		if err != nil {
			return err
		}
	}

	// If there's a next column, update its prev_id
	if nextID.Valid {
		var newPrevID interface{}
		if prevID.Valid {
			newPrevID = prevID.Int64
		} else {
			newPrevID = nil
		}
		_, err = tx.Exec(`UPDATE columns SET prev_id = ? WHERE id = ?`, newPrevID, nextID.Int64)
		if err != nil {
			return err
		}
	}

	// 3. Delete all tasks in the column
	_, err = tx.Exec("DELETE FROM tasks WHERE column_id = ?", columnID)
	if err != nil {
		return err
	}

	// 4. Delete the column
	_, err = tx.Exec("DELETE FROM columns WHERE id = ?", columnID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetTaskCountByColumn returns the number of tasks in a specific column
func GetTaskCountByColumn(db *sql.DB, columnID int) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE column_id = ?", columnID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// MoveTask moves a task to a different column and position using a transaction
// This is a wrapper around UpdateTaskColumn that provides transactional guarantees
func MoveTask(db *sql.DB, taskID, newColumnID, newPosition int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update task's column and position
	_, err = tx.Exec(
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
func MoveTaskToNextColumn(db *sql.DB, taskID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get the task's current column_id
	var currentColumnID int
	err = tx.QueryRow(
		`SELECT column_id FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&currentColumnID)
	if err != nil {
		return err
	}

	// 2. Get the next column's ID
	var nextColumnID sql.NullInt64
	err = tx.QueryRow(
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
	err = tx.QueryRow(
		`SELECT COUNT(*) FROM tasks WHERE column_id = ?`,
		nextColumnID.Int64,
	).Scan(&taskCount)
	if err != nil {
		return err
	}

	// 5. Move the task to the next column (append to end)
	_, err = tx.Exec(
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
func MoveTaskToPrevColumn(db *sql.DB, taskID int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get the task's current column_id
	var currentColumnID int
	err = tx.QueryRow(
		`SELECT column_id FROM tasks WHERE id = ?`,
		taskID,
	).Scan(&currentColumnID)
	if err != nil {
		return err
	}

	// 2. Get the previous column's ID
	var prevColumnID sql.NullInt64
	err = tx.QueryRow(
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
	err = tx.QueryRow(
		`SELECT COUNT(*) FROM tasks WHERE column_id = ?`,
		prevColumnID.Int64,
	).Scan(&taskCount)
	if err != nil {
		return err
	}

	// 5. Move the task to the previous column (append to end)
	_, err = tx.Exec(
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
// Label Operations
// ============================================================================

// CreateLabel creates a new label in the database for a specific project
func CreateLabel(db *sql.DB, projectID int, name, color string) (*models.Label, error) {
	result, err := db.Exec(
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
func GetAllLabels(db *sql.DB) ([]*models.Label, error) {
	rows, err := db.Query(`SELECT id, name, color, project_id FROM labels ORDER BY name`)
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
func GetLabelsByProject(db *sql.DB, projectID int) ([]*models.Label, error) {
	rows, err := db.Query(
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
func UpdateLabel(db *sql.DB, labelID int, name, color string) error {
	_, err := db.Exec(
		`UPDATE labels SET name = ?, color = ? WHERE id = ?`,
		name, color, labelID,
	)
	return err
}

// DeleteLabel removes a label from the database (cascade removes task associations)
func DeleteLabel(db *sql.DB, labelID int) error {
	_, err := db.Exec("DELETE FROM labels WHERE id = ?", labelID)
	return err
}

// AddLabelToTask associates a label with a task
func AddLabelToTask(db *sql.DB, taskID, labelID int) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO task_labels (task_id, label_id) VALUES (?, ?)`,
		taskID, labelID,
	)
	return err
}

// RemoveLabelFromTask removes the association between a label and a task
func RemoveLabelFromTask(db *sql.DB, taskID, labelID int) error {
	_, err := db.Exec(
		`DELETE FROM task_labels WHERE task_id = ? AND label_id = ?`,
		taskID, labelID,
	)
	return err
}

// GetLabelsForTask retrieves all labels associated with a task
func GetLabelsForTask(db *sql.DB, taskID int) ([]*models.Label, error) {
	rows, err := db.Query(`
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
func SetTaskLabels(db *sql.DB, taskID int, labelIDs []int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove all existing labels
	_, err = tx.Exec(`DELETE FROM task_labels WHERE task_id = ?`, taskID)
	if err != nil {
		return err
	}

	// Add new labels
	for _, labelID := range labelIDs {
		_, err = tx.Exec(
			`INSERT INTO task_labels (task_id, label_id) VALUES (?, ?)`,
			taskID, labelID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ============================================================================
// Task Summary/Detail Operations (DTOs with Labels)
// ============================================================================

// GetTaskSummariesByColumn retrieves task summaries for a column, including labels
func GetTaskSummariesByColumn(db *sql.DB, columnID int) ([]*models.TaskSummary, error) {
	// Get tasks
	rows, err := db.Query(
		`SELECT id, title, column_id, position
		 FROM tasks
		 WHERE column_id = ?
		 ORDER BY position`,
		columnID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.TaskSummary
	for rows.Next() {
		summary := &models.TaskSummary{}
		if err := rows.Scan(&summary.ID, &summary.Title, &summary.ColumnID, &summary.Position); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get labels for each task
	for _, summary := range summaries {
		labels, err := GetLabelsForTask(db, summary.ID)
		if err != nil {
			return nil, err
		}
		summary.Labels = labels
	}

	return summaries, nil
}

// GetTaskDetail retrieves full task details including description, timestamps, and labels
func GetTaskDetail(db *sql.DB, taskID int) (*models.TaskDetail, error) {
	detail := &models.TaskDetail{}
	err := db.QueryRow(
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
	labels, err := GetLabelsForTask(db, taskID)
	if err != nil {
		return nil, err
	}
	detail.Labels = labels

	return detail, nil
}

// GetColumnByID retrieves a single column by its ID
func GetColumnByID(db *sql.DB, columnID int) (*models.Column, error) {
	column := &models.Column{}
	err := db.QueryRow(
		`SELECT id, name, project_id, prev_id, next_id FROM columns WHERE id = ?`,
		columnID,
	).Scan(&column.ID, &column.Name, &column.ProjectID, &column.PrevID, &column.NextID)
	if err != nil {
		return nil, err
	}
	return column, nil
}

// UpdateTask updates a task's title and description
func UpdateTask(db *sql.DB, taskID int, title, description string) error {
	_, err := db.Exec(
		`UPDATE tasks
		 SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		title, description, taskID,
	)
	return err
}
