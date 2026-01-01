package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"

	"github.com/thenoetrevino/paso/internal/converters"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// taskSummaryMapPool is a sync.Pool for reusing maps in batch processing operations
// This reduces allocations when frequently querying task summaries
var taskSummaryMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[int][]*models.TaskSummary)
	},
}

// ============================================================================
// SEGREGATED INTERFACES - Following Interface Segregation Principle (ISP)
// ============================================================================

// TaskReader defines read-only operations for retrieving task data.
// This interface segregates read operations, making it easier to mock for testing
// and understand which operations don't modify state.
//
// Use this interface when you only need to retrieve task information without
// modification capabilities. This reduces coupling and makes testing easier.
type TaskReader interface {
	// Get single task details
	GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error)

	// Get task summaries/lists grouped by column
	GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error)
	GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error)
	GetReadyTaskSummariesByProject(ctx context.Context, projectID int) ([]*models.TaskSummary, error)
	GetInProgressTasksByProject(ctx context.Context, projectID int) ([]*models.TaskDetail, error)

	// Get task references and hierarchies
	GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error)
	GetTaskTreeByProject(ctx context.Context, projectID int) ([]*models.TaskTreeNode, error)
}

// TaskWriter defines write operations for creating, updating, and deleting tasks.
// This interface segregates state-modifying operations, allowing clients to specify
// exactly what write capabilities they need.
//
// Use this interface when you need to modify tasks (create, update, delete) but don't
// need movement or relationship operations. This provides focused control over write access.
type TaskWriter interface {
	// Create and update operations
	CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error)
	UpdateTask(ctx context.Context, req UpdateTaskRequest) error
	DeleteTask(ctx context.Context, taskID int) error
}

// TaskMover defines task movement operations within the task management system.
// This interface segregates operations that change task position/column, making
// it clear which operations affect task workflow state.
//
// Use this interface when you need to move tasks between columns or reorder them,
// but don't need other write operations like creation or deletion.
type TaskMover interface {
	// Column-based movement (workflow progression)
	MoveTaskToNextColumn(ctx context.Context, taskID int) error
	MoveTaskToPrevColumn(ctx context.Context, taskID int) error
	MoveTaskToColumn(ctx context.Context, taskID, columnID int) error
	MoveTaskToReadyColumn(ctx context.Context, taskID int) error
	MoveTaskToCompletedColumn(ctx context.Context, taskID int) error
	MoveTaskToInProgressColumn(ctx context.Context, taskID int) error

	// Position-based movement (ordering within column)
	MoveTaskUp(ctx context.Context, taskID int) error
	MoveTaskDown(ctx context.Context, taskID int) error
}

// TaskRelationer defines task relationship operations (parent/child/blocking relationships).
// This interface segregates relationship management operations, allowing fine-grained
// control over which clients can modify task dependencies.
//
// Use this interface when you need to manage task relationships (dependencies, blocking)
// but don't need other modification operations.
type TaskRelationer interface {
	// Parent/child relationships
	AddParentRelation(ctx context.Context, taskID, parentID int, relationTypeID int) error
	AddChildRelation(ctx context.Context, taskID, childID int, relationTypeID int) error
	RemoveParentRelation(ctx context.Context, taskID, parentID int) error
	RemoveChildRelation(ctx context.Context, taskID, childID int) error
}

// TaskLabeler defines label management operations for tasks.
// This interface segregates label operations, providing focused control over
// label attachment and detachment.
//
// Use this interface when you only need to manage labels for tasks, allowing
// independent control over label operations.
type TaskLabeler interface {
	// Label management
	AttachLabel(ctx context.Context, taskID, labelID int) error
	DetachLabel(ctx context.Context, taskID, labelID int) error
}

// TaskCommenter defines comment operations on tasks.
// This interface segregates comment management, allowing independent control
// over comment creation, updates, and deletion.
//
// Use this interface when you only need to manage comments on tasks,
// independent from other task operations.
type TaskCommenter interface {
	// Comment operations
	CreateComment(ctx context.Context, req CreateCommentRequest) (*models.Comment, error)
	UpdateComment(ctx context.Context, req UpdateCommentRequest) error
	DeleteComment(ctx context.Context, commentID int) error
	GetCommentsByTask(ctx context.Context, taskID int) ([]*models.Comment, error)
}

