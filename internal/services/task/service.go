package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/thenoetrevino/paso/internal/database/generated"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// Service defines all task-related business operations
type Service interface {
	// Read operations
	GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error)
	GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error)
	GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error)
	GetReadyTaskSummariesByProject(ctx context.Context, projectID int) ([]*models.TaskSummary, error)
	GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error)
	GetTaskTreeByProject(ctx context.Context, projectID int) ([]*models.TaskTreeNode, error)

	// Write operations
	CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error)
	UpdateTask(ctx context.Context, req UpdateTaskRequest) error
	DeleteTask(ctx context.Context, taskID int) error

	// Task movements
	MoveTaskToNextColumn(ctx context.Context, taskID int) error
	MoveTaskToPrevColumn(ctx context.Context, taskID int) error
	MoveTaskToColumn(ctx context.Context, taskID, columnID int) error
	MoveTaskToReadyColumn(ctx context.Context, taskID int) error
	MoveTaskToCompletedColumn(ctx context.Context, taskID int) error
	MoveTaskUp(ctx context.Context, taskID int) error
	MoveTaskDown(ctx context.Context, taskID int) error

	// Task relationships
	AddParentRelation(ctx context.Context, taskID, parentID int, relationTypeID int) error
	AddChildRelation(ctx context.Context, taskID, childID int, relationTypeID int) error
	RemoveParentRelation(ctx context.Context, taskID, parentID int) error
	RemoveChildRelation(ctx context.Context, taskID, childID int) error

	// Label management
	AttachLabel(ctx context.Context, taskID, labelID int) error
	DetachLabel(ctx context.Context, taskID, labelID int) error
}

