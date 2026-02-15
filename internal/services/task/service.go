package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/thenoetrevino/paso/internal/converters"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/types"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/taskevent"
	"github.com/thenoetrevino/paso/internal/user"
)

var ErrTaskAlreadyInTargetProject = errors.New("task is already in the target project")

// TaskFilterParams holds optional filter criteria for the unified filter query.
// Zero values / nil pointers mean "no filter" for that field.
type TaskFilterParams struct {
	ProjectID  int
	Title      *string // nil = no title filter; non-nil = LIKE %value%
	PriorityID *int    // nil = no filter; value = exact match
	TypeID     *int    // nil = no filter; value = exact match
	AssigneeID *int    // nil = no filter; -1 = unassigned; value = exact match
	LabelIDs   []int   // empty = no filter; non-empty = OR match (task has ANY of these)
}

// TaskReader defines task reading operations
type TaskReader interface {
	GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error)
	GetTaskActivities(ctx context.Context, taskID int) ([]models.ActivityItem, error)
	GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error)
	GetTaskSummariesWithFilters(ctx context.Context, params TaskFilterParams) (map[int][]*models.TaskSummary, error)
	GetReadyTaskSummariesByProject(ctx context.Context, projectID int) ([]*models.TaskSummary, error)
	GetInProgressTasksByProject(ctx context.Context, projectID int) ([]*models.TaskDetail, error)
	GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error)
	GetTaskTreeByProject(ctx context.Context, projectID int) ([]*models.TaskTreeNode, error)
	GetTaskTypeAndPriorityIDs(ctx context.Context, taskID int) (typeID, priorityID int, err error)
}

// TaskWriter defines task writing operations
type TaskWriter interface {
	CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error)
	UpdateTask(ctx context.Context, req UpdateTaskRequest) error
	UpdateTaskAssignee(ctx context.Context, taskID int, assigneeID *int) error
	UpdateTaskEstimate(ctx context.Context, taskID int, estimate *string) error
	UpdateTaskDueDate(ctx context.Context, taskID int, dueDate *time.Time) error
	DeleteTask(ctx context.Context, taskID int) error
}

// TaskMover defines task movement operations within the task management system
type TaskMover interface {
	// Column-based movement (workflow progression)
	MoveTaskToNextColumn(ctx context.Context, taskID int) error
	MoveTaskToPrevColumn(ctx context.Context, taskID int) error
	MoveTaskToColumn(ctx context.Context, taskID, columnID int) error
	MoveTaskToReadyColumn(ctx context.Context, taskID int) error
	MoveTaskToCompletedColumn(ctx context.Context, taskID int) error
	MoveTaskToInProgressColumn(ctx context.Context, taskID int) error

	// Cross-project movement
	MoveTaskToProject(ctx context.Context, taskID int, targetProjectID int) error

	// Position-based movement (ordering within column)
	MoveTaskUp(ctx context.Context, taskID int) error
	MoveTaskDown(ctx context.Context, taskID int) error
}

// TaskRelationer defines task relationship operations (parent/child/blocking relationships)
type TaskRelationer interface {
	AddParentRelation(ctx context.Context, taskID, parentID int, relationTypeID int) error
	AddChildRelation(ctx context.Context, taskID, childID int, relationTypeID int) error
	RemoveParentRelation(ctx context.Context, taskID, parentID int) error
	RemoveChildRelation(ctx context.Context, taskID, childID int) error
}

// TaskLabeler defines label management operations for tasks
type TaskLabeler interface {
	AttachLabel(ctx context.Context, taskID, labelID int) error
	DetachLabel(ctx context.Context, taskID, labelID int) error
}

// TaskCommenter defines comment operations on tasks
type TaskCommenter interface {
	CreateComment(ctx context.Context, req CreateCommentRequest) (*models.Comment, error)
	UpdateComment(ctx context.Context, req UpdateCommentRequest) error
	DeleteComment(ctx context.Context, commentID int) error
	GetCommentsByTask(ctx context.Context, taskID int) ([]*models.Comment, error)
}

// Service defines all task-related business operations as a composition of focused interfaces.
// This composite interface provides better separation of concerns through interface segregation.
type Service interface {
	TaskReader
	TaskMover
	TaskWriter
	TaskRelationer
	TaskLabeler
	TaskCommenter
}

// CreateTaskRequest encapsulates all data needed to create a task
type CreateTaskRequest struct {
	Title        string
	Description  string
	ColumnID     int
	Position     int
	PriorityID   int        // Optional: 0 means use default
	TypeID       int        // Optional: 0 means use default
	AssigneeID   int        // Optional: 0 means use active assignee from config
	Estimate     string     // Optional: empty means no estimate
	DueDate      *time.Time // Optional: nil means no due date
	LabelIDs     []int
	ParentIDs    []int // Parent task IDs (tasks that depend on this task)
	ChildIDs     []int // Child task IDs (tasks this task depends on)
	BlockedByIDs []int // Tasks that block this task
	BlocksIDs    []int // Tasks that are blocked by this task
}

// UpdateTaskRequest encapsulates all data needed to update a task
// Fields with pointers are optional - nil means don't update
type UpdateTaskRequest struct {
	TaskID      int
	Title       *string
	Description *string
	PriorityID  *int
	TypeID      *int
}

// CreateCommentRequest encapsulates data for creating a comment
type CreateCommentRequest struct {
	TaskID  int
	Message string
	Author  string
}

// UpdateCommentRequest encapsulates data for updating a comment
type UpdateCommentRequest struct {
	CommentID int
	Message   string
}

// service implements Service interface using database.Querier abstraction
type service struct {
	db           *sql.DB
	dbType       database.DatabaseType
	queries      database.Querier
	eventClient  events.EventPublisher
	eventService taskevent.Service
}

// NewService creates a new task service with database-agnostic queries.
func NewService(db *sql.DB, dbType database.DatabaseType, eventClient events.EventPublisher, eventService taskevent.Service) (Service, error) {
	queries, err := database.NewQuerier(db, dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create task service: %w", err)
	}
	return &service{
		db:           db,
		dbType:       dbType,
		queries:      queries,
		eventClient:  eventClient,
		eventService: eventService,
	}, nil
}