// Service defines all task-related business operations as a composition of focused interfaces.
// This composite interface maintains backward compatibility while providing better separation of concerns.
//
// Components that need only specific operations (e.g., reading, writing, moving) should depend on
// the corresponding focused interface (TaskReader, TaskWriter, TaskMover, etc.) instead of this
// composite interface. This makes the system more testable and maintainable by reducing coupling.
//
// The service implementation satisfies all segregated interfaces, so existing code can continue
// to use the full Service interface without changes.
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
	PriorityID   int // Optional: 0 means use default
	TypeID       int // Optional: 0 means use default
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

// service implements Service interface using SQLC directly
type service struct {
	db          *sql.DB
	queries     generated.Querier
	eventClient events.EventPublisher
}

// NewService creates a new task service with SQLC queries
func NewService(db *sql.DB, eventClient events.EventPublisher) Service {
	return &service{
		db:          db,
		queries:     generated.New(db),
		eventClient: eventClient,
	}
}

// CreateTask handles task creation with validation and business rules
func (s *service) CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error) {
	// Validate request
	if err := s.validateCreateTask(req); err != nil {
		return nil, err
	}

	// Get project ID from column
	projectID, err := s.queries.GetProjectIDFromColumn(ctx, int64(req.ColumnID))
	if err != nil {
		return nil, fmt.Errorf("failed to get project ID: %w", err)
	}

	var createdTask generated.Task

	// Use WithTx helper for transaction management
	err = database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := generated.New(tx)

		// Get next ticket number
		ticketNumber, err := qtx.GetNextTicketNumber(ctx, projectID)
		if err != nil {
			return fmt.Errorf("failed to get ticket number: %w", err)
		}

		// Create task
		var desc sql.NullString
		if req.Description != "" {
			desc = sql.NullString{String: req.Description, Valid: true}
		}

		var taskErr error
		createdTask, taskErr = qtx.CreateTask(ctx, generated.CreateTaskParams{
			Title:        req.Title,
			Description:  desc,
			ColumnID:     int64(req.ColumnID),
			Position:     int64(req.Position),
			TicketNumber: ticketNumber,
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
			if err := qtx.UpdateTaskPriority(ctx, generated.UpdateTaskPriorityParams{
				PriorityID: int64(req.PriorityID),
				ID:         createdTask.ID,
			}); err != nil {
				return fmt.Errorf("failed to set priority: %w", err)
			}
		}

		// Set type if provided (default is handled by database)
		if req.TypeID > 0 {
			if err := qtx.UpdateTaskType(ctx, generated.UpdateTaskTypeParams{
				TypeID: int64(req.TypeID),
				ID:     createdTask.ID,
			}); err != nil {
				return fmt.Errorf("failed to set type: %w", err)
			}
		}

		// Attach labels
		for _, labelID := range req.LabelIDs {
			if err := qtx.AddLabelToTask(ctx, generated.AddLabelToTaskParams{
				TaskID:  createdTask.ID,
				LabelID: int64(labelID),
			}); err != nil {
				return fmt.Errorf("failed to attach label %d: %w", labelID, err)
			}
		}

		// Add parent relationships (tasks that depend on this task)
		for _, parentID := range req.ParentIDs {
			if err := qtx.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
				ParentID:       int64(parentID),
				ChildID:        createdTask.ID,
				RelationTypeID: 1,
			}); err != nil {
				return fmt.Errorf("failed to add parent relation: %w", err)
			}
		}

		// Add child relationships (tasks this task depends on)
		for _, childID := range req.ChildIDs {
			if err := qtx.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
				ParentID:       createdTask.ID,
				ChildID:        int64(childID),
				RelationTypeID: 1,
			}); err != nil {
				return fmt.Errorf("failed to add child relation: %w", err)
			}
		}

		// Add blocking relationships (tasks that block this task)
		for _, blockerID := range req.BlockedByIDs {
			// This task (Parent) is blocked by blockerID (Child)
			if err := qtx.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
				ParentID:       createdTask.ID,
				ChildID:        int64(blockerID),
				RelationTypeID: 2, // Blocking relationship
			}); err != nil {
				return fmt.Errorf("failed to add blocked-by relation: %w", err)
			}
		}

		// Add blocked relationships (tasks that are blocked by this task)
		for _, blockedID := range req.BlocksIDs {
			// blockedID (Parent) is blocked by this task (Child)
			if err := qtx.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
				ParentID:       int64(blockedID),
				ChildID:        createdTask.ID,
				RelationTypeID: 2, // Blocking relationship
			}); err != nil {
				return fmt.Errorf("failed to add blocks relation: %w", err)
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
	// Validate task ID
	if req.TaskID <= 0 {
		return ErrInvalidTaskID
	}

	// Validate fields if provided
	if req.Title != nil && *req.Title == "" {
		return ErrEmptyTitle
	}
	if req.Title != nil && len(*req.Title) > 255 {
		return ErrTitleTooLong
	}
	if req.PriorityID != nil && *req.PriorityID <= 0 {
		return ErrInvalidPriority
	}
	if req.TypeID != nil && *req.TypeID <= 0 {
		return ErrInvalidType
	}

	// Update basic fields if provided
	if req.Title != nil || req.Description != nil {
		title := ""
		var description sql.NullString

		if req.Title != nil {
			title = *req.Title
		} else {
			// Need to get existing title
			detail, err := s.queries.GetTaskDetail(ctx, int64(req.TaskID))
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}
			title = detail.Title
		}

		if req.Description != nil {
			description = sql.NullString{String: *req.Description, Valid: true}
		} else {
			// Need to get existing description
			detail, err := s.queries.GetTaskDetail(ctx, int64(req.TaskID))
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}
			description = detail.Description
		}

		if err := s.queries.UpdateTask(ctx, generated.UpdateTaskParams{
			Title:       title,
			Description: description,
			ID:          int64(req.TaskID),
		}); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}
	}

	// Update priority if provided
	if req.PriorityID != nil {
		if err := s.queries.UpdateTaskPriority(ctx, generated.UpdateTaskPriorityParams{
			PriorityID: int64(*req.PriorityID),
			ID:         int64(req.TaskID),
		}); err != nil {
			return fmt.Errorf("failed to update priority: %w", err)
		}
	}

	// Update type if provided
	if req.TypeID != nil {
		if err := s.queries.UpdateTaskType(ctx, generated.UpdateTaskTypeParams{
			TypeID: int64(*req.TypeID),
			ID:     int64(req.TaskID),
		}); err != nil {
			return fmt.Errorf("failed to update type: %w", err)
		}
	}

	// Publish event
	s.publishTaskEvent(ctx, req.TaskID)

	return nil
}

