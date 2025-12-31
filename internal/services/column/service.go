package column

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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
	return toColumnModelsFromRows(columns), nil
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
	return toColumnModelFromRow(column), nil
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

	// Start transaction for linked list updates
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	qtx := generated.New(tx)

	// If creating a ready column, clear any existing ready column first
	if req.HoldsReadyTasks {
		if err := qtx.ClearReadyColumnByProject(ctx, int64(req.ProjectID)); err != nil {
			return nil, fmt.Errorf("failed to clear existing ready column: %w", err)
		}
	}

	// If creating an in-progress column, clear any existing in-progress column first
	if req.HoldsInProgressTasks {
		if err := qtx.ClearInProgressColumnByProject(ctx, int64(req.ProjectID)); err != nil {
			return nil, fmt.Errorf("failed to clear existing in-progress column: %w", err)
		}
	}

	// If creating a completed column, check if one already exists
	if req.HoldsCompletedTasks {
		existingCompleted, err := qtx.GetCompletedColumnByProject(ctx, int64(req.ProjectID))
		if err == nil {
			// A completed column exists - return error
			return nil, fmt.Errorf("%w: column '%s' (ID: %d)", ErrCompletedColumnExists, existingCompleted.Name, existingCompleted.ID)
		}
		// sql.ErrNoRows is expected if no completed column exists
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check for existing completed column: %w", err)
		}
	}

	// Create new column
	column, err := qtx.CreateColumn(ctx, generated.CreateColumnParams{
		Name:                 req.Name,
		ProjectID:            int64(req.ProjectID),
		PrevID:               prevID,
		NextID:               nextID,
		HoldsReadyTasks:      req.HoldsReadyTasks,
		HoldsCompletedTasks:  req.HoldsCompletedTasks,
		HoldsInProgressTasks: req.HoldsInProgressTasks,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create column: %w", err)
	}

	// Update prev column's next_id to point to new column
	if prevID != nil {
		nextIDPtr := column.ID
		if err := qtx.UpdateColumnNextID(ctx, generated.UpdateColumnNextIDParams{
			NextID: &nextIDPtr,
			ID:     prevID.(int64),
		}); err != nil {
			return nil, fmt.Errorf("failed to update prev column: %w", err)
		}
	}

	// Update next column's prev_id to point to new column
	if nextID != nil {
		prevIDPtr := column.ID
		if err := qtx.UpdateColumnPrevID(ctx, generated.UpdateColumnPrevIDParams{
			PrevID: &prevIDPtr,
			ID:     nextID.(int64),
		}); err != nil {
			return nil, fmt.Errorf("failed to update next column: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event after successful commit
	s.publishColumnEvent(int(column.ID), int(column.ProjectID))

	return toColumnModel(column), nil
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
	s.publishColumnEvent(id, int(column.ProjectID))

	return nil
}

// SetHoldsReadyTasks sets a column as holding ready tasks
// Only one column per project can hold ready tasks
func (s *service) SetHoldsReadyTasks(ctx context.Context, columnID int) (*models.Column, error) {
	// Validate
	if columnID <= 0 {
		return nil, ErrInvalidColumnID
	}

	// Get column to verify it exists and get project ID
	column, err := s.queries.GetColumnByID(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get column: %w", err)
	}

	// Check if a ready column already exists
	existingReady, err := s.queries.GetReadyColumnByProject(ctx, column.ProjectID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing ready column: %w", err)
	}

	// If a ready column exists, return error
	if err == nil {
		return nil, fmt.Errorf("%w: column '%s' (ID: %d)", ErrReadyColumnExists, existingReady.Name, existingReady.ID)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	qtx := generated.New(tx)

	// Set this column as ready
	if err := qtx.UpdateColumnHoldsReadyTasks(ctx, generated.UpdateColumnHoldsReadyTasksParams{
		HoldsReadyTasks: true,
		ID:              int64(columnID),
	}); err != nil {
		return nil, fmt.Errorf("failed to set column as ready: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event
	s.publishColumnEvent(columnID, int(column.ProjectID))

	// Return updated column
	return s.GetColumnByID(ctx, columnID)
}

// SetHoldsCompletedTasks sets a column as holding completed tasks
// Only one column per project can hold completed tasks
// This method will return an error if a completed column already exists,
// unless the force flag is set to true
func (s *service) SetHoldsCompletedTasks(ctx context.Context, columnID int, force bool) (*models.Column, error) {
	// Validate
	if columnID <= 0 {
		return nil, ErrInvalidColumnID
	}

	// Get column to verify it exists and get project ID
	column, err := s.queries.GetColumnByID(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get column: %w", err)
	}

	// Check if a completed column already exists
	existingCompleted, err := s.queries.GetCompletedColumnByProject(ctx, column.ProjectID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing completed column: %w", err)
	}

	// If a completed column exists and force is not set, return error
	if err == nil && !force {
		return nil, fmt.Errorf("%w: column '%s' (ID: %d)", ErrCompletedColumnExists, existingCompleted.Name, existingCompleted.ID)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	qtx := generated.New(tx)

	// Clear any existing completed column
	if err := qtx.ClearCompletedColumnByProject(ctx, column.ProjectID); err != nil {
		return nil, fmt.Errorf("failed to clear existing completed column: %w", err)
	}

	// Set this column as completed
	if err := qtx.UpdateColumnHoldsCompletedTasks(ctx, generated.UpdateColumnHoldsCompletedTasksParams{
		HoldsCompletedTasks: true,
		ID:                  int64(columnID),
	}); err != nil {
		return nil, fmt.Errorf("failed to set column as completed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event
	s.publishColumnEvent(columnID, int(column.ProjectID))

	// Return updated column
	return s.GetColumnByID(ctx, columnID)
}

// SetHoldsInProgressTasks sets a column as holding in-progress tasks
// Only one column per project can hold in-progress tasks
func (s *service) SetHoldsInProgressTasks(ctx context.Context, columnID int) (*models.Column, error) {
	// Validate
	if columnID <= 0 {
		return nil, ErrInvalidColumnID
	}

	// Get column to verify it exists and get project ID
	column, err := s.queries.GetColumnByID(ctx, int64(columnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get column: %w", err)
	}

	// Check if an in-progress column already exists
	existingInProgress, err := s.queries.GetInProgressColumnByProject(ctx, column.ProjectID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing in-progress column: %w", err)
	}

	// If an in-progress column exists, return error
	if err == nil {
		return nil, fmt.Errorf("%w: column '%s' (ID: %d)", ErrInProgressColumnExists, existingInProgress.Name, existingInProgress.ID)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	qtx := generated.New(tx)

	// Set this column as in-progress
	if err := qtx.UpdateColumnHoldsInProgressTasks(ctx, generated.UpdateColumnHoldsInProgressTasksParams{
		HoldsInProgressTasks: true,
		ID:                   int64(columnID),
	}); err != nil {
		return nil, fmt.Errorf("failed to set column as in-progress: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event
	s.publishColumnEvent(columnID, int(column.ProjectID))

	// Return updated column
	return s.GetColumnByID(ctx, columnID)
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

	prevID := database.InterfaceToIntPtr(linkedListInfo.PrevID)
	nextID := database.InterfaceToIntPtr(linkedListInfo.NextID)
	projectID := int(linkedListInfo.ProjectID)

	// Start transaction for linked list updates
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event after successful deletion
	s.publishColumnEvent(id, projectID)

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

// publishColumnEvent publishes a column event
func (s *service) publishColumnEvent(columnID, projectID int) {
	if s.eventClient == nil {
		return
	}

	if err := s.eventClient.SendEvent(events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: projectID,
	}); err != nil {
		log.Printf("failed to send event for column %d: %v", columnID, err)
	}
}

// Model conversion helpers

func toColumnModel(c generated.Column) *models.Column {
	return &models.Column{
		ID:                   int(c.ID),
		Name:                 c.Name,
		ProjectID:            int(c.ProjectID),
		PrevID:               database.InterfaceToIntPtr(c.PrevID),
		NextID:               database.InterfaceToIntPtr(c.NextID),
		HoldsReadyTasks:      c.HoldsReadyTasks,
		HoldsCompletedTasks:  c.HoldsCompletedTasks,
		HoldsInProgressTasks: c.HoldsInProgressTasks,
	}
}

func toColumnModelFromRow(r generated.GetColumnByIDRow) *models.Column {
	return &models.Column{
		ID:                   int(r.ID),
		Name:                 r.Name,
		ProjectID:            int(r.ProjectID),
		PrevID:               database.InterfaceToIntPtr(r.PrevID),
		NextID:               database.InterfaceToIntPtr(r.NextID),
		HoldsReadyTasks:      r.HoldsReadyTasks,
		HoldsCompletedTasks:  r.HoldsCompletedTasks,
		HoldsInProgressTasks: r.HoldsInProgressTasks,
	}
}

func toColumnModelsFromRows(rows []generated.GetColumnsByProjectRow) []*models.Column {
	result := make([]*models.Column, len(rows))
	for i, r := range rows {
		result[i] = &models.Column{
			ID:                   int(r.ID),
			Name:                 r.Name,
			ProjectID:            int(r.ProjectID),
			PrevID:               database.InterfaceToIntPtr(r.PrevID),
			NextID:               database.InterfaceToIntPtr(r.NextID),
			HoldsReadyTasks:      r.HoldsReadyTasks,
			HoldsCompletedTasks:  r.HoldsCompletedTasks,
			HoldsInProgressTasks: r.HoldsInProgressTasks,
		}
	}
	return result
}
