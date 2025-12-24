package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// ColumnRepo handles all column-related database operations.
type ColumnRepo struct {
	db          *sql.DB
	eventClient events.EventPublisher
}

// CreateColumn creates a new column in the database for a specific project
// If afterColumnID is nil, the column is appended to the end of the project's list
// Otherwise, it's inserted after the specified column
func (r *ColumnRepo) CreateColumn(ctx context.Context, name string, projectID int, afterColumnID *int) (*models.Column, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var prevID *int
	var nextID *int

	if afterColumnID == nil {
		// Append to end: find tail (column where next_id IS NULL) for this project
		var tailID sql.NullInt64
		err = tx.QueryRowContext(ctx, `SELECT id FROM columns WHERE next_id IS NULL AND project_id = ? LIMIT 1`, projectID).Scan(&tailID)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to find tail column for project %d: %w", projectID, err)
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
			return nil, fmt.Errorf("failed to get next_id for column %d: %w", *afterColumnID, err)
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
		return nil, fmt.Errorf("failed to insert column: %w", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	newIDInt := int(newID)

	// Update the previous column's next_id to point to the new column
	if prevID != nil {
		_, err = tx.ExecContext(ctx, `UPDATE columns SET next_id = ? WHERE id = ?`, newIDInt, *prevID)
		if err != nil {
			return nil, fmt.Errorf("failed to update prev column %d: %w", *prevID, err)
		}
	}

	// Update the next column's prev_id to point to the new column
	if nextID != nil {
		_, err = tx.ExecContext(ctx, `UPDATE columns SET prev_id = ? WHERE id = ?`, newIDInt, *nextID)
		if err != nil {
			return nil, fmt.Errorf("failed to update next column %d: %w", *nextID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Send event notification after commit
	sendEvent(r.eventClient, projectID)

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
func (r *ColumnRepo) GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error) {
	// Fetch ALL columns for the project in a single query (fixes N+1 problem)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, project_id, prev_id, next_id FROM columns WHERE project_id = ?`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("querying columns for project: %w", err)
	}
	defer closeRows(rows, "GetColumnsByProject")

	// Build a map for O(1) lookups and find the head
	columnMap := make(map[int]*models.Column)
	var headID *int

	for rows.Next() {
		col := &models.Column{}
		var prevID, nextID sql.NullInt64

		if err := rows.Scan(&col.ID, &col.Name, &col.ProjectID, &prevID, &nextID); err != nil {
			return nil, fmt.Errorf("scanning column row: %w", err)
		}

		// Set pointer fields
		col.PrevID = nullInt64ToPtr(prevID)
		col.NextID = nullInt64ToPtr(nextID)

		// Track head (where prev_id is NULL)
		if !prevID.Valid {
			headID = &col.ID
		}

		columnMap[col.ID] = col
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating column rows: %w", err)
	}

	// If no columns found, return empty slice
	if len(columnMap) == 0 {
		return []*models.Column{}, nil
	}

	// If no head found, database is in inconsistent state
	if headID == nil {
		return nil, fmt.Errorf("no head column found for project %d (linked list broken)", projectID)
	}

	// Traverse the linked list in memory using the map
	columns := make([]*models.Column, 0, len(columnMap))
	currentID := *headID

	for {
		col, exists := columnMap[currentID]
		if !exists {
			return nil, fmt.Errorf("column %d not found in map (linked list broken)", currentID)
		}

		columns = append(columns, col)

		// Move to next column
		if col.NextID == nil {
			// We've reached the tail
			break
		}
		currentID = *col.NextID
	}

	return columns, nil
}

// GetColumnByID retrieves a column by its ID
func (r *ColumnRepo) GetColumnByID(ctx context.Context, columnID int) (*models.Column, error) {
	column := &models.Column{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, project_id, prev_id, next_id FROM columns WHERE id = ?`,
		columnID,
	).Scan(&column.ID, &column.Name, &column.ProjectID, &column.PrevID, &column.NextID)
	if err != nil {
		return nil, err
	}
	return column, nil
}

// UpdateColumnName updates the name of an existing column
func (r *ColumnRepo) UpdateColumnName(ctx context.Context, columnID int, name string) error {
	// Get projectID for event notification
	projectID, err := getProjectIDFromTable(ctx, r.db, "columns", columnID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE columns SET name = ? WHERE id = ?`,
		name, columnID,
	)
	if err != nil {
		return err
	}

	// Send event notification
	sendEvent(r.eventClient, projectID)

	return nil
}

// DeleteColumn removes a column and all its tasks from the database
// This operation maintains the linked list structure by updating adjacent columns
func (r *ColumnRepo) DeleteColumn(ctx context.Context, columnID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// 1. Get the column's prev_id, next_id, and projectID
	var prevID, nextID sql.NullInt64
	var projectID int
	err = tx.QueryRowContext(ctx,
		`SELECT prev_id, next_id, project_id FROM columns WHERE id = ?`,
		columnID,
	).Scan(&prevID, &nextID, &projectID)
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

	if err := tx.Commit(); err != nil {
		return err
	}

	// Send event notification after commit
	sendEvent(r.eventClient, projectID)

	return nil
}