// DeleteTask handles task deletion
func (s *service) DeleteTask(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	if err := s.queries.DeleteTask(ctx, int64(taskID)); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Publish event
	s.publishTaskEvent(ctx, taskID)

	return nil
}

// GetTaskDetail retrieves full task details
func (s *service) GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error) {
	if taskID <= 0 {
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
		IsBlocked:   taskRow.IsBlocked > 0,
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

	return detail, nil
}

// GetTaskSummariesByProject retrieves task summaries for a project
// Uses sync.Pool to optimize memory allocations for frequently called batch operations
func (s *service) GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	rows, err := s.queries.GetTaskSummariesByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task summaries: %w", err)
	}

	// Get map from pool for reuse, reducing allocation overhead
	result := taskSummaryMapPool.Get().(map[int][]*models.TaskSummary)
	defer func() {
		// Clear map before returning to pool for reuse
		for k := range result {
			delete(result, k)
		}
		taskSummaryMapPool.Put(result)
	}()

	// Group by column
	for _, row := range rows {
		summary := converters.TaskSummaryFromRowToModel(row)
		columnID := int(row.ColumnID)
		result[columnID] = append(result[columnID], summary)
	}

	// Create a copy to return since the original goes back to the pool
	returnResult := make(map[int][]*models.TaskSummary, len(result))
	for columnID, summaries := range result {
		returnResult[columnID] = summaries
	}

	return returnResult, nil
}