// CreateTaskRequest encapsulates all data needed to create a task
type CreateTaskRequest struct {
	Title       string
	Description string
	ColumnID    int
	Position    int
	PriorityID  int // Optional: 0 means use default
	TypeID      int // Optional: 0 means use default
	LabelIDs    []int
	ParentIDs   []int // Parent task IDs (tasks that depend on this task)
	ChildIDs    []int // Child task IDs (tasks this task depends on)
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

	// Get next ticket number
	ticketNumber, err := qtx.GetNextTicketNumber(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket number: %w", err)
	}

	// Create task
	var desc sql.NullString
	if req.Description != "" {
		desc = sql.NullString{String: req.Description, Valid: true}
	}

	createdTask, err := qtx.CreateTask(ctx, generated.CreateTaskParams{
		Title:        req.Title,
		Description:  desc,
		ColumnID:     int64(req.ColumnID),
		Position:     int64(req.Position),
		TicketNumber: ticketNumber,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Increment ticket number
	if err := qtx.IncrementTicketNumber(ctx, projectID); err != nil {
		return nil, fmt.Errorf("failed to increment ticket number: %w", err)
	}

	// Set priority if provided (default is handled by database)
	if req.PriorityID > 0 {
		if err := qtx.UpdateTaskPriority(ctx, generated.UpdateTaskPriorityParams{
			PriorityID: int64(req.PriorityID),
			ID:         createdTask.ID,
		}); err != nil {
			return nil, fmt.Errorf("failed to set priority: %w", err)
		}
	}

	// Set type if provided (default is handled by database)
	if req.TypeID > 0 {
		if err := qtx.UpdateTaskType(ctx, generated.UpdateTaskTypeParams{
			TypeID: int64(req.TypeID),
			ID:     createdTask.ID,
		}); err != nil {
			return nil, fmt.Errorf("failed to set type: %w", err)
		}
	}

	// Attach labels
	for _, labelID := range req.LabelIDs {
		if err := qtx.AddLabelToTask(ctx, generated.AddLabelToTaskParams{
			TaskID:  createdTask.ID,
			LabelID: int64(labelID),
		}); err != nil {
			return nil, fmt.Errorf("failed to attach label %d: %w", labelID, err)
		}
	}

	// Add parent relationships (tasks that depend on this task)
	for _, parentID := range req.ParentIDs {
		if err := qtx.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
			ParentID:       int64(parentID),
			ChildID:        createdTask.ID,
			RelationTypeID: 1,
		}); err != nil {
			return nil, fmt.Errorf("failed to add parent relation: %w", err)
		}
	}

	// Add child relationships (tasks this task depends on)
	for _, childID := range req.ChildIDs {
		if err := qtx.AddSubtaskWithRelationType(ctx, generated.AddSubtaskWithRelationTypeParams{
			ParentID:       createdTask.ID,
			ChildID:        int64(childID),
			RelationTypeID: 1,
		}); err != nil {
			return nil, fmt.Errorf("failed to add child relation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish event after successful commit
	s.publishTaskEvent(int(createdTask.ID))

	// Convert to model
	return convertToTaskModel(createdTask), nil
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
	s.publishTaskEvent(req.TaskID)

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
	s.publishTaskEvent(taskID)

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

	// Convert to model
	detail := &models.TaskDetail{
		ID:          int(taskRow.ID),
		Title:       taskRow.Title,
		Description: taskRow.Description.String,
		ColumnID:    int(taskRow.ColumnID),
		ColumnName:  taskRow.ColumnName,
		ProjectName: taskRow.ProjectName,
		Position:    int(taskRow.Position),
		Labels:      convertLabelsToModels(labels),
		ParentTasks: convertParentTasksToReferences(parentRows),
		ChildTasks:  convertChildTasksToReferences(childRows),
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
func (s *service) GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	rows, err := s.queries.GetTaskSummariesByProject(ctx, int64(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task summaries: %w", err)
	}

	// Group by column
	result := make(map[int][]*models.TaskSummary)
	for _, row := range rows {
		summary := convertTaskSummaryRowToModel(row)
		columnID := int(row.ColumnID)
		result[columnID] = append(result[columnID], summary)
	}

	return result, nil
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
		summary := convertFilteredTaskSummaryRowToModel(row)
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
			summary := convertReadyTaskSummaryRowToModel(row)
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

	s.publishTaskEvent(taskID)
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

	s.publishTaskEvent(taskID)
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

	s.publishTaskEvent(taskID)
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

// MoveTaskUp moves task up in its column
func (s *service) MoveTaskUp(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Start transaction to avoid UNIQUE constraint violations during swap
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.publishTaskEvent(taskID)
	return nil
}

// MoveTaskDown moves task down in its column
func (s *service) MoveTaskDown(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	// Start transaction to avoid UNIQUE constraint violations during swap
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.publishTaskEvent(taskID)
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

	s.publishTaskEvent(taskID)
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

	s.publishTaskEvent(taskID)
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

	s.publishTaskEvent(taskID)
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

	s.publishTaskEvent(taskID)
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

	s.publishTaskEvent(taskID)
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

	s.publishTaskEvent(taskID)
	return nil
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

// publishTaskEvent publishes a task event
func (s *service) publishTaskEvent(taskID int) {
	if s.eventClient == nil {
		return
	}

	// Get project ID for the task
	projectID, err := s.queries.GetProjectIDFromTask(context.Background(), int64(taskID))
	if err != nil {
		log.Printf("failed to get project ID for task %d: %v", taskID, err)
		return
	}

	if err := s.eventClient.SendEvent(events.Event{
		Type:      events.EventDatabaseChanged,
		ProjectID: int(projectID),
	}); err != nil {
		log.Printf("failed to send event for task %d: %v", taskID, err)
	}
}

// ============================================================================
// MODEL CONVERSION FUNCTIONS
// ============================================================================

// convertToTaskModel converts a generated.Task to models.Task
func convertToTaskModel(t generated.Task) *models.Task {
	task := &models.Task{
		ID:         int(t.ID),
		Title:      t.Title,
		ColumnID:   int(t.ColumnID),
		Position:   int(t.Position),
		TypeID:     int(t.TypeID),
		PriorityID: int(t.PriorityID),
	}

	if t.Description.Valid {
		task.Description = t.Description.String
	}
	if t.CreatedAt.Valid {
		task.CreatedAt = t.CreatedAt.Time
	}
	if t.UpdatedAt.Valid {
		task.UpdatedAt = t.UpdatedAt.Time
	}

	return task
}

// convertLabelsToModels converts generated.Label slice to models.Label slice
func convertLabelsToModels(labels []generated.Label) []*models.Label {
	result := make([]*models.Label, 0, len(labels))
	for _, l := range labels {
		result = append(result, &models.Label{
			ID:        int(l.ID),
			Name:      l.Name,
			Color:     l.Color,
			ProjectID: int(l.ProjectID),
		})
	}
	return result
}

// convertParentTasksToReferences converts parent task rows to TaskReference slice
func convertParentTasksToReferences(rows []generated.GetParentTasksRow) []*models.TaskReference {
	result := make([]*models.TaskReference, 0, len(rows))
	for _, row := range rows {
		ref := &models.TaskReference{
			ID:             int(row.ID),
			Title:          row.Title,
			ProjectName:    row.Name,
			RelationTypeID: int(row.ID_2),
			RelationLabel:  row.PToCLabel,
			RelationColor:  row.Color,
			IsBlocking:     row.IsBlocking,
		}
		if row.TicketNumber.Valid {
			ref.TicketNumber = int(row.TicketNumber.Int64)
		}
		result = append(result, ref)
	}
	return result
}

// convertChildTasksToReferences converts child task rows to TaskReference slice
func convertChildTasksToReferences(rows []generated.GetChildTasksRow) []*models.TaskReference {
	result := make([]*models.TaskReference, 0, len(rows))
	for _, row := range rows {
		ref := &models.TaskReference{
			ID:             int(row.ID),
			Title:          row.Title,
			ProjectName:    row.Name,
			RelationTypeID: int(row.ID_2),
			RelationLabel:  row.CToPLabel,
			RelationColor:  row.Color,
			IsBlocking:     row.IsBlocking,
		}
		if row.TicketNumber.Valid {
			ref.TicketNumber = int(row.TicketNumber.Int64)
		}
		result = append(result, ref)
	}
	return result
}

// convertTaskSummaryRowToModel converts a task summary row to models.TaskSummary
func convertTaskSummaryRowToModel(row generated.GetTaskSummariesByProjectRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked > 0,
		Labels:    parseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors),
	}

	if row.TypeDescription.Valid {
		summary.TypeDescription = row.TypeDescription.String
	}
	if row.PriorityDescription.Valid {
		summary.PriorityDescription = row.PriorityDescription.String
	}
	if row.PriorityColor.Valid {
		summary.PriorityColor = row.PriorityColor.String
	}

	return summary
}

// convertReadyTaskSummaryRowToModel converts a ready task summary row to models.TaskSummary
func convertReadyTaskSummaryRowToModel(row generated.GetReadyTaskSummariesByProjectRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked > 0,
		Labels:    parseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors),
	}

	if row.TypeDescription.Valid {
		summary.TypeDescription = row.TypeDescription.String
	}
	if row.PriorityDescription.Valid {
		summary.PriorityDescription = row.PriorityDescription.String
	}
	if row.PriorityColor.Valid {
		summary.PriorityColor = row.PriorityColor.String
	}

	return summary
}

// convertFilteredTaskSummaryRowToModel converts a filtered task summary row to models.TaskSummary
func convertFilteredTaskSummaryRowToModel(row generated.GetTaskSummariesByProjectFilteredRow) *models.TaskSummary {
	summary := &models.TaskSummary{
		ID:        int(row.ID),
		Title:     row.Title,
		ColumnID:  int(row.ColumnID),
		Position:  int(row.Position),
		IsBlocked: row.IsBlocked > 0,
		Labels:    parseLabelsFromConcatenated(row.LabelIds, row.LabelNames, row.LabelColors),
	}

	if row.TypeDescription.Valid {
		summary.TypeDescription = row.TypeDescription.String
	}
	if row.PriorityDescription.Valid {
		summary.PriorityDescription = row.PriorityDescription.String
	}
	if row.PriorityColor.Valid {
		summary.PriorityColor = row.PriorityColor.String
	}

	return summary
}

// parseLabelsFromConcatenated parses GROUP_CONCAT label data into Label slice
func parseLabelsFromConcatenated(ids, names, colors string) []*models.Label {
	if ids == "" || names == "" || colors == "" {
		return []*models.Label{}
	}

	idParts := strings.Split(ids, string(rune(31))) // CHAR(31) separator
	nameParts := strings.Split(names, string(rune(31)))
	colorParts := strings.Split(colors, string(rune(31)))

	if len(idParts) != len(nameParts) || len(idParts) != len(colorParts) {
		return []*models.Label{}
	}

	labels := make([]*models.Label, 0, len(idParts))
	for i := range idParts {
		// Parse ID
		var id int
		_, _ = fmt.Sscanf(idParts[i], "%d", &id)

		labels = append(labels, &models.Label{
			ID:    id,
			Name:  nameParts[i],
			Color: colorParts[i],
		})
	}

	return labels
}
