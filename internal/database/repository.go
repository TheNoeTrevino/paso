package database

import (
	"database/sql"

	"github.com/thenoetrevino/paso/internal/models"
)

// CreateColumn creates a new column in the database
func CreateColumn(db *sql.DB, name string, position int) (*models.Column, error) {
	result, err := db.Exec(
		"INSERT INTO columns (name, position) VALUES (?, ?)",
		name, position,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Column{
		ID:       int(id),
		Name:     name,
		Position: position,
	}, nil
}

// GetAllColumns retrieves all columns from the database, ordered by position
func GetAllColumns(db *sql.DB) ([]*models.Column, error) {
	rows, err := db.Query("SELECT id, name, position FROM columns ORDER BY position")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []*models.Column
	for rows.Next() {
		col := &models.Column{}
		if err := rows.Scan(&col.ID, &col.Name, &col.Position); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
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