// GetTaskSummariesByProjectFiltered retrieves filtered task summaries
func (s *service) GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error) {
	// Add wildcards for LIKE query
	searchPattern := "%" + searchQuery + "%"

	rows, err := s.queries.GetTaskSummariesByProjectFiltered(ctx, generated.GetTaskSummariesByProjectFilteredParams{
		ProjectID: int64(projectID),
		Title:     searchPattern,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered task summaries: %w", err)
	}

	// Group by column
	result := make(map[int][]*models.TaskSummary)
	for _, row := range rows {
		summary := converters.FilteredTaskSummaryFromRowToModel(row)
		columnID := int(row.ColumnID)
		result[columnID] = append(result[columnID], summary)
	}

	return result, nil
}

// GetReadyTaskSummariesByProject retrieves task summaries for tasks in ready columns (and not blocked)
func (s *service) GetReadyTaskSummariesByProject(ctx context.Context, projectID int) ([]*models.TaskSummary, error) {
	rows, err := s.queries.GetReadyTaskSummariesByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ready task summaries: %w", err)
	}

	result := make([]*models.TaskSummary, 0, len(rows))
	for _, row := range rows {
		// Only include unblocked tasks
		if row.IsBlocked == 0 {
			summary := converters.ReadyTaskSummaryFromRowToModel(row)
			result = append(result, summary)
		}
	}

	return result, nil
}