// CreateTask handles task creation with validation and business rules
func (s *service) CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateCreateTaskRequest(req); err != nil {
		return nil, err
	}

	// Get project ID from column
	projectID, err := s.queries.GetProjectIDFromColumn(ctx, int64(req.ColumnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get project ID: %w", err)
	}

	var createdTask types.Task

	// Use WithTx helper for transaction management
	err = database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := database.MustNewQuerier(tx, s.dbType)

		// Get next ticket number
		ticketNumber, err := qtx.GetNextTicketNumber(ctx, projectID)
		if err != nil {
			return fmt.Errorf("failed to get ticket number: %w", err)
		}

		// Create task
		var desc types.NullString
		if req.Description != "" {
			desc = types.NullString{String: req.Description, Valid: true}
		}

		var assigneeID types.NullInt64
		if req.AssigneeID > 0 {
			assigneeID = types.NullInt64{Int64: int64(req.AssigneeID), Valid: true}
		}

		var estimate types.NullString
		if req.Estimate != "" {
			estimate = types.NullString{String: req.Estimate, Valid: true}
		}

		var dueDate types.NullTime
		if req.DueDate != nil {
			dueDate = types.NullTime{Time: *req.DueDate, Valid: true}
		}

		var taskErr error
		createdTask, taskErr = qtx.CreateTask(ctx, types.CreateTaskParams{
			Title:        req.Title,
			Description:  desc,
			ColumnID:     int64(req.ColumnID),
			Position:     int64(req.Position),
			TicketNumber: ticketNumber,
			AssigneeID:   assigneeID,
			Estimate:     estimate,
			DueDate:      dueDate,
		})
		if taskErr != nil {
			return fmt.Errorf("failed to create task: %w", taskErr)
		}

		// Increment ticket number
		if err := qtx.IncrementTicketNumber(ctx, projectID); err != nil {
			return fmt.Errorf("failed to increment ticket number: %w", err)
		}

		// Set priority if provided (default is handled by database)
		if req.PriorityID > 0 {
			if err := qtx.UpdateTaskPriority(ctx, types.UpdateTaskPriorityParams{
				PriorityID: int64(req.PriorityID),
				ID:         createdTask.ID,
			}); err != nil {
				return fmt.Errorf("failed to set priority: %w", err)
			}
		}

		// Set type if provided (default is handled by database)
		if req.TypeID > 0 {
			if err := qtx.UpdateTaskType(ctx, types.UpdateTaskTypeParams{
				TypeID: int64(req.TypeID),
				ID:     createdTask.ID,
			}); err != nil {
				return fmt.Errorf("failed to set type: %w", err)
			}
		}

		// Attach labels
		for _, labelID := range req.LabelIDs {
			if err := qtx.AddLabelToTask(ctx, types.AddLabelToTaskParams{
				TaskID:  createdTask.ID,
				LabelID: int64(labelID),
			}); err != nil {
				return fmt.Errorf("failed to attach label %d: %w", labelID, err)
			}
		}

		// Add parent relationships (tasks that depend on this task)
		for _, parentID := range req.ParentIDs {
			if err := qtx.AddSubtaskWithRelationType(ctx, types.AddSubtaskWithRelationTypeParams{
				ParentID:       int64(parentID),
				ChildID:        createdTask.ID,
				RelationTypeID: models.RelationTypeParentChild,
			}); err != nil {
				return fmt.Errorf("failed to add parent relation: %w", err)
			}
		}

		// Add child relationships (tasks this task depends on)
		for _, childID := range req.ChildIDs {
			if err := qtx.AddSubtaskWithRelationType(ctx, types.AddSubtaskWithRelationTypeParams{
				ParentID:       createdTask.ID,
				ChildID:        int64(childID),
				RelationTypeID: models.RelationTypeParentChild,
			}); err != nil {
				return fmt.Errorf("failed to add child relation: %w", err)
			}
		}

		// Add blocking relationships (tasks that block this task)
		for _, blockerID := range req.BlockedByIDs {
			// This task (Parent) is blocked by blockerID (Child)
			if err := qtx.AddSubtaskWithRelationType(ctx, types.AddSubtaskWithRelationTypeParams{
				ParentID:       createdTask.ID,
				ChildID:        int64(blockerID),
				RelationTypeID: models.RelationTypeBlocking,
			}); err != nil {
				return fmt.Errorf("failed to add blocked-by relation: %w", err)
			}
		}

		// Add blocked relationships (tasks that are blocked by this task)
		for _, blockedID := range req.BlocksIDs {
			// blockedID (Parent) is blocked by this task (Child)
			if err := qtx.AddSubtaskWithRelationType(ctx, types.AddSubtaskWithRelationTypeParams{
				ParentID:       int64(blockedID),
				ChildID:        createdTask.ID,
				RelationTypeID: models.RelationTypeBlocking,
			}); err != nil {
				return fmt.Errorf("failed to add blocks relation: %w", err)
			}
		}

		// Emit TaskCreated event within the transaction
		if s.eventService != nil {
			if err := s.eventService.CreateTaskCreatedEvent(ctx, qtx, int(createdTask.ID), req.Title, user.GetCurrentUsername()); err != nil {
				return fmt.Errorf("failed to create task event: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Publish event after successful commit
	s.publishTaskEvent(ctx, int(createdTask.ID))

	// Convert to model
	return converters.TaskToModel(createdTask), nil
}

// UpdateTask handles task updates with validation
func (s *service) UpdateTask(ctx context.Context, req UpdateTaskRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateUpdateTaskRequest(req); err != nil {
		return err
	}

	// Before the update, get current values for event emission
	var oldPriorityName, oldTypeName string
	var currentPriorityID, currentTypeID int64
	if s.eventService != nil && (req.PriorityID != nil || req.TypeID != nil) {
		ids, err := s.queries.GetTaskTypeAndPriorityIDs(ctx, int64(req.TaskID))
		if err == nil {
			currentPriorityID = ids.PriorityID
			currentTypeID = ids.TypeID

			if req.PriorityID != nil {
				priorities, _ := s.queries.GetAllPriorities(ctx)
				for _, p := range priorities {
					if p.ID == currentPriorityID {
						oldPriorityName = p.Description
						break
					}
				}
			}
			if req.TypeID != nil {
				taskTypes, _ := s.queries.GetAllTypes(ctx)
				for _, t := range taskTypes {
					if t.ID == currentTypeID {
						oldTypeName = t.Description
						break
					}
				}
			}
		}
	}

	// Use WithTx helper for transaction management to ensure atomicity
	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := database.MustNewQuerier(tx, s.dbType)

		// Update basic fields if provided
		if req.Title != nil || req.Description != nil {
			var title string
			var description types.NullString

			// Only query if we need to preserve existing values for fields not being updated
			if req.Title == nil || req.Description == nil {
				detail, err := qtx.GetTaskDetail(ctx, int64(req.TaskID))
				if err != nil {
					return fmt.Errorf("failed to get task: %w", err)
				}
				if req.Title == nil {
					title = detail.Title
				}
				if req.Description == nil {
					description = detail.Description
				}
			}

			// Use provided values or fall back to existing values
			if req.Title != nil {
				title = *req.Title
			}
			if req.Description != nil {
				description = types.NullString{String: *req.Description, Valid: true}
			}

			if err := qtx.UpdateTask(ctx, types.UpdateTaskParams{
				Title:       title,
				Description: description,
				ID:          int64(req.TaskID),
			}); err != nil {
				return fmt.Errorf("failed to update task: %w", err)
			}
		}

		// Update priority if provided
		if req.PriorityID != nil {
			if err := qtx.UpdateTaskPriority(ctx, types.UpdateTaskPriorityParams{
				PriorityID: int64(*req.PriorityID),
				ID:         int64(req.TaskID),
			}); err != nil {
				return fmt.Errorf("failed to update priority: %w", err)
			}
		}

		// Update type if provided
		if req.TypeID != nil {
			if err := qtx.UpdateTaskType(ctx, types.UpdateTaskTypeParams{
				TypeID: int64(*req.TypeID),
				ID:     int64(req.TaskID),
			}); err != nil {
				return fmt.Errorf("failed to update type: %w", err)
			}
		}

		// Emit events within transaction for consistency
		if s.eventService != nil {
			if req.PriorityID != nil && int64(*req.PriorityID) != currentPriorityID {
				priorities, _ := qtx.GetAllPriorities(ctx)
				for _, p := range priorities {
					if p.ID == int64(*req.PriorityID) {
						if err := s.eventService.CreatePriorityChangedEvent(ctx, qtx, req.TaskID, oldPriorityName, p.Description, user.GetCurrentUsername()); err != nil {
							slog.Warn("failed to create priority changed event", "error", err, "taskID", req.TaskID, "eventType", "PriorityChanged")
						}
						break
					}
				}
			}
			if req.TypeID != nil && int64(*req.TypeID) != currentTypeID {
				taskTypes, _ := qtx.GetAllTypes(ctx)
				for _, t := range taskTypes {
					if t.ID == int64(*req.TypeID) {
						if err := s.eventService.CreateTypeChangedEvent(ctx, qtx, req.TaskID, oldTypeName, t.Description, user.GetCurrentUsername()); err != nil {
							slog.Warn("failed to create type changed event", "error", err, "taskID", req.TaskID, "eventType", "TypeChanged")
						}
						break
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Publish event after successful commit
	s.publishTaskEvent(ctx, req.TaskID)

	return nil
}

// UpdateTaskAssignee updates the assignee for a task
func (s *service) UpdateTaskAssignee(ctx context.Context, taskID int, assigneeID *int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	var nullID types.NullInt64
	if assigneeID != nil {
		nullID = types.NullInt64{Int64: int64(*assigneeID), Valid: true}
	}

	if err := s.queries.UpdateTaskAssignee(ctx, types.UpdateTaskAssigneeParams{
		AssigneeID: nullID,
		ID:         int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to update task assignee: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// UpdateTaskEstimate updates the estimate for a task
func (s *service) UpdateTaskEstimate(ctx context.Context, taskID int, estimate *string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	if err := ValidateEstimate(estimate); err != nil {
		return err
	}

	var nullEstimate types.NullString
	if estimate != nil {
		nullEstimate = types.NullString{String: *estimate, Valid: true}
	}

	if err := s.queries.UpdateTaskEstimate(ctx, types.UpdateTaskEstimateParams{
		Estimate: nullEstimate,
		ID:       int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to update task estimate: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// UpdateTaskDueDate updates the due date field for an existing task.
// Pass nil to clear the due date.
func (s *service) UpdateTaskDueDate(ctx context.Context, taskID int, dueDate *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	var nullDueDate types.NullTime
	if dueDate != nil {
		nullDueDate = types.NullTime{Time: *dueDate, Valid: true}
	}

	if err := s.queries.UpdateTaskDueDate(ctx, types.UpdateTaskDueDateParams{
		DueDate: nullDueDate,
		ID:      int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to update task due date: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// DeleteTask handles task deletion
func (s *service) DeleteTask(ctx context.Context, taskID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	// Get project ID BEFORE deletion to avoid race condition in event publishing
	projectID, err := s.queries.GetProjectIDFromTask(ctx, int64(taskID))
	if err != nil {
		return fmt.Errorf("failed to get project ID for task: %w", err)
	}

	if err := s.queries.DeleteTask(ctx, int64(taskID)); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Publish event using the pre-fetched project ID
	s.publishTaskEventForProject(int(projectID))

	return nil
}

// GetTaskDetail retrieves full task details
func (s *service) GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return nil, ErrInvalidTaskID
	}

	// Get task detail
	taskRow, err := s.queries.GetTaskDetail(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task detail: %w", err)
	}

	// Get labels
	labels, err := s.queries.GetTaskLabels(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task labels: %w", err)
	}

	// Get parent tasks
	parentRows, err := s.queries.GetParentTasks(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get parent tasks: %w", err)
	}

	// Get child tasks
	childRows, err := s.queries.GetChildTasks(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get child tasks: %w", err)
	}

	// Get comments
	commentRows, err := s.queries.GetCommentsByTask(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task comments: %w", err)
	}

	// Convert to model
	detail := &models.TaskDetail{
		ID:          int(taskRow.ID),
		Title:       taskRow.Title,
		Description: taskRow.Description.String,
		ColumnID:    int(taskRow.ColumnID),
		ColumnName:  taskRow.ColumnName,
		ProjectName: taskRow.ProjectName,
		Position:    int(taskRow.Position),
		Labels:      converters.LabelsToModels(labels),
		ParentTasks: converters.ParentTasksToReferences(parentRows),
		ChildTasks:  converters.ChildTasksToReferences(childRows),
		Comments:    converters.CommentsToModels(commentRows),
		IsBlocked:   taskRow.IsBlocked,
	}

	if taskRow.TicketNumber.Valid {
		detail.TicketNumber = int(taskRow.TicketNumber.Int64)
	}
	if taskRow.TypeDescription.Valid {
		detail.TypeDescription = taskRow.TypeDescription.String
	}
	if taskRow.PriorityDescription.Valid {
		detail.PriorityDescription = taskRow.PriorityDescription.String
	}
	if taskRow.PriorityColor.Valid {
		detail.PriorityColor = taskRow.PriorityColor.String
	}
	if taskRow.CreatedAt.Valid {
		detail.CreatedAt = taskRow.CreatedAt.Time
	}
	if taskRow.UpdatedAt.Valid {
		detail.UpdatedAt = taskRow.UpdatedAt.Time
	}
	if taskRow.AssigneeID.Valid {
		id := int(taskRow.AssigneeID.Int64)
		detail.AssigneeID = &id
	}
	if taskRow.AssigneeName.Valid {
		name := taskRow.AssigneeName.String
		detail.AssigneeName = &name
	}
	if taskRow.Estimate.Valid {
		estimate := taskRow.Estimate.String
		detail.Estimate = &estimate
	}
	if taskRow.DueDate.Valid {
		detail.DueDate = &taskRow.DueDate.Time
	}

	return detail, nil
}

// GetTaskActivities retrieves all activity items (events and comments) for a task.
// Returns a unified list sorted by creation time (newest first).
func (s *service) GetTaskActivities(ctx context.Context, taskID int) ([]models.ActivityItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return nil, ErrInvalidTaskID
	}

	// Get events from event service (if available)
	var events []models.TaskEvent
	if s.eventService != nil {
		var err error
		events, err = s.eventService.GetEventsByTask(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task events: %w", err)
		}
	}

	// Get comments
	commentRows, err := s.queries.GetCommentsByTask(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task comments: %w", err)
	}

	// Convert comments to models
	comments := converters.CommentsToModels(commentRows)

	// Convert []*models.Comment to []models.Comment for MergeActivities
	commentSlice := make([]models.Comment, len(comments))
	for i, c := range comments {
		commentSlice[i] = *c
	}

	// Merge and sort activities (newest first)
	return models.MergeActivities(events, commentSlice), nil
}

// GetTaskTypeAndPriorityIDs retrieves only the type and priority IDs for a task
// This is a lightweight alternative to GetTaskDetail when only these IDs are needed
func (s *service) GetTaskTypeAndPriorityIDs(ctx context.Context, taskID int) (typeID, priorityID int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return 0, 0, ErrInvalidTaskID
	}

	row, err := s.queries.GetTaskTypeAndPriorityIDs(ctx, int64(taskID))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get task type and priority IDs: %w", err)
	}

	return int(row.TypeID), int(row.PriorityID), nil
}

// GetTaskSummariesByProject retrieves task summaries for a project, grouped by column
func (s *service) GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := s.queries.GetTaskSummariesByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task summaries: %w", err)
	}

	result := make(map[int][]*models.TaskSummary)
	for _, row := range rows {
		summary := converters.TaskSummaryFromRowToModel(row)
		columnID := int(row.ColumnID)
		result[columnID] = append(result[columnID], summary)
	}

	return result, nil
}

// GetTaskSummariesWithFilters retrieves task summaries using the unified filter query.
// All filter fields in params are optional — nil/zero values skip that filter.
func (s *service) GetTaskSummariesWithFilters(ctx context.Context, params TaskFilterParams) (map[int][]*models.TaskSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	dbParams := types.GetTaskSummariesWithFiltersParams{
		ProjectID: int64(params.ProjectID),
	}

	if params.Title != nil {
		dbParams.TitleFilter = types.NullString{String: "%" + *params.Title + "%", Valid: true}
	}
	if params.PriorityID != nil {
		dbParams.PriorityID = types.NullInt64{Int64: int64(*params.PriorityID), Valid: true}
	}
	if params.TypeID != nil {
		dbParams.TypeID = types.NullInt64{Int64: int64(*params.TypeID), Valid: true}
	}
	if params.AssigneeID != nil {
		dbParams.AssigneeID = types.NullInt64{Int64: int64(*params.AssigneeID), Valid: true}
	}
	if len(params.LabelIDs) > 0 {
		dbParams.LabelIdsCsv = buildLabelIdsCsv(params.LabelIDs)
	}

	rows, err := s.queries.GetTaskSummariesWithFilters(ctx, dbParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered task summaries: %w", err)
	}

	result := make(map[int][]*models.TaskSummary)
	for _, row := range rows {
		summary := converters.WithFiltersTaskSummaryFromRowToModel(row)
		columnID := int(row.ColumnID)
		result[columnID] = append(result[columnID], summary)
	}

	return result, nil
}

// buildLabelIdsCsv constructs a comma-wrapped CSV string for SQL label matching.
// Example: [1, 5, 12] -> ",1,5,12,"
func buildLabelIdsCsv(ids []int) string {
	csv := ","
	for i, id := range ids {
		if i > 0 {
			csv += ","
		}
		csv += fmt.Sprintf("%d", id)
	}
	csv += ","
	return csv
}

// GetReadyTaskSummariesByProject retrieves task summaries for tasks in ready columns (and not blocked)
func (s *service) GetReadyTaskSummariesByProject(ctx context.Context, projectID int) ([]*models.TaskSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := s.queries.GetReadyTaskSummariesByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ready task summaries: %w", err)
	}

	result := make([]*models.TaskSummary, 0, len(rows))
	for _, row := range rows {
		if row.IsBlocked {
			continue
		}
		summary := converters.ReadyTaskSummaryFromRowToModel(row)
		result = append(result, summary)
	}

	return result, nil
}

// GetTaskReferencesForProject retrieves task references for a project
func (s *service) GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := s.queries.GetTaskReferencesForProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task references: %w", err)
	}

	references := make([]*models.TaskReference, 0, len(rows))
	for _, row := range rows {
		ref := &models.TaskReference{
			ID:          int(row.ID),
			Title:       row.Title,
			ProjectName: row.Name,
		}
		if row.TicketNumber.Valid {
			ref.TicketNumber = int(row.TicketNumber.Int64)
		}
		references = append(references, ref)
	}

	return references, nil
}

// childRelation is a helper struct for building the tree
type childRelation struct {
	childID       int
	relationLabel string
	relationColor string
	isBlocking    bool
}

// GetTaskTreeByProject builds a hierarchical tree of tasks for a project
// Returns root tasks (tasks with no parents) with their children nested recursively
func (s *service) GetTaskTreeByProject(ctx context.Context, projectID int) ([]*models.TaskTreeNode, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get all tasks in the project
	taskRows, err := s.queries.GetTasksForTree(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for tree: %w", err)
	}

	if len(taskRows) == 0 {
		return []*models.TaskTreeNode{}, nil
	}

	// Get all relations for the project
	relationRows, err := s.queries.GetTaskRelationsForProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task relations: %w", err)
	}

	// Build a map of task ID -> task info
	taskMap := make(map[int]*models.TaskTreeNode)
	for _, row := range taskRows {
		node := &models.TaskTreeNode{
			ID:          int(row.ID),
			Title:       row.Title,
			ColumnName:  row.ColumnName,
			ProjectName: row.ProjectName,
			IsCompleted: row.IsCompleted,
			Children:    []*models.TaskTreeNode{},
		}
		if row.TicketNumber.Valid {
			node.TicketNumber = int(row.TicketNumber.Int64)
		}
		taskMap[node.ID] = node
	}

	// Build parent -> children map and track which tasks have parents
	hasParent := make(map[int]bool)
	childrenByParent := make(map[int][]*childRelation)

	for _, rel := range relationRows {
		parentID := int(rel.ParentID)
		childID := int(rel.ChildID)

		hasParent[childID] = true

		childrenByParent[parentID] = append(childrenByParent[parentID], &childRelation{
			childID:       childID,
			relationLabel: rel.RelationLabel,
			relationColor: rel.RelationColor,
			isBlocking:    rel.IsBlocking,
		})
	}

	// Build the tree structure
	// For each parent, attach its children with relation info
	visited := make(map[int]bool)
	var buildChildren func(parentID int, depth int) []*models.TaskTreeNode
	buildChildren = func(parentID int, depth int) []*models.TaskTreeNode {
		// Prevent infinite loops from circular dependencies
		if depth > 100 || visited[parentID] {
			return nil
		}
		visited[parentID] = true
		defer func() { visited[parentID] = false }()

		children := childrenByParent[parentID]
		if len(children) == 0 {
			return nil
		}

		result := make([]*models.TaskTreeNode, 0, len(children))
		for _, childRel := range children {
			childNode, exists := taskMap[childRel.childID]
			if !exists {
				continue
			}

			// Create a copy with relation info for this specific parent-child relationship
			nodeCopy := &models.TaskTreeNode{
				ID:            childNode.ID,
				TicketNumber:  childNode.TicketNumber,
				Title:         childNode.Title,
				ColumnName:    childNode.ColumnName,
				ProjectName:   childNode.ProjectName,
				RelationLabel: childRel.relationLabel,
				RelationColor: childRel.relationColor,
				IsBlocking:    childRel.isBlocking,
				IsCompleted:   childNode.IsCompleted,
				Children:      buildChildren(childRel.childID, depth+1),
			}
			result = append(result, nodeCopy)
		}
		return result
	}

	// Find root tasks (tasks with no parents)
	var roots []*models.TaskTreeNode
	for _, node := range taskMap {
		if !hasParent[node.ID] {
			// This is a root task - build its children
			rootCopy := &models.TaskTreeNode{
				ID:           node.ID,
				TicketNumber: node.TicketNumber,
				Title:        node.Title,
				ColumnName:   node.ColumnName,
				ProjectName:  node.ProjectName,
				IsCompleted:  node.IsCompleted,
				Children:     buildChildren(node.ID, 0),
			}
			roots = append(roots, rootCopy)
		}
	}

	// Sort roots by ticket number for deterministic output order
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].TicketNumber < roots[j].TicketNumber
	})

	return roots, nil
}

// extractColumnID safely converts the types.NullInt64 result from GetNextColumnID or GetPrevColumnID
// to an int64, handling NULL values properly.
func extractColumnID(columnID types.NullInt64) (int64, error) {
	if !columnID.Valid {
		return 0, fmt.Errorf("failed to extract column ID: value is NULL")
	}
	return columnID.Int64, nil
}

// MoveTaskToNextColumn moves task to next column
func (s *service) MoveTaskToNextColumn(ctx context.Context, taskID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	// Get current column
	posRow, err := s.queries.GetTaskPosition(ctx, int64(taskID))
	if err != nil {
		return fmt.Errorf("failed to get task position: %w", err)
	}

	// Get current column name for event
	var fromColumnName string
	if s.eventService != nil {
		fromCol, err := s.queries.GetColumnByID(ctx, posRow.ColumnID)
		if err == nil {
			fromColumnName = fromCol.Name
		}
	}

	// Get next column
	nextColumnID, err := s.queries.GetNextColumnID(ctx, posRow.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get next column: %w", err)
	}

	// Convert any to int64 with proper error handling
	nextColID, err := extractColumnID(nextColumnID)
	if err != nil {
		return fmt.Errorf("failed to move task: no next column available")
	}

	// Get task count in target column to append at the end
	taskCount, err := s.queries.GetTaskCountByColumn(ctx, nextColID)
	if err != nil {
		return fmt.Errorf("failed to get task count: %w", err)
	}

	// Move task to next column
	if err := s.queries.MoveTaskToColumn(ctx, types.MoveTaskToColumnParams{
		ColumnID: nextColID,
		Position: taskCount + 1,
		ID:       int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	// Emit TaskMoved event
	if s.eventService != nil && fromColumnName != "" {
		toCol, err := s.queries.GetColumnByID(ctx, nextColID)
		if err == nil {
			if err := s.eventService.CreateTaskMovedEvent(ctx, s.queries, taskID, fromColumnName, toCol.Name, user.GetCurrentUsername()); err != nil {
				slog.Warn("failed to create task moved event", "error", err, "taskID", taskID, "eventType", "TaskMoved")
			}
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// MoveTaskToPrevColumn moves task to previous column
func (s *service) MoveTaskToPrevColumn(ctx context.Context, taskID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	// Get current column
	posRow, err := s.queries.GetTaskPosition(ctx, int64(taskID))
	if err != nil {
		return fmt.Errorf("failed to get task position: %w", err)
	}

	// Get current column name for event
	var fromColumnName string
	if s.eventService != nil {
		fromCol, err := s.queries.GetColumnByID(ctx, posRow.ColumnID)
		if err == nil {
			fromColumnName = fromCol.Name
		}
	}

	// Get previous column
	prevColumnID, err := s.queries.GetPrevColumnID(ctx, posRow.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get previous column: %w", err)
	}

	// Convert any to int64 with proper error handling
	prevColID, err := extractColumnID(prevColumnID)
	if err != nil {
		return fmt.Errorf("failed to move task: no previous column available")
	}

	// Get task count in target column to append at the end
	taskCount, err := s.queries.GetTaskCountByColumn(ctx, prevColID)
	if err != nil {
		return fmt.Errorf("failed to get task count: %w", err)
	}

	// Move task to previous column
	if err := s.queries.MoveTaskToColumn(ctx, types.MoveTaskToColumnParams{
		ColumnID: prevColID,
		Position: taskCount + 1,
		ID:       int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	// Emit TaskMoved event
	if s.eventService != nil && fromColumnName != "" {
		toCol, err := s.queries.GetColumnByID(ctx, prevColID)
		if err == nil {
			if err := s.eventService.CreateTaskMovedEvent(ctx, s.queries, taskID, fromColumnName, toCol.Name, user.GetCurrentUsername()); err != nil {
				slog.Warn("failed to create task moved event", "error", err, "taskID", taskID, "eventType", "TaskMoved")
			}
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// MoveTaskToColumn moves task to specific column
func (s *service) MoveTaskToColumn(ctx context.Context, taskID, columnID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}
	if err := validateColumnID(columnID); err != nil {
		return err
	}

	// Get current column info before moving (for event)
	var fromColumnName string
	taskPos, err := s.queries.GetTaskPosition(ctx, int64(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("failed to verify task exists: %w", err)
	}

	if s.eventService != nil {
		fromCol, err := s.queries.GetColumnByID(ctx, taskPos.ColumnID)
		if err == nil {
			fromColumnName = fromCol.Name
		}
	}

	// Get task count in target column to append at the end
	taskCount, err := s.queries.GetTaskCountByColumn(ctx, int64(columnID))
	if err != nil {
		return fmt.Errorf("failed to get task count: %w", err)
	}

	if err := s.queries.MoveTaskToColumn(ctx, types.MoveTaskToColumnParams{
		ColumnID: int64(columnID),
		Position: taskCount + 1,
		ID:       int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	// Emit TaskMoved event if column actually changed
	if s.eventService != nil && fromColumnName != "" {
		toCol, err := s.queries.GetColumnByID(ctx, int64(columnID))
		if err == nil && toCol.Name != fromColumnName {
			if err := s.eventService.CreateTaskMovedEvent(ctx, s.queries, taskID, fromColumnName, toCol.Name, user.GetCurrentUsername()); err != nil {
				slog.Warn("failed to create task moved event", "error", err, "taskID", taskID, "eventType", "TaskMoved")
			}
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// specialColumnFinder is a function type that finds a special column for a project.
// It takes a project ID and returns the column ID or an error.
type specialColumnFinder func(ctx context.Context, projectID int64) (int64, error)

// moveToSpecialColumn is a helper that handles the common logic for moving tasks
// to special columns (ready, completed, in-progress). It validates the task,
// finds the project, gets the target column via the finder function, and moves the task.
func (s *service) moveToSpecialColumn(ctx context.Context, taskID int, findColumn specialColumnFinder, columnTypeName string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	// Get task detail to find project
	taskDetail, err := s.queries.GetTaskDetail(ctx, int64(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get the column via project ID
	column, err := s.queries.GetColumnByID(ctx, taskDetail.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get column: %w", err)
	}

	// Get target special column for project
	targetColumnID, err := findColumn(ctx, column.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no %s column configured for this project", columnTypeName)
		}
		return fmt.Errorf("failed to get %s column: %w", columnTypeName, err)
	}

	// Check if already in target column
	if taskDetail.ColumnID == targetColumnID {
		return ErrTaskAlreadyInTargetColumn
	}

	// Move task to target column
	return s.MoveTaskToColumn(ctx, taskID, int(targetColumnID))
}

// MoveTaskToReadyColumn moves task to the column marked as holding ready tasks
func (s *service) MoveTaskToReadyColumn(ctx context.Context, taskID int) error {
	return s.moveToSpecialColumn(ctx, taskID, func(ctx context.Context, projectID int64) (int64, error) {
		col, err := s.queries.GetReadyColumnByProject(ctx, projectID)
		return col.ID, err
	}, "ready")
}

// MoveTaskToCompletedColumn moves task to the column marked as holding completed tasks
func (s *service) MoveTaskToCompletedColumn(ctx context.Context, taskID int) error {
	return s.moveToSpecialColumn(ctx, taskID, func(ctx context.Context, projectID int64) (int64, error) {
		col, err := s.queries.GetCompletedColumnByProject(ctx, projectID)
		return col.ID, err
	}, "completed")
}

// MoveTaskToInProgressColumn moves a task to the column marked as holding in-progress tasks
func (s *service) MoveTaskToInProgressColumn(ctx context.Context, taskID int) error {
	return s.moveToSpecialColumn(ctx, taskID, func(ctx context.Context, projectID int64) (int64, error) {
		col, err := s.queries.GetInProgressColumnByProject(ctx, projectID)
		return col.ID, err
	}, "in-progress")
}

// MoveTaskToProject moves a task to the ready column of a different project.
// Returns ErrTaskAlreadyInTargetProject if the task is already in that project.
func (s *service) MoveTaskToProject(ctx context.Context, taskID int, targetProjectID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	// Get current project ID for this task (single lightweight query)
	currentProjectID, err := s.queries.GetProjectIDFromTask(ctx, int64(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task is already in the target project
	if currentProjectID == int64(targetProjectID) {
		return ErrTaskAlreadyInTargetProject
	}

	// Verify the target project exists
	if _, err := s.queries.GetProjectByID(ctx, int64(targetProjectID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("target project not found")
		}
		return fmt.Errorf("failed to get target project: %w", err)
	}

	// Get the ready column in the target project
	readyColumn, err := s.queries.GetReadyColumnByProject(ctx, int64(targetProjectID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no ready column configured for target project")
		}
		return fmt.Errorf("failed to get ready column for target project: %w", err)
	}

	// Move the task to the ready column of the target project
	return s.MoveTaskToColumn(ctx, taskID, int(readyColumn.ID))
}

// GetInProgressTasksByProject retrieves all tasks in in-progress columns for a project
// OPTIMIZED: Uses a single efficiently-designed query that fetches task details,
// labels, and blocking status in one go. This avoids the N+1 query problem where
// the old implementation called GetTaskDetail for each task, which itself made
// 5+ additional queries (labels, comments, parents, children, etc).
//
// Performance improvement: Reduces queries from O(N*5+) to O(1)
// where N is the number of in-progress tasks.
func (s *service) GetInProgressTasksByProject(ctx context.Context, projectID int) ([]*models.TaskDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateProjectID(projectID); err != nil {
		return nil, ErrInvalidProjectID
	}

	// Fetch all in-progress tasks with their details in a SINGLE efficient query
	detailRows, err := s.queries.GetInProgressTaskDetails(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get in-progress tasks: %w", err)
	}

	tasks := make([]*models.TaskDetail, 0, len(detailRows))
	for _, row := range detailRows {
		// Convert row to TaskDetail - no additional queries needed!
		taskDetail := &models.TaskDetail{
			ID:           int(row.ID),
			Title:        row.Title,
			Description:  row.Description.String,
			ColumnID:     int(row.ColumnID),
			ColumnName:   row.ColumnName,
			ProjectName:  row.ProjectName,
			Position:     int(row.Position),
			TicketNumber: int(row.TicketNumber.Int64),
			Labels:       converters.ParseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors),
			IsBlocked:    row.IsBlocked,
			CreatedAt:    row.CreatedAt.Time,
			UpdatedAt:    row.UpdatedAt.Time,
		}

		// Set optional fields
		if row.TypeDescription.Valid {
			taskDetail.TypeDescription = row.TypeDescription.String
		}
		if row.PriorityDescription.Valid {
			taskDetail.PriorityDescription = row.PriorityDescription.String
		}
		if row.PriorityColor.Valid {
			taskDetail.PriorityColor = row.PriorityColor.String
		}
		if row.AssigneeID.Valid {
			id := int(row.AssigneeID.Int64)
			taskDetail.AssigneeID = &id
		}
		if row.AssigneeName.Valid {
			name := row.AssigneeName.String
			taskDetail.AssigneeName = &name
		}
		if row.Estimate.Valid {
			estimate := row.Estimate.String
			taskDetail.Estimate = &estimate
		}
		if row.DueDate.Valid {
			taskDetail.DueDate = &row.DueDate.Time
		}

		// NOTE: Parent/child tasks and comments are NOT fetched here.
		// If full task relationships are needed, use GetTaskDetail instead,
		// but only for the specific tasks that require it, not in bulk.

		tasks = append(tasks, taskDetail)
	}

	return tasks, nil
}

// MoveTaskUp moves task up in its column
func (s *service) MoveTaskUp(ctx context.Context, taskID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	// Use WithTx helper to avoid UNIQUE constraint violations during swap
	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := database.MustNewQuerier(tx, s.dbType)

		// Get task position
		posRow, err := qtx.GetTaskPosition(ctx, int64(taskID))
		if err != nil {
			return fmt.Errorf("failed to get task position: %w", err)
		}

		// Get task above
		aboveRow, err := qtx.GetTaskAbove(ctx, types.GetTaskAboveParams(posRow))
		if err != nil {
			return fmt.Errorf("no task above: %w", err)
		}

		// Swap positions using temporary negative position to avoid UNIQUE constraint violation
		// Step 1: Move current task to temporary position
		if err := qtx.SetTaskPosition(ctx, types.SetTaskPositionParams{
			Position: -1,
			ID:       int64(taskID),
		}); err != nil {
			return fmt.Errorf("failed to set temporary position: %w", err)
		}

		// Step 2: Move task above to current task's position
		if err := qtx.SetTaskPosition(ctx, types.SetTaskPositionParams{
			Position: posRow.Position,
			ID:       aboveRow.ID,
		}); err != nil {
			return fmt.Errorf("failed to move other task down: %w", err)
		}

		// Step 3: Move current task to above position
		if err := qtx.SetTaskPosition(ctx, types.SetTaskPositionParams{
			Position: aboveRow.Position,
			ID:       int64(taskID),
		}); err != nil {
			return fmt.Errorf("failed to move task up: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// MoveTaskDown moves task down in its column
func (s *service) MoveTaskDown(ctx context.Context, taskID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}

	// Use WithTx helper to avoid UNIQUE constraint violations during swap
	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := database.MustNewQuerier(tx, s.dbType)

		// Get task position
		posRow, err := qtx.GetTaskPosition(ctx, int64(taskID))
		if err != nil {
			return fmt.Errorf("failed to get task position: %w", err)
		}

		// Get task below
		belowRow, err := qtx.GetTaskBelow(ctx, types.GetTaskBelowParams(posRow))
		if err != nil {
			return fmt.Errorf("no task below: %w", err)
		}

		// Swap positions using temporary negative position to avoid UNIQUE constraint violation
		// Step 1: Move current task to temporary position
		if err := qtx.SetTaskPosition(ctx, types.SetTaskPositionParams{
			Position: -1,
			ID:       int64(taskID),
		}); err != nil {
			return fmt.Errorf("failed to set temporary position: %w", err)
		}

		// Step 2: Move task below to current task's position
		if err := qtx.SetTaskPosition(ctx, types.SetTaskPositionParams{
			Position: posRow.Position,
			ID:       belowRow.ID,
		}); err != nil {
			return fmt.Errorf("failed to move other task up: %w", err)
		}

		// Step 3: Move current task to below position
		if err := qtx.SetTaskPosition(ctx, types.SetTaskPositionParams{
			Position: belowRow.Position,
			ID:       int64(taskID),
		}); err != nil {
			return fmt.Errorf("failed to move task down: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// wouldCreateCycle checks if adding a parent->child relationship would create a cycle
// by checking if parentID is reachable from childID through existing relationships
func (s *service) wouldCreateCycle(ctx context.Context, parentID, childID int) (bool, error) {
	// Use BFS to check if parentID is reachable from childID
	visited := make(map[int]bool)
	queue := []int{childID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		// If we reach the parent, we found a cycle
		if currentID == parentID {
			return true, nil
		}

		// Get all children of current task
		children, err := s.queries.GetChildTasks(ctx, int64(currentID))
		if err != nil {
			return false, fmt.Errorf("failed to get child tasks: %w", err)
		}

		// Add unvisited children to queue
		for _, child := range children {
			childTaskID := int(child.ID)
			if !visited[childTaskID] {
				queue = append(queue, childTaskID)
			}
		}
	}

	return false, nil
}

// getRelationLabel looks up the label for a relation type.
// useChildToParent determines whether to return the child-to-parent label (true)
// or the parent-to-child label (false).
func (s *service) getRelationLabel(ctx context.Context, relationTypeID int, useChildToParent bool, defaultLabel string) string {
	relTypes, err := s.queries.GetAllRelationTypes(ctx)
	if err != nil {
		return defaultLabel
	}
	for _, rt := range relTypes {
		if rt.ID == int64(relationTypeID) {
			if useChildToParent {
				return rt.CToPLabel
			}
			return rt.PToCLabel
		}
	}
	return defaultLabel
}

// AddParentRelation adds a parent relationship (parent depends on this task)
func (s *service) AddParentRelation(ctx context.Context, taskID, parentID int, relationTypeID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}
	if err := validateTaskID(parentID); err != nil {
		return err
	}
	if taskID == parentID {
		return ErrSelfRelation
	}

	// Check for circular dependency before adding
	wouldCycle, err := s.wouldCreateCycle(ctx, parentID, taskID)
	if err != nil {
		return err
	}
	if wouldCycle {
		return ErrCircularRelation
	}

	// Add the relationship (this task is the child)
	if err := s.queries.AddSubtaskWithRelationType(ctx, types.AddSubtaskWithRelationTypeParams{
		ParentID:       int64(parentID),
		ChildID:        int64(taskID),
		RelationTypeID: int64(relationTypeID),
	}); err != nil {
		return fmt.Errorf("failed to add parent relation: %w", err)
	}

	// Emit TaskAssociated event
	if s.eventService != nil {
		relationLabel := s.getRelationLabel(ctx, relationTypeID, true, "Parent")
		parentTask, err := s.queries.GetTask(ctx, int64(parentID))
		if err == nil {
			if err := s.eventService.CreateTaskAssociatedEvent(ctx, s.queries, taskID, parentID, parentTask.Title, relationLabel, user.GetCurrentUsername()); err != nil {
				slog.Warn("failed to create task associated event", "error", err, "taskID", taskID, "eventType", "TaskAssociated")
			}
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// AddChildRelation adds a child relationship (this task depends on child)
func (s *service) AddChildRelation(ctx context.Context, taskID, childID int, relationTypeID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}
	if err := validateTaskID(childID); err != nil {
		return err
	}
	if taskID == childID {
		return ErrSelfRelation
	}

	// Check for circular dependency before adding
	wouldCycle, err := s.wouldCreateCycle(ctx, taskID, childID)
	if err != nil {
		return err
	}
	if wouldCycle {
		return ErrCircularRelation
	}

	// Add the relationship (this task is the parent)
	if err := s.queries.AddSubtaskWithRelationType(ctx, types.AddSubtaskWithRelationTypeParams{
		ParentID:       int64(taskID),
		ChildID:        int64(childID),
		RelationTypeID: int64(relationTypeID),
	}); err != nil {
		return fmt.Errorf("failed to add child relation: %w", err)
	}

	// Emit TaskAssociated event
	if s.eventService != nil {
		relationLabel := s.getRelationLabel(ctx, relationTypeID, false, "Child")
		childTask, err := s.queries.GetTask(ctx, int64(childID))
		if err == nil {
			if err := s.eventService.CreateTaskAssociatedEvent(ctx, s.queries, taskID, childID, childTask.Title, relationLabel, user.GetCurrentUsername()); err != nil {
				slog.Warn("failed to create task associated event", "error", err, "taskID", taskID, "eventType", "TaskAssociated")
			}
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// RemoveParentRelation removes a parent relationship
func (s *service) RemoveParentRelation(ctx context.Context, taskID, parentID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}
	if err := validateTaskID(parentID); err != nil {
		return err
	}

	if err := s.queries.RemoveSubtask(ctx, types.RemoveSubtaskParams{
		ParentID: int64(parentID),
		ChildID:  int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to remove parent relation: %w", err)
	}

	// Emit TaskDisassociated event
	if s.eventService != nil {
		if err := s.eventService.CreateTaskDisassociatedEvent(ctx, s.queries, taskID, parentID, user.GetCurrentUsername()); err != nil {
			slog.Warn("failed to create task disassociated event", "error", err, "taskID", taskID, "eventType", "TaskDisassociated")
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// RemoveChildRelation removes a child relationship
func (s *service) RemoveChildRelation(ctx context.Context, taskID, childID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}
	if err := validateTaskID(childID); err != nil {
		return err
	}

	if err := s.queries.RemoveSubtask(ctx, types.RemoveSubtaskParams{
		ParentID: int64(taskID),
		ChildID:  int64(childID),
	}); err != nil {
		return fmt.Errorf("failed to remove child relation: %w", err)
	}

	// Emit TaskDisassociated event
	if s.eventService != nil {
		if err := s.eventService.CreateTaskDisassociatedEvent(ctx, s.queries, taskID, childID, user.GetCurrentUsername()); err != nil {
			slog.Warn("failed to create task disassociated event", "error", err, "taskID", taskID, "eventType", "TaskDisassociated")
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// AttachLabel attaches a label to a task
func (s *service) AttachLabel(ctx context.Context, taskID, labelID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}
	if err := validateLabelID(labelID); err != nil {
		return err
	}

	if err := s.queries.AddLabelToTask(ctx, types.AddLabelToTaskParams{
		TaskID:  int64(taskID),
		LabelID: int64(labelID),
	}); err != nil {
		return fmt.Errorf("failed to attach label: %w", err)
	}

	// Emit LabelAdded event
	if s.eventService != nil {
		label, err := s.queries.GetLabelByID(ctx, int64(labelID))
		if err == nil {
			if err := s.eventService.CreateLabelAddedEvent(ctx, s.queries, taskID, label.Name, user.GetCurrentUsername()); err != nil {
				slog.Warn("failed to create label added event", "error", err, "taskID", taskID, "eventType", "LabelAdded")
			}
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// DetachLabel detaches a label from a task
func (s *service) DetachLabel(ctx context.Context, taskID, labelID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return err
	}
	if err := validateLabelID(labelID); err != nil {
		return err
	}

	// Get label name before removal for event
	var labelName string
	if s.eventService != nil {
		label, err := s.queries.GetLabelByID(ctx, int64(labelID))
		if err == nil {
			labelName = label.Name
		}
	}

	if err := s.queries.RemoveLabelFromTask(ctx, types.RemoveLabelFromTaskParams{
		TaskID:  int64(taskID),
		LabelID: int64(labelID),
	}); err != nil {
		return fmt.Errorf("failed to detach label: %w", err)
	}

	// Emit LabelRemoved event
	if s.eventService != nil && labelName != "" {
		if err := s.eventService.CreateLabelRemovedEvent(ctx, s.queries, taskID, labelName, user.GetCurrentUsername()); err != nil {
			slog.Warn("failed to create label removed event", "error", err, "taskID", taskID, "eventType", "LabelRemoved")
		}
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// CreateComment creates a new comment on a task
func (s *service) CreateComment(ctx context.Context, req CreateCommentRequest) (*models.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Validate request
	if err := validateCreateCommentRequest(req); err != nil {
		return nil, err
	}

	// Verify task exists before creating comment
	_, err := s.queries.GetTask(ctx, int64(req.TaskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to verify task exists: %w", err)
	}

	comment, err := s.queries.CreateComment(ctx, types.CreateCommentParams{
		TaskID:  int64(req.TaskID),
		Content: req.Message,
		Author:  req.Author,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	s.publishTaskEvent(ctx, req.TaskID)

	return &models.Comment{
		ID:        int(comment.ID),
		TaskID:    int(comment.TaskID),
		Message:   comment.Content,
		Author:    comment.Author,
		CreatedAt: comment.CreatedAt.Time,
	}, nil
}

// UpdateComment updates a comment's message
func (s *service) UpdateComment(ctx context.Context, req UpdateCommentRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Validate request
	if err := validateUpdateCommentRequest(req); err != nil {
		return err
	}

	comment, err := s.queries.GetComment(ctx, int64(req.CommentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("failed to get comment: %w", err)
	}

	if err := s.queries.UpdateComment(ctx, types.UpdateCommentParams{
		Content: req.Message,
		ID:      int64(req.CommentID),
	}); err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	s.publishTaskEvent(ctx, int(comment.TaskID))
	return nil
}

// DeleteComment deletes a comment
func (s *service) DeleteComment(ctx context.Context, commentID int) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateCommentID(commentID); err != nil {
		return err
	}

	comment, err := s.queries.GetComment(ctx, int64(commentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("failed to get comment: %w", err)
	}

	if err := s.queries.DeleteComment(ctx, int64(commentID)); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	s.publishTaskEvent(ctx, int(comment.TaskID))
	return nil
}

// GetCommentsByTask retrieves all comments for a task
func (s *service) GetCommentsByTask(ctx context.Context, taskID int) ([]*models.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := validateTaskID(taskID); err != nil {
		return nil, ErrInvalidTaskID
	}

	rows, err := s.queries.GetCommentsByTask(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return converters.CommentsToModels(rows), nil
}

// publishTaskEvent publishes a task event with retry logic
func (s *service) publishTaskEvent(ctx context.Context, taskID int) {
	if s.eventClient == nil {
		return
	}

	// Get project ID for the task
	projectID, err := s.queries.GetProjectIDFromTask(ctx, int64(taskID))
	if err != nil {
		slog.Error("failed to retrieve project ID for task event publishing",
			"task_id", taskID,
			"error", err.Error(),
		)
		return
	}

	s.publishTaskEventForProject(int(projectID))
}

// publishTaskEventForProject publishes a task event for a known project ID.
// Use this when the project ID is already known (e.g., after task deletion
// where the task no longer exists in the database).
func (s *service) publishTaskEventForProject(projectID int) {
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
