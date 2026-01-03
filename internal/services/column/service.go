package column

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thenoetrevino/paso/internal/converters"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// Service defines all column-related business operations
type Service interface {
	// Read operations
	GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error)
	GetColumnByID(ctx context.Context, id int) (*models.Column, error)

	// Write operations
	CreateColumn(ctx context.Context, req CreateColumnRequest) (*models.Column, error)
	UpdateColumnName(ctx context.Context, id int, name string) error
	SetHoldsReadyTasks(ctx context.Context, columnID int) (*models.Column, error)
	SetHoldsCompletedTasks(ctx context.Context, columnID int, force bool) (*models.Column, error)
	SetHoldsInProgressTasks(ctx context.Context, columnID int) (*models.Column, error)
	DeleteColumn(ctx context.Context, id int) error
}

// CreateColumnRequest encapsulates data for creating a column
type CreateColumnRequest struct {
	Name                 string
	ProjectID            int
	AfterID              *int // Optional: ID of column to insert after (nil = append to end)
	HoldsReadyTasks      bool // Whether this column holds ready tasks
	HoldsCompletedTasks  bool // Whether this column holds completed tasks
	HoldsInProgressTasks bool // Whether this column holds in-progress tasks
}

// service implements Service interface using SQLC directly
type service struct {
	db          *sql.DB
	queries     generated.Querier
	eventClient events.EventPublisher
}

// NewService creates a new column service
func NewService(db *sql.DB, eventClient events.EventPublisher) Service {
	return &service{
		db:          db,
		queries:     generated.New(db),
		eventClient: eventClient,
	}
}

// GetColumnsByProject retrieves all columns for a project
func (s *service) GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	columns, err := s.queries.GetColumnsByProject(ctx, int64(projectID))
	if err != nil {
		return nil, err
	}
	return converters.ColumnsFromRowsToModels(columns), nil
}

// GetColumnByID retrieves a specific column
func (s *service) GetColumnByID(ctx context.Context, id int) (*models.Column, error) {
	if id <= 0 {
		return nil, ErrInvalidColumnID
	}
	column, err := s.queries.GetColumnByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return converters.ColumnFromIDRowToModel(column), nil
}

