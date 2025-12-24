package task

import (
	"context"
	"fmt"

	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// Service defines all task-related business operations
type Service interface {
	// Read operations
	GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error)
	GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error)
	GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error)
	GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error)

	// Write operations
	CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error)
	UpdateTask(ctx context.Context, req UpdateTaskRequest) error
	DeleteTask(ctx context.Context, taskID int) error

	// Task movements
	MoveTaskToNextColumn(ctx context.Context, taskID int) error
	MoveTaskToPrevColumn(ctx context.Context, taskID int) error
	MoveTaskToColumn(ctx context.Context, taskID, columnID int) error
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

// service implements Service interface
type service struct {
	repo        database.DataStore
	eventClient events.EventPublisher
}

// NewService creates a new task service
func NewService(repo database.DataStore, eventClient events.EventPublisher) Service {
	return &service{
		repo:        repo,
		eventClient: eventClient,
	}
}

// CreateTask handles task creation with validation and business rules
func (s *service) CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error) {
	// Validate request
	if err := s.validateCreateTask(req); err != nil {
		return nil, err
	}

	// Create task in repository
	task, err := s.repo.CreateTask(ctx, req.Title, req.Description, req.ColumnID, req.Position)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Set priority if provided (default is handled by database)
	if req.PriorityID > 0 {
		if err := s.repo.UpdateTaskPriority(ctx, task.ID, req.PriorityID); err != nil {
			return nil, fmt.Errorf("failed to set priority: %w", err)
		}
	}

	// Set type if provided (default is handled by database)
	if req.TypeID > 0 {
		if err := s.repo.UpdateTaskType(ctx, task.ID, req.TypeID); err != nil {
			return nil, fmt.Errorf("failed to set type: %w", err)
		}
	}

	// Attach labels
	for _, labelID := range req.LabelIDs {
		if err := s.repo.AddLabelToTask(ctx, task.ID, labelID); err != nil {
			return nil, fmt.Errorf("failed to attach label %d: %w", labelID, err)
		}
	}

	// Add parent relationships (tasks that depend on this task)
	for _, parentID := range req.ParentIDs {
		if err := s.AddParentRelation(ctx, task.ID, parentID, 1); err != nil {
			return nil, fmt.Errorf("failed to add parent relation: %w", err)
		}
	}

	// Add child relationships (tasks this task depends on)
	for _, childID := range req.ChildIDs {
		if err := s.AddChildRelation(ctx, task.ID, childID, 1); err != nil {
			return nil, fmt.Errorf("failed to add child relation: %w", err)
		}
	}

	// Publish event (if event client exists)
	s.publishTaskEvent(task.ID)

	return task, nil
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
		description := ""

		if req.Title != nil {
			title = *req.Title
		} else {
			// Need to get existing title
			detail, err := s.repo.GetTaskDetail(ctx, req.TaskID)
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}
			title = detail.Title
		}

		if req.Description != nil {
			description = *req.Description
		} else {
			// Need to get existing description
			detail, err := s.repo.GetTaskDetail(ctx, req.TaskID)
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}
			description = detail.Description
		}

		if err := s.repo.UpdateTask(ctx, req.TaskID, title, description); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}
	}

	// Update priority if provided
	if req.PriorityID != nil {
		if err := s.repo.UpdateTaskPriority(ctx, req.TaskID, *req.PriorityID); err != nil {
			return fmt.Errorf("failed to update priority: %w", err)
		}
	}

	// Update type if provided
	if req.TypeID != nil {
		if err := s.repo.UpdateTaskType(ctx, req.TaskID, *req.TypeID); err != nil {
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

	if err := s.repo.DeleteTask(ctx, taskID); err != nil {
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

	return s.repo.GetTaskDetail(ctx, taskID)
}

// GetTaskSummariesByProject retrieves task summaries for a project
func (s *service) GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	return s.repo.GetTaskSummariesByProject(ctx, projectID)
}

// GetTaskSummariesByProjectFiltered retrieves filtered task summaries
func (s *service) GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error) {
	return s.repo.GetTaskSummariesByProjectFiltered(ctx, projectID, searchQuery)
}

// GetTaskReferencesForProject retrieves task references for a project
func (s *service) GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error) {
	return s.repo.GetTaskReferencesForProject(ctx, projectID)
}

// MoveTaskToNextColumn moves task to next column
func (s *service) MoveTaskToNextColumn(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	if err := s.repo.MoveTaskToNextColumn(ctx, taskID); err != nil {
		return err
	}

	s.publishTaskEvent(taskID)
	return nil
}

// MoveTaskToPrevColumn moves task to previous column
func (s *service) MoveTaskToPrevColumn(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	if err := s.repo.MoveTaskToPrevColumn(ctx, taskID); err != nil {
		return err
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

	if err := s.repo.MoveTaskToColumn(ctx, taskID, columnID); err != nil {
		return err
	}

	s.publishTaskEvent(taskID)
	return nil
}

// MoveTaskUp moves task up in its column
func (s *service) MoveTaskUp(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	if err := s.repo.SwapTaskUp(ctx, taskID); err != nil {
		return err
	}

	s.publishTaskEvent(taskID)
	return nil
}

// MoveTaskDown moves task down in its column
func (s *service) MoveTaskDown(ctx context.Context, taskID int) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	if err := s.repo.SwapTaskDown(ctx, taskID); err != nil {
		return err
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
	if err := s.repo.AddSubtaskWithRelationType(ctx, parentID, taskID, relationTypeID); err != nil {
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
	if err := s.repo.AddSubtaskWithRelationType(ctx, taskID, childID, relationTypeID); err != nil {
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

	if err := s.repo.RemoveSubtask(ctx, parentID, taskID); err != nil {
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

	if err := s.repo.RemoveSubtask(ctx, taskID, childID); err != nil {
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

	if err := s.repo.AddLabelToTask(ctx, taskID, labelID); err != nil {
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

	if err := s.repo.RemoveLabelFromTask(ctx, taskID, labelID); err != nil {
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

// publishTaskEvent publishes a task event if event client exists
func (s *service) publishTaskEvent(taskID int) {
	if s.eventClient == nil {
		return
	}

	// Publish database changed event
	// The daemon will notify all connected clients to refresh
	// We don't include project ID for now - clients will refresh on any DB change
	_ = taskID // Used for future enhancement when we track project-specific changes
}