// GetTaskReferencesForProject retrieves task references for a project
func (s *service) GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error) {
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

// MoveTaskToNextColumn moves task to next column
func (s *service) MoveTaskToNextColumn(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Get current column
	posRow, err := s.queries.GetTaskPosition(ctx, int64(taskID))
	if err != nil {
		return fmt.Errorf("failed to get task position: %w", err)
	}

	// Get next column
	nextColumnID, err := s.queries.GetNextColumnID(ctx, posRow.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get next column: %w", err)
	}
	if nextColumnID == nil {
		return fmt.Errorf("no next column available")
	}

	// Convert interface{} to int64
	var nextColID int64
	switch v := nextColumnID.(type) {
	case int64:
		nextColID = v
	case nil:
		return fmt.Errorf("no next column available")
	default:
		return fmt.Errorf("unexpected column ID type")
	}

	// Get task count in target column to append at the end
	taskCount, err := s.queries.GetTaskCountByColumn(ctx, nextColID)
	if err != nil {
		return fmt.Errorf("failed to get task count: %w", err)
	}

	// Move task to next column
	if err := s.queries.MoveTaskToColumn(ctx, generated.MoveTaskToColumnParams{
		ColumnID: nextColID,
		Position: taskCount + 1,
		ID:       int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// MoveTaskToPrevColumn moves task to previous column
func (s *service) MoveTaskToPrevColumn(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Get current column
	posRow, err := s.queries.GetTaskPosition(ctx, int64(taskID))
	if err != nil {
		return fmt.Errorf("failed to get task position: %w", err)
	}

	// Get previous column
	prevColumnID, err := s.queries.GetPrevColumnID(ctx, posRow.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get previous column: %w", err)
	}
	if prevColumnID == nil {
		return fmt.Errorf("no previous column available")
	}

	// Convert interface{} to int64
	var prevColID int64
	switch v := prevColumnID.(type) {
	case int64:
		prevColID = v
	case nil:
		return fmt.Errorf("no previous column available")
	default:
		return fmt.Errorf("unexpected column ID type")
	}

	// Get task count in target column to append at the end
	taskCount, err := s.queries.GetTaskCountByColumn(ctx, prevColID)
	if err != nil {
		return fmt.Errorf("failed to get task count: %w", err)
	}

	// Move task to previous column
	if err := s.queries.MoveTaskToColumn(ctx, generated.MoveTaskToColumnParams{
		ColumnID: prevColID,
		Position: taskCount + 1,
		ID:       int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// MoveTaskToColumn moves task to specific column
func (s *service) MoveTaskToColumn(ctx context.Context, taskID, columnID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}
	if columnID <= 0 {
		return ErrInvalidColumnID
	}

	// Verify task exists before moving
	_, err := s.queries.GetTaskPosition(ctx, int64(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidTaskID
		}
		return fmt.Errorf("failed to verify task exists: %w", err)
	}

	// Get task count in target column to append at the end
	taskCount, err := s.queries.GetTaskCountByColumn(ctx, int64(columnID))
	if err != nil {
		return fmt.Errorf("failed to get task count: %w", err)
	}

	if err := s.queries.MoveTaskToColumn(ctx, generated.MoveTaskToColumnParams{
		ColumnID: int64(columnID),
		Position: taskCount + 1,
		ID:       int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// MoveTaskToReadyColumn moves task to the column marked as holding ready tasks
func (s *service) MoveTaskToReadyColumn(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Get task detail to find project
	taskDetail, err := s.queries.GetTaskDetail(ctx, int64(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidTaskID
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get the column via project ID
	column, err := s.queries.GetColumnByID(ctx, taskDetail.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get column: %w", err)
	}

	// Get ready column for project
	readyColumn, err := s.queries.GetReadyColumnByProject(ctx, column.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no ready column configured for this project")
		}
		return fmt.Errorf("failed to get ready column: %w", err)
	}

	// Check if already in ready column
	if taskDetail.ColumnID == readyColumn.ID {
		return ErrTaskAlreadyInTargetColumn
	}

	// Move task to ready column
	return s.MoveTaskToColumn(ctx, taskID, int(readyColumn.ID))
}

// MoveTaskToCompletedColumn moves task to the column marked as holding completed tasks
func (s *service) MoveTaskToCompletedColumn(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Get task detail to find project
	taskDetail, err := s.queries.GetTaskDetail(ctx, int64(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidTaskID
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get the column via project ID
	column, err := s.queries.GetColumnByID(ctx, taskDetail.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get column: %w", err)
	}

	// Get completed column for project
	completedColumn, err := s.queries.GetCompletedColumnByProject(ctx, column.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no completed column configured for this project")
		}
		return fmt.Errorf("failed to get completed column: %w", err)
	}

	// Check if already in completed column
	if taskDetail.ColumnID == completedColumn.ID {
		return ErrTaskAlreadyInTargetColumn
	}

	// Move task to completed column
	return s.MoveTaskToColumn(ctx, taskID, int(completedColumn.ID))
}

// MoveTaskToInProgressColumn moves a task to the column marked as holding in-progress tasks
func (s *service) MoveTaskToInProgressColumn(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Get task detail to find project
	taskDetail, err := s.queries.GetTaskDetail(ctx, int64(taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidTaskID
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get the column via project ID
	column, err := s.queries.GetColumnByID(ctx, taskDetail.ColumnID)
	if err != nil {
		return fmt.Errorf("failed to get column: %w", err)
	}

	// Get in-progress column for project
	inProgressColumn, err := s.queries.GetInProgressColumnByProject(ctx, column.ProjectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no in-progress column configured for this project")
		}
		return fmt.Errorf("failed to get in-progress column: %w", err)
	}

	// Check if already in in-progress column
	if taskDetail.ColumnID == inProgressColumn.ID {
		return ErrTaskAlreadyInTargetColumn
	}

	// Move task to in-progress column
	return s.MoveTaskToColumn(ctx, taskID, int(inProgressColumn.ID))
}

// GetInProgressTasksByProject retrieves all tasks in in-progress columns for a project
func (s *service) GetInProgressTasksByProject(ctx context.Context, projectID int) ([]*models.TaskDetail, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}

	rows, err := s.queries.GetInProgressTasksByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get in-progress tasks: %w", err)
	}

	tasks := make([]*models.TaskDetail, 0, len(rows))
	for _, row := range rows {
		// Get full task detail for each in-progress task
		taskDetail, err := s.GetTaskDetail(ctx, int(row.ID))
		if err != nil {
			slog.Error("failed to load task details for in-progress task",
				"task_id", row.ID,
				"error", err.Error(),
			)
			continue
		}
		tasks = append(tasks, taskDetail)
	}

	return tasks, nil
}

// MoveTaskUp moves task up in its column
func (s *service) MoveTaskUp(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Use WithTx helper to avoid UNIQUE constraint violations during swap
	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := generated.New(tx)

		// Get task position
		posRow, err := qtx.GetTaskPosition(ctx, int64(taskID))
		if err != nil {
			return fmt.Errorf("failed to get task position: %w", err)
		}

		// Get task above
		aboveRow, err := qtx.GetTaskAbove(ctx, generated.GetTaskAboveParams(posRow))
		if err != nil {
			return fmt.Errorf("no task above: %w", err)
		}

		// Swap positions using temporary negative position to avoid UNIQUE constraint violation
		// Step 1: Move current task to temporary position
		if err := qtx.SetTaskPosition(ctx, generated.SetTaskPositionParams{
			Position: -1,
			ID:       int64(taskID),
		}); err != nil {
			return fmt.Errorf("failed to set temporary position: %w", err)
		}

		// Step 2: Move task above to current task's position
		if err := qtx.SetTaskPosition(ctx, generated.SetTaskPositionParams{
			Position: posRow.Position,
			ID:       aboveRow.ID,
		}); err != nil {
			return fmt.Errorf("failed to move other task down: %w", err)
		}

		// Step 3: Move current task to above position
		if err := qtx.SetTaskPosition(ctx, generated.SetTaskPositionParams{
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
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Use WithTx helper to avoid UNIQUE constraint violations during swap
	err := database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		qtx := generated.New(tx)

		// Get task position
		posRow, err := qtx.GetTaskPosition(ctx, int64(taskID))
		if err != nil {
			return fmt.Errorf("failed to get task position: %w", err)
		}

		// Get task below
		belowRow, err := qtx.GetTaskBelow(ctx, generated.GetTaskBelowParams(posRow))
		if err != nil {
			return fmt.Errorf("no task below: %w", err)
		}

		// Swap positions using temporary negative position to avoid UNIQUE constraint violation
		// Step 1: Move current task to temporary position
		if err := qtx.SetTaskPosition(ctx, generated.SetTaskPositionParams{
			Position: -1,
			ID:       int64(taskID),
		}); err != nil {
			return fmt.Errorf("failed to set temporary position: %w", err)
		}

		// Step 2: Move task below to current task's position
		if err := qtx.SetTaskPosition(ctx, generated.SetTaskPositionParams{
			Position: posRow.Position,
			ID:       belowRow.ID,
		}); err != nil {
			return fmt.Errorf("failed to move other task up: %w", err)
		}

		// Step 3: Move current task to below position
		if err := qtx.SetTaskPosition(ctx, generated.SetTaskPositionParams{
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

// AddParentRelation adds a parent relationship (parent depends on this task)
func (s *service) AddParentRelation(ctx context.Context, taskID, parentID int, relationTypeID int) error {
	if taskID <= 0 || parentID <= 0 {
		return ErrInvalidTaskID
	}
	if taskID == parentID {
		return ErrSelfRelation
	}

	// Add the relationship (this task is the child)
	if err := s.queries.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
		ParentID:       int64(parentID),
		ChildID:        int64(taskID),
		RelationTypeID: int64(relationTypeID),
	}); err != nil {
		return fmt.Errorf("failed to add parent relation: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// AddChildRelation adds a child relationship (this task depends on child)
func (s *service) AddChildRelation(ctx context.Context, taskID, childID int, relationTypeID int) error {
	if taskID <= 0 || childID <= 0 {
		return ErrInvalidTaskID
	}
	if taskID == childID {
		return ErrSelfRelation
	}

	// Add the relationship (this task is the parent)
	if err := s.queries.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
		ParentID:       int64(taskID),
		ChildID:        int64(childID),
		RelationTypeID: int64(relationTypeID),
	}); err != nil {
		return fmt.Errorf("failed to add child relation: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// RemoveParentRelation removes a parent relationship
func (s *service) RemoveParentRelation(ctx context.Context, taskID, parentID int) error {
	if taskID <= 0 || parentID <= 0 {
		return ErrInvalidTaskID
	}

	if err := s.queries.RemoveSubtask(ctx, generated.RemoveSubtaskParams{
		ParentID: int64(parentID),
		ChildID:  int64(taskID),
	}); err != nil {
		return fmt.Errorf("failed to remove parent relation: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// RemoveChildRelation removes a child relationship
func (s *service) RemoveChildRelation(ctx context.Context, taskID, childID int) error {
	if taskID <= 0 || childID <= 0 {
		return ErrInvalidTaskID
	}

	if err := s.queries.RemoveSubtask(ctx, generated.RemoveSubtaskParams{
		ParentID: int64(taskID),
		ChildID:  int64(childID),
	}); err != nil {
		return fmt.Errorf("failed to remove child relation: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// AttachLabel attaches a label to a task
func (s *service) AttachLabel(ctx context.Context, taskID, labelID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}
	if labelID <= 0 {
		return ErrInvalidLabelID
	}

	if err := s.queries.AddLabelToTask(ctx, generated.AddLabelToTaskParams{
		TaskID:  int64(taskID),
		LabelID: int64(labelID),
	}); err != nil {
		return fmt.Errorf("failed to attach label: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// DetachLabel detaches a label from a task
func (s *service) DetachLabel(ctx context.Context, taskID, labelID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}
	if labelID <= 0 {
		return ErrInvalidLabelID
	}

	if err := s.queries.RemoveLabelFromTask(ctx, generated.RemoveLabelFromTaskParams{
		TaskID:  int64(taskID),
		LabelID: int64(labelID),
	}); err != nil {
		return fmt.Errorf("failed to detach label: %w", err)
	}

	s.publishTaskEvent(ctx, taskID)
	return nil
}

// ============================================================================
// COMMENT OPERATIONS
// ============================================================================

// CreateComment creates a new comment on a task
func (s *service) CreateComment(ctx context.Context, req CreateCommentRequest) (*models.Comment, error) {
	// Validate task ID
	if req.TaskID <= 0 {
		return nil, ErrInvalidTaskID
	}

	// Validate message
	if err := validateCommentMessage(req.Message); err != nil {
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

	// Create comment
	comment, err := s.queries.CreateComment(ctx, generated.CreateCommentParams{
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
	// Validate comment ID
	if req.CommentID <= 0 {
		return ErrInvalidCommentID
	}

	// Validate message
	if err := validateCommentMessage(req.Message); err != nil {
		return err
	}

	// Verify comment exists and get task ID for event publishing
	comment, err := s.queries.GetComment(ctx, int64(req.CommentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Update comment
	if err := s.queries.UpdateComment(ctx, generated.UpdateCommentParams{
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
	// Validate comment ID
	if commentID <= 0 {
		return ErrInvalidCommentID
	}

	// Get task ID before deletion for event publishing
	comment, err := s.queries.GetComment(ctx, int64(commentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCommentNotFound
		}
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Delete comment
	if err := s.queries.DeleteComment(ctx, int64(commentID)); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	s.publishTaskEvent(ctx, int(comment.TaskID))
	return nil
}

// GetCommentsByTask retrieves all comments for a task
func (s *service) GetCommentsByTask(ctx context.Context, taskID int) ([]*models.Comment, error) {
	if taskID <= 0 {
		return nil, ErrInvalidTaskID
	}

	rows, err := s.queries.GetCommentsByTask(ctx, int64(taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return converters.CommentsToModels(rows), nil
}

// validateCreateTask validates a CreateTaskRequest
func (s *service) validateCreateTask(req CreateTaskRequest) error {
	if req.Title == "" {
		return ErrEmptyTitle
	}
	if len(req.Title) > 255 {
		return ErrTitleTooLong
	}
	if req.ColumnID <= 0 {
		return ErrInvalidColumnID
	}
	if req.Position < 0 {
		return ErrInvalidPosition
	}
	if req.PriorityID < 0 {
		return ErrInvalidPriority
	}
	if req.TypeID < 0 {
		return ErrInvalidType
	}
	return nil
}

// validateCommentMessage validates a comment message
func validateCommentMessage(message string) error {
	if message == "" {
		return ErrEmptyCommentMessage
	}
	if len(message) > 1000 {
		return ErrCommentMessageTooLong
	}
	return nil
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

	// Publish with retry (3 attempts with exponential backoff)
	// Non-blocking: errors are logged but don't affect the operation
	_ = events.PublishWithRetry(s.eventClient, events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: int(projectID),
	}, 3)
}