// CreateColumn creates a new column with validation and linked list management
func (s *service) CreateColumn(ctx context.Context, req CreateColumnRequest) (*models.Column, error) {
	// Validate request
	if err := s.validateCreateColumn(req); err != nil {
		return nil, err
	}

	var prevID, nextID interface{}

	if req.AfterID == nil {
		// Append to end: find tail column
		tailIDVal, err := s.queries.GetTailColumnForProject(ctx, int64(req.ProjectID))
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get tail column: %w", err)
		}
		if tailIDVal != 0 {
			prevID = tailIDVal
		}
		nextID = nil
	} else {
		// Insert after specified column
		prevID = int64(*req.AfterID)
		// Get the next_id of the "after" column
		afterNextID, err := s.queries.GetColumnNextID(ctx, int64(*req.AfterID))
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get next column: %w", err)
		}
		nextID = afterNextID
	}

	var column generated.Column

	// Use WithTx helper for linked list updates
	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := generated.New(tx)

		// If creating a ready column, clear any existing ready column first
		if req.HoldsReadyTasks {
			if err := qtx.ClearReadyColumnByProject(ctx, int64(req.ProjectID)); err != nil {
				return fmt.Errorf("failed to clear existing ready column: %w", err)
			}
		}

		// If creating an in-progress column, clear any existing in-progress column first
		if req.HoldsInProgressTasks {
			if err := qtx.ClearInProgressColumnByProject(ctx, int64(req.ProjectID)); err != nil {
				return fmt.Errorf("failed to clear existing in-progress column: %w", err)
			}
		}

		// If creating a completed column, check if one already exists
		if req.HoldsCompletedTasks {
			existingCompleted, err := qtx.GetCompletedColumnByProject(ctx, int64(req.ProjectID))
			if err == nil {
				// A completed column exists - return error
				return fmt.Errorf("%w: column '%s' (ID: %d)", ErrCompletedColumnExists, existingCompleted.Name, existingCompleted.ID)
			}
			// sql.ErrNoRows is expected if no completed column exists
			if err != sql.ErrNoRows {
				return fmt.Errorf("failed to check for existing completed column: %w", err)
			}
		}

		// Create new column
		var colErr error
		column, colErr = qtx.CreateColumn(ctx, generated.CreateColumnParams{
			Name:                 req.Name,
			ProjectID:            int64(req.ProjectID),
			PrevID:               prevID,
			NextID:               nextID,
			HoldsReadyTasks:      req.HoldsReadyTasks,
			HoldsCompletedTasks:  req.HoldsCompletedTasks,
			HoldsInProgressTasks: req.HoldsInProgressTasks,
		})
		if colErr != nil {
			return fmt.Errorf("failed to create column: %w", colErr)
		}

		// Update prev column's next_id to point to new column
		if prevID != nil {
			nextIDPtr := column.ID
			if err := qtx.UpdateColumnNextID(ctx, generated.UpdateColumnNextIDParams{
				NextID: &nextIDPtr,
				ID:     prevID.(int64),
			}); err != nil {
				return fmt.Errorf("failed to update prev column: %w", err)
			}
		}

		// Update next column's prev_id to point to new column
		if nextID != nil {
			prevIDPtr := column.ID
			if err := qtx.UpdateColumnPrevID(ctx, generated.UpdateColumnPrevIDParams{
				PrevID: &prevIDPtr,
				ID:     nextID.(int64),
			}); err != nil {
				return fmt.Errorf("failed to update next column: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event after successful commit
	s.publishColumnEvent(ctx, int(column.ID), int(column.ProjectID))

	return converters.ColumnToModel(column), nil
}

// UpdateColumnName updates a column's name
func (s *service) UpdateColumnName(ctx context.Context, id int, name string) error {
	// Validate
	if id <= 0 {
		return ErrInvalidColumnID
	}
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 50 {
		return ErrNameTooLong
	}

	// Get column to find project ID for event
	column, err := s.queries.GetColumnByID(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("failed to get column: %w", err)
	}

	// Update column
	if err := s.queries.UpdateColumnName(ctx, generated.UpdateColumnNameParams{
		Name: name,
		ID:   int64(id),
	}); err != nil {
		return fmt.Errorf("failed to update column: %w", err)
	}

	// Publish event
	s.publishColumnEvent(ctx, id, int(column.ProjectID))

	return nil
}

// specialColumnStateType defines the type of special column state to set
type specialColumnStateType int

const (
	// stateReady indicates a column that holds ready tasks
	stateReady specialColumnStateType = iota
	// stateCompleted indicates a column that holds completed tasks
	stateCompleted
	// stateInProgress indicates a column that holds in-progress tasks
	stateInProgress
)

// setSpecialColumnState is a parametrized helper function that sets a column's special state.
// It handles the common pattern of:
// 1. Validating the column ID
// 2. Getting the column to verify it exists and get the project ID
// 3. Checking if another column with this state already exists
// 4. Starting a transaction
// 5. Optionally clearing the flag from other columns
// 6. Setting the flag on the target column
// 7. Committing the transaction
// 8. Publishing the column event
// 9. Returning the updated column
//
// Parameters:
//   - ctx: context
//   - columnID: the ID of the column to set as special
//   - stateType: the type of special state to set (ready, completed, or in-progress)
//   - allowForce: if true and a column with this state exists, it will be replaced;
//     if false, an error is returned instead. Only applies to completed state.
func (s *service) setSpecialColumnState(ctx context.Context, columnID int, stateType specialColumnStateType, allowForce bool) (*models.Column, error) {
	// Validate
	if columnID <= 0 {
		return nil, ErrInvalidColumnID
	}

	// Get column to verify it exists and get project ID
	column, err := s.queries.GetColumnByID(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get column: %w", err)
	}

	// Check if a column with this state already exists and enforce state-specific rules
	if err := s.checkExistingState(ctx, stateType, column.ProjectID, allowForce); err != nil {
		return nil, err
	}

	// Use WithTx helper for transaction management
	err = database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := generated.New(tx)

		// Clear any existing column with this state (if applicable)
		if err := s.clearExistingState(ctx, qtx, stateType, column.ProjectID); err != nil {
			return err
		}

		// Set this column's special state
		if err := s.setColumnState(ctx, qtx, columnID, stateType); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publish event
	s.publishColumnEvent(ctx, columnID, int(column.ProjectID))

	// Return updated column
	return s.GetColumnByID(ctx, columnID)
}

// checkExistingState verifies that no conflicting column state exists.
// For ready and in-progress states: returns error if any column with that state exists.
// For completed state: returns error only if force is false and a column with that state exists.
func (s *service) checkExistingState(ctx context.Context, stateType specialColumnStateType, projectID int64, allowForce bool) error {
	switch stateType {
	case stateReady:
		existingReady, err := s.queries.GetReadyColumnByProject(ctx, projectID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check for existing ready column: %w", err)
		}
		if err == nil {
			return fmt.Errorf("%w: column '%s' (ID: %d)", ErrReadyColumnExists, existingReady.Name, existingReady.ID)
		}

	case stateCompleted:
		existingCompleted, err := s.queries.GetCompletedColumnByProject(ctx, projectID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check for existing completed column: %w", err)
		}
		if err == nil && !allowForce {
			return fmt.Errorf("%w: column '%s' (ID: %d)", ErrCompletedColumnExists, existingCompleted.Name, existingCompleted.ID)
		}

	case stateInProgress:
		existingInProgress, err := s.queries.GetInProgressColumnByProject(ctx, projectID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check for existing in-progress column: %w", err)
		}
		if err == nil {
			return fmt.Errorf("%w: column '%s' (ID: %d)", ErrInProgressColumnExists, existingInProgress.Name, existingInProgress.ID)
		}
	}

	return nil
}

// clearExistingState clears the special state flag from other columns (only for completed state).
// For ready and in-progress states, the unique constraint prevents duplicates naturally.
// For completed state, we allow force-setting a new column as completed, clearing the old one.
func (s *service) clearExistingState(ctx context.Context, qtx generated.Querier, stateType specialColumnStateType, projectID int64) error {
	switch stateType {
	case stateCompleted:
		if err := qtx.ClearCompletedColumnByProject(ctx, projectID); err != nil {
			return fmt.Errorf("failed to clear existing completed column: %w", err)
		}
	case stateReady, stateInProgress:
		// For ready and in-progress, the unique constraint prevents duplicates,
		// so we don't need to explicitly clear them
	}

	return nil
}

// setColumnState updates a column to have the given special state.
func (s *service) setColumnState(ctx context.Context, qtx generated.Querier, columnID int, stateType specialColumnStateType) error {
	switch stateType {
	case stateReady:
		if err := qtx.UpdateColumnHoldsReadyTasks(ctx, generated.UpdateColumnHoldsReadyTasksParams{
			HoldsReadyTasks: true,
			ID:              int64(columnID),
		}); err != nil {
			return fmt.Errorf("failed to set column as ready: %w", err)
		}

	case stateCompleted:
		if err := qtx.UpdateColumnHoldsCompletedTasks(ctx, generated.UpdateColumnHoldsCompletedTasksParams{
			HoldsCompletedTasks: true,
			ID:                  int64(columnID),
		}); err != nil {
			return fmt.Errorf("failed to set column as completed: %w", err)
		}

	case stateInProgress:
		if err := qtx.UpdateColumnHoldsInProgressTasks(ctx, generated.UpdateColumnHoldsInProgressTasksParams{
			HoldsInProgressTasks: true,
			ID:                   int64(columnID),
		}); err != nil {
			return fmt.Errorf("failed to set column as in-progress: %w", err)
		}
	}

	return nil
}

// SetHoldsReadyTasks sets a column as holding ready tasks.
// Only one column per project can hold ready tasks.
func (s *service) SetHoldsReadyTasks(ctx context.Context, columnID int) (*models.Column, error) {
	return s.setSpecialColumnState(ctx, columnID, stateReady, false)
}

// SetHoldsCompletedTasks sets a column as holding completed tasks.
// Only one column per project can hold completed tasks.
// This method will return an error if a completed column already exists,
// unless the force flag is set to true.
func (s *service) SetHoldsCompletedTasks(ctx context.Context, columnID int, force bool) (*models.Column, error) {
	return s.setSpecialColumnState(ctx, columnID, stateCompleted, force)
}

// SetHoldsInProgressTasks sets a column as holding in-progress tasks.
// Only one column per project can hold in-progress tasks.
func (s *service) SetHoldsInProgressTasks(ctx context.Context, columnID int) (*models.Column, error) {
	return s.setSpecialColumnState(ctx, columnID, stateInProgress, false)
}

// DeleteColumn deletes a column (business rule: must not have tasks)
func (s *service) DeleteColumn(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidColumnID
	}

	// Business rule: Check if column has tasks
	tasks, err := s.queries.GetTasksByColumn(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("failed to check column tasks: %w", err)
	}
	if len(tasks) > 0 {
		return ErrColumnHasTasks
	}

	// Get column info for linked list updates and project ID
	linkedListInfo, err := s.queries.GetColumnLinkedListInfo(ctx, int64(id))
	if err != nil {
		return fmt.Errorf("failed to get column info: %w", err)
	}

	prevID := database.AnyToIntPtr(linkedListInfo.PrevID)
	nextID := database.AnyToIntPtr(linkedListInfo.NextID)
	projectID := int(linkedListInfo.ProjectID)

	// Use WithTx helper for linked list updates
	err = database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := generated.New(tx)

		// Update prev column's next_id to skip deleted column
		if prevID != nil {
			var nextIDInterface interface{}
			if nextID != nil {
				nextIDInterface = int64(*nextID)
			}
			if err := qtx.UpdateColumnNextID(ctx, generated.UpdateColumnNextIDParams{
				NextID: nextIDInterface,
				ID:     int64(*prevID),
			}); err != nil {
				return fmt.Errorf("failed to update prev column: %w", err)
			}
		}

		// Update next column's prev_id to skip deleted column
		if nextID != nil {
			var prevIDInterface interface{}
			if prevID != nil {
				prevIDInterface = int64(*prevID)
			}
			if err := qtx.UpdateColumnPrevID(ctx, generated.UpdateColumnPrevIDParams{
				PrevID: prevIDInterface,
				ID:     int64(*nextID),
			}); err != nil {
				return fmt.Errorf("failed to update next column: %w", err)
			}
		}

		// Delete column
		if err := qtx.DeleteColumn(ctx, int64(id)); err != nil {
			return fmt.Errorf("failed to delete column: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Publish event after successful deletion
	s.publishColumnEvent(ctx, id, projectID)

	return nil
}

// validateCreateColumn validates a CreateColumnRequest
func (s *service) validateCreateColumn(req CreateColumnRequest) error {
	if req.Name == "" {
		return ErrEmptyName
	}
	if len(req.Name) > 50 {
		return ErrNameTooLong
	}
	if req.ProjectID <= 0 {
		return ErrInvalidProjectID
	}
	if req.AfterID != nil && *req.AfterID <= 0 {
		return ErrInvalidColumnID
	}
	return nil
}

// publishColumnEvent publishes a column event with retry logic
func (s *service) publishColumnEvent(ctx context.Context, columnID, projectID int) {
	if s.eventClient == nil {
		return
	}

	// Publish with retry (3 attempts with exponential backoff)
	// Non-blocking: errors are logged but don't affect the operation
	_ = events.PublishWithRetry(s.eventClient, events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: projectID,
	}, 3)
}
