package database

import (
	"context"
	"database/sql"

	"github.com/thenoetrevino/paso/internal/models"
)

// ============================================================================
// Column Operations
// ============================================================================

// CreateColumn creates a new column in the database for a specific project
// If afterColumnID is nil, the column is appended to the end of the project's list
// Otherwise, it's inserted after the specified column
func CreateColumn(ctx context.Context, db *sql.DB, name string, projectID int, afterColumnID *int) (*models.Column, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var prevID *int
	var nextID *int

	if afterColumnID == nil {
		// Append to end: find tail (column where next_id IS NULL) for this project
		var tailID sql.NullInt64
		err = tx.QueryRowContext(ctx, `SELECT id FROM columns WHERE next_id IS NULL AND project_id = ? LIMIT 1`, projectID).Scan(&tailID)
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
		err = tx.QueryRowContext(ctx, `SELECT next_id FROM columns WHERE id = ?`, *afterColumnID).Scan(&currentNextID)
		if err != nil {
			return nil, err
		}

		if currentNextID.Valid {
			nextIDInt := int(currentNextID.Int64)
			nextID = &nextIDInt
		}
	}

	// Create the new column
	result, err := tx.ExecContext(ctx,
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
		_, err = tx.ExecContext(ctx, `UPDATE columns SET next_id = ? WHERE id = ?`, newIDInt, *prevID)
		if err != nil {
			return nil, err
		}
	}

	// Update the next column's prev_id to point to the new column
	if nextID != nil {
		_, err = tx.ExecContext(ctx, `UPDATE columns SET prev_id = ? WHERE id = ?`, newIDInt, *nextID)
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
func GetColumnsByProject(ctx context.Context, db *sql.DB, projectID int) ([]*models.Column, error) {
	// 1. Find the head (column where prev_id IS NULL) for this project
	var headID sql.NullInt64
	err := db.QueryRowContext(ctx, `SELECT id FROM columns WHERE prev_id IS NULL AND project_id = ? LIMIT 1`, projectID).Scan(&headID)
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

		err := db.QueryRowContext(ctx,
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
func GetAllColumns(ctx context.Context, db *sql.DB) ([]*models.Column, error) {
	// Get the first project and return its columns
	// For backward compatibility with tests
	var projectID int
	err := db.QueryRowContext(ctx, `SELECT id FROM projects ORDER BY id LIMIT 1`).Scan(&projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.Column{}, nil
		}
		return nil, err
	}
	return GetColumnsByProject(ctx, db, projectID)
}

// GetColumnByID retrieves a column by its ID
func GetColumnByID(ctx context.Context, db *sql.DB, columnID int) (*models.Column, error) {
	column := &models.Column{}
	err := db.QueryRowContext(ctx,
		`SELECT id, name, project_id, prev_id, next_id FROM columns WHERE id = ?`,
		columnID,
	).Scan(&column.ID, &column.Name, &column.ProjectID, &column.PrevID, &column.NextID)
	if err != nil {
		return nil, err
	}
	return column, nil
}

// UpdateColumnName updates the name of an existing column
func UpdateColumnName(ctx context.Context, db *sql.DB, columnID int, name string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE columns SET name = ? WHERE id = ?`,
		name, columnID,
	)
	return err
}

// DeleteColumn removes a column and all its tasks from the database
// This operation maintains the linked list structure by updating adjacent columns
func DeleteColumn(ctx context.Context, db *sql.DB, columnID int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get the column's prev_id and next_id
	var prevID, nextID sql.NullInt64
	err = tx.QueryRowContext(ctx,
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
		_, err = tx.ExecContext(ctx, `UPDATE columns SET next_id = ? WHERE id = ?`, newNextID, prevID.Int64)
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
		_, err = tx.ExecContext(ctx, `UPDATE columns SET prev_id = ? WHERE id = ?`, newPrevID, nextID.Int64)
		if err != nil {
			return err
		}
	}

	// 3. Delete all tasks in the column
	_, err = tx.ExecContext(ctx, "DELETE FROM tasks WHERE column_id = ?", columnID)
	if err != nil {
		return err
	}

	// 4. Delete the column
	_, err = tx.ExecContext(ctx, "DELETE FROM columns WHERE id = ?", columnID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
