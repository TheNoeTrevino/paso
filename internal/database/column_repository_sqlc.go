package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/models"
)

// columnRepository handles pure data access for columns
// No business logic, no events, no validation - just database operations
type columnRepository struct {
	queries *generated.Queries
	db      *sql.DB
}

// newColumnRepository creates a new column repository
func newColumnRepository(queries *generated.Queries, db *sql.DB) *columnRepository {
	return &columnRepository{
		queries: queries,
		db:      db,
	}
}

// CreateColumn inserts a new column
func (r *columnRepository) CreateColumn(ctx context.Context, name string, projectID int, prevID, nextID *int) (*models.Column, error) {
	var prevIDParam, nextIDParam interface{}
	if prevID != nil {
		prevIDParam = int64(*prevID)
	}
	if nextID != nil {
		nextIDParam = int64(*nextID)
	}

	row, err := r.queries.CreateColumn(ctx, generated.CreateColumnParams{
		Name:      name,
		ProjectID: int64(projectID),
		PrevID:    prevIDParam,
		NextID:    nextIDParam,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create column: %w", err)
	}
	return toColumnModel(row), nil
}

// GetColumnByID retrieves a column by ID
func (r *columnRepository) GetColumnByID(ctx context.Context, columnID int) (*models.Column, error) {
	row, err := r.queries.GetColumnByID(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get column %d: %w", columnID, err)
	}
	return &models.Column{
		ID:        int(row.ID),
		Name:      row.Name,
		ProjectID: int(row.ProjectID),
		PrevID:    interfaceToIntPtr(row.PrevID),
		NextID:    interfaceToIntPtr(row.NextID),
	}, nil
}

// GetColumnsByProject retrieves all columns for a project
func (r *columnRepository) GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error) {
	rows, err := r.queries.GetColumnsByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for project %d: %w", projectID, err)
	}

	columns := make([]*models.Column, len(rows))
	for i, row := range rows {
		columns[i] = &models.Column{
			ID:        int(row.ID),
			Name:      row.Name,
			ProjectID: int(row.ProjectID),
			PrevID:    interfaceToIntPtr(row.PrevID),
			NextID:    interfaceToIntPtr(row.NextID),
		}
	}
	return columns, nil
}

// GetTailColumnForProject finds the last column in a project (next_id IS NULL)
func (r *columnRepository) GetTailColumnForProject(ctx context.Context, projectID int) (*int, error) {
	id, err := r.queries.GetTailColumnForProject(ctx, int64(projectID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tail column for project %d: %w", projectID, err)
	}
	result := int(id)
	return &result, nil
}

// GetColumnNextID retrieves the next_id for a column
func (r *columnRepository) GetColumnNextID(ctx context.Context, columnID int) (*int, error) {
	result, err := r.queries.GetColumnNextID(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get next_id for column %d: %w", columnID, err)
	}
	return interfaceToIntPtr(result), nil
}

// GetColumnLinkedListInfo retrieves prev_id, next_id, and project_id for a column
func (r *columnRepository) GetColumnLinkedListInfo(ctx context.Context, columnID int) (prevID, nextID *int, projectID int, err error) {
	row, err := r.queries.GetColumnLinkedListInfo(ctx, int64(columnID))
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get linked list info for column %d: %w", columnID, err)
	}
	return interfaceToIntPtr(row.PrevID), interfaceToIntPtr(row.NextID), int(row.ProjectID), nil
}

// UpdateColumnName updates a column's name
func (r *columnRepository) UpdateColumnName(ctx context.Context, columnID int, name string) error {
	err := r.queries.UpdateColumnName(ctx, generated.UpdateColumnNameParams{
		Name: name,
		ID:   int64(columnID),
	})
	if err != nil {
		return fmt.Errorf("failed to update column %d name: %w", columnID, err)
	}
	return nil
}

// UpdateColumnNextID updates a column's next_id
func (r *columnRepository) UpdateColumnNextID(ctx context.Context, columnID int, nextID *int) error {
	var nextIDParam interface{}
	if nextID != nil {
		nextIDParam = int64(*nextID)
	}
	err := r.queries.UpdateColumnNextID(ctx, generated.UpdateColumnNextIDParams{
		NextID: nextIDParam,
		ID:     int64(columnID),
	})
	if err != nil {
		return fmt.Errorf("failed to update column %d next_id: %w", columnID, err)
	}
	return nil
}

// UpdateColumnPrevID updates a column's prev_id
func (r *columnRepository) UpdateColumnPrevID(ctx context.Context, columnID int, prevID *int) error {
	var prevIDParam interface{}
	if prevID != nil {
		prevIDParam = int64(*prevID)
	}
	err := r.queries.UpdateColumnPrevID(ctx, generated.UpdateColumnPrevIDParams{
		PrevID: prevIDParam,
		ID:     int64(columnID),
	})
	if err != nil {
		return fmt.Errorf("failed to update column %d prev_id: %w", columnID, err)
	}
	return nil
}

// DeleteColumn removes a column
func (r *columnRepository) DeleteColumn(ctx context.Context, columnID int) error {
	err := r.queries.DeleteColumn(ctx, int64(columnID))
	if err != nil {
		return fmt.Errorf("failed to delete column %d: %w", columnID, err)
	}
	return nil
}

// DeleteTasksByColumn removes all tasks in a column
func (r *columnRepository) DeleteTasksByColumn(ctx context.Context, columnID int) error {
	err := r.queries.DeleteTasksByColumn(ctx, int64(columnID))
	if err != nil {
		return fmt.Errorf("failed to delete tasks for column %d: %w", columnID, err)
	}
	return nil
}

// ColumnExists checks if a column exists
func (r *columnRepository) ColumnExists(ctx context.Context, columnID int) (bool, error) {
	count, err := r.queries.ColumnExists(ctx, int64(columnID))
	if err != nil {
		return false, fmt.Errorf("failed to check if column %d exists: %w", columnID, err)
	}
	return count > 0, nil
}

// WithTx returns a new repository instance that uses the given transaction
func (r *columnRepository) WithTx(tx *sql.Tx) *columnRepository {
	return &columnRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

// BeginTx starts a new transaction
func (r *columnRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// ============================================================================
// MODEL CONVERSION HELPERS
// ============================================================================

func toColumnModel(row generated.Column) *models.Column {
	return &models.Column{
		ID:        int(row.ID),
		Name:      row.Name,
		ProjectID: int(row.ProjectID),
		PrevID:    interfaceToIntPtr(row.PrevID),
		NextID:    interfaceToIntPtr(row.NextID),
	}
}

func interfaceToIntPtr(val interface{}) *int {
	if val == nil {
		return nil
	}
	if i64, ok := val.(int64); ok {
		result := int(i64)
		return &result
	}
	return nil
}
