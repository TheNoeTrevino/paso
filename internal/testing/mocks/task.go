package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/services/task"
)

// Compile-time interface verification
var _ task.Service = (*MockTaskService)(nil)

// MockTaskService is a mock implementation of task.Service for testing.
// It records all method calls and allows injection of return values and errors.
type MockTaskService struct {
	mu    sync.Mutex
	Calls []MockCall

	// TaskReader error injection
	GetTaskDetailErr                  error
	GetTaskActivitiesErr              error
	GetTaskSummariesByProjectErr      error
	GetTaskSummariesWithFiltersErr    error
	GetReadyTaskSummariesByProjectErr error
	GetInProgressTasksByProjectErr    error
	GetTaskReferencesForProjectErr    error
	GetTaskTreeByProjectErr           error
	GetTaskTypeAndPriorityIDsErr      error

	// TaskReader function callbacks
	GetTaskDetailFunc func(ctx context.Context, taskID int) (*models.TaskDetail, error)

	// TaskReader result injection
	GetTaskDetailResult                  *models.TaskDetail
	GetTaskActivitiesResult              []models.ActivityItem
	GetTaskSummariesByProjectResult      map[int][]*models.TaskSummary
	GetTaskSummariesWithFiltersResult    map[int][]*models.TaskSummary
	GetReadyTaskSummariesByProjectResult []*models.TaskSummary
	GetInProgressTasksByProjectResult    []*models.TaskDetail
	GetTaskReferencesForProjectResult    []*models.TaskReference
	GetTaskTreeByProjectResult           []*models.TaskTreeNode
	GetTaskTypeAndPriorityIDsTypeID      int
	GetTaskTypeAndPriorityIDsPriorityID  int

	// TaskWriter error injection
	CreateTaskErr         error
	UpdateTaskErr         error
	UpdateTaskAssigneeErr error
	UpdateTaskEstimateErr error
	UpdateTaskDueDateErr  error
	DeleteTaskErr         error
	ArchiveTaskErr        error

	// TaskWriter result injection
	CreateTaskResult *models.Task

	// TaskMover error injection
	MoveTaskToNextColumnErr       error
	MoveTaskToPrevColumnErr       error
	MoveTaskToColumnErr           error
	MoveTaskToReadyColumnErr      error
	MoveTaskToCompletedColumnErr  error
	MoveTaskToInProgressColumnErr error
	MoveTaskToProjectErr          error
	MoveTaskUpErr                 error
	MoveTaskDownErr               error

	// TaskRelationer error injection
	AddParentRelationErr    error
	AddChildRelationErr     error
	RemoveParentRelationErr error
	RemoveChildRelationErr  error

	// TaskLabeler error injection
	AttachLabelErr error
	DetachLabelErr error

	// TaskCommenter error injection
	CreateCommentErr     error
	UpdateCommentErr     error
	DeleteCommentErr     error
	GetCommentsByTaskErr error

	// TaskCommenter result injection
	CreateCommentResult     *models.Comment
	GetCommentsByTaskResult []*models.Comment
}

// NewMockTaskService creates a new mock task service.
func NewMockTaskService() *MockTaskService {
	return &MockTaskService{
		Calls: make([]MockCall, 0),
	}
}

func (m *MockTaskService) recordCall(method string, taskID int, args map[string]any) {
	if args == nil {
		args = make(map[string]any)
	}
	args["taskID"] = taskID // Keep for backward compatibility
	m.Calls = append(m.Calls, MockCall{
		Method: method,
		TaskID: taskID, // Also set field for ergonomic access
		Args:   args,
	})
}

// Reset clears all recorded calls and function callbacks.
func (m *MockTaskService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = make([]MockCall, 0)
	m.GetTaskDetailFunc = nil
}

// GetCalls returns a copy of all recorded calls.
func (m *MockTaskService) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockCall, len(m.Calls))
	copy(result, m.Calls)
	return result
}

// HasCall checks if a method was called with the given taskID.
func (m *MockTaskService) HasCall(method string, taskID int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.Calls {
		if call.Method == method {
			if id, ok := call.Args["taskID"].(int); ok && id == taskID {
				return true
			}
		}
	}
	return false
}

// CallCount returns the number of times a method was called.
func (m *MockTaskService) CallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// TaskReader methods

func (m *MockTaskService) GetTaskDetail(ctx context.Context, taskID int) (*models.TaskDetail, error) {
	m.mu.Lock()
	m.recordCall("GetTaskDetail", taskID, nil)
	fn := m.GetTaskDetailFunc
	result := m.GetTaskDetailResult
	err := m.GetTaskDetailErr
	m.mu.Unlock()

	if fn != nil {
		return fn(ctx, taskID)
	}
	return result, err
}

func (m *MockTaskService) GetTaskActivities(_ context.Context, taskID int) ([]models.ActivityItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetTaskActivities", taskID, nil)
	return m.GetTaskActivitiesResult, m.GetTaskActivitiesErr
}

func (m *MockTaskService) GetTaskSummariesByProject(_ context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetTaskSummariesByProject", projectID, nil)
	return m.GetTaskSummariesByProjectResult, m.GetTaskSummariesByProjectErr
}

func (m *MockTaskService) GetTaskSummariesWithFilters(_ context.Context, params task.TaskFilterParams) (map[int][]*models.TaskSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetTaskSummariesWithFilters", params.ProjectID, map[string]any{
		"params": params,
	})
	return m.GetTaskSummariesWithFiltersResult, m.GetTaskSummariesWithFiltersErr
}

func (m *MockTaskService) GetReadyTaskSummariesByProject(_ context.Context, projectID int) ([]*models.TaskSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetReadyTaskSummariesByProject", projectID, nil)
	return m.GetReadyTaskSummariesByProjectResult, m.GetReadyTaskSummariesByProjectErr
}

func (m *MockTaskService) GetInProgressTasksByProject(_ context.Context, projectID int) ([]*models.TaskDetail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetInProgressTasksByProject", projectID, nil)
	return m.GetInProgressTasksByProjectResult, m.GetInProgressTasksByProjectErr
}

func (m *MockTaskService) GetTaskReferencesForProject(_ context.Context, projectID int) ([]*models.TaskReference, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetTaskReferencesForProject", projectID, nil)
	return m.GetTaskReferencesForProjectResult, m.GetTaskReferencesForProjectErr
}

func (m *MockTaskService) GetTaskTreeByProject(_ context.Context, projectID int) ([]*models.TaskTreeNode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetTaskTreeByProject", projectID, nil)
	return m.GetTaskTreeByProjectResult, m.GetTaskTreeByProjectErr
}

func (m *MockTaskService) GetTaskTypeAndPriorityIDs(_ context.Context, taskID int) (typeID, priorityID int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetTaskTypeAndPriorityIDs", taskID, nil)
	return m.GetTaskTypeAndPriorityIDsTypeID, m.GetTaskTypeAndPriorityIDsPriorityID, m.GetTaskTypeAndPriorityIDsErr
}

// TaskWriter methods

func (m *MockTaskService) CreateTask(_ context.Context, req task.CreateTaskRequest) (*models.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateTask", 0, map[string]any{
		"title":       req.Title,
		"description": req.Description,
		"columnID":    req.ColumnID,
		"position":    req.Position,
		"priorityID":  req.PriorityID,
		"typeID":      req.TypeID,
		"assigneeID":  req.AssigneeID,
	})
	return m.CreateTaskResult, m.CreateTaskErr
}

func (m *MockTaskService) UpdateTask(_ context.Context, req task.UpdateTaskRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("UpdateTask", req.TaskID, map[string]any{
		"title":       req.Title,
		"description": req.Description,
		"priorityID":  req.PriorityID,
		"typeID":      req.TypeID,
	})
	return m.UpdateTaskErr
}

func (m *MockTaskService) UpdateTaskAssignee(_ context.Context, taskID int, assigneeID *int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("UpdateTaskAssignee", taskID, map[string]any{
		"assigneeID": assigneeID,
	})
	return m.UpdateTaskAssigneeErr
}

func (m *MockTaskService) UpdateTaskEstimate(_ context.Context, taskID int, estimate *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("UpdateTaskEstimate", taskID, map[string]any{
		"estimate": estimate,
	})
	return m.UpdateTaskEstimateErr
}

func (m *MockTaskService) UpdateTaskDueDate(_ context.Context, taskID int, dueDate *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("UpdateTaskDueDate", taskID, map[string]any{
		"dueDate": dueDate,
	})
	return m.UpdateTaskDueDateErr
}

func (m *MockTaskService) DeleteTask(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DeleteTask", taskID, nil)
	return m.DeleteTaskErr
}

func (m *MockTaskService) ArchiveTask(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("ArchiveTask", taskID, nil)
	return m.ArchiveTaskErr
}

// TaskMover methods

func (m *MockTaskService) MoveTaskToNextColumn(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskToNextColumn", taskID, nil)
	return m.MoveTaskToNextColumnErr
}

func (m *MockTaskService) MoveTaskToPrevColumn(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskToPrevColumn", taskID, nil)
	return m.MoveTaskToPrevColumnErr
}

func (m *MockTaskService) MoveTaskToColumn(_ context.Context, taskID, columnID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskToColumn", taskID, map[string]any{
		"columnID": columnID,
	})
	return m.MoveTaskToColumnErr
}

func (m *MockTaskService) MoveTaskToReadyColumn(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskToReadyColumn", taskID, nil)
	return m.MoveTaskToReadyColumnErr
}

func (m *MockTaskService) MoveTaskToCompletedColumn(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskToCompletedColumn", taskID, nil)
	return m.MoveTaskToCompletedColumnErr
}

func (m *MockTaskService) MoveTaskToInProgressColumn(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskToInProgressColumn", taskID, nil)
	return m.MoveTaskToInProgressColumnErr
}

func (m *MockTaskService) MoveTaskToProject(_ context.Context, taskID int, targetProjectID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskToProject", taskID, map[string]any{
		"targetProjectID": targetProjectID,
	})
	return m.MoveTaskToProjectErr
}

func (m *MockTaskService) MoveTaskUp(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskUp", taskID, nil)
	return m.MoveTaskUpErr
}

func (m *MockTaskService) MoveTaskDown(_ context.Context, taskID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("MoveTaskDown", taskID, nil)
	return m.MoveTaskDownErr
}

// TaskRelationer methods

func (m *MockTaskService) AddParentRelation(_ context.Context, taskID, parentID int, relationTypeID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("AddParentRelation", taskID, map[string]any{
		"parentID":       parentID,
		"relationTypeID": relationTypeID,
	})
	return m.AddParentRelationErr
}

func (m *MockTaskService) AddChildRelation(_ context.Context, taskID, childID int, relationTypeID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("AddChildRelation", taskID, map[string]any{
		"childID":        childID,
		"relationTypeID": relationTypeID,
	})
	return m.AddChildRelationErr
}

func (m *MockTaskService) RemoveParentRelation(_ context.Context, taskID, parentID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("RemoveParentRelation", taskID, map[string]any{
		"parentID": parentID,
	})
	return m.RemoveParentRelationErr
}

func (m *MockTaskService) RemoveChildRelation(_ context.Context, taskID, childID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("RemoveChildRelation", taskID, map[string]any{
		"childID": childID,
	})
	return m.RemoveChildRelationErr
}

// TaskLabeler methods

func (m *MockTaskService) AttachLabel(_ context.Context, taskID, labelID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("AttachLabel", taskID, map[string]any{
		"labelID": labelID,
	})
	return m.AttachLabelErr
}

func (m *MockTaskService) DetachLabel(_ context.Context, taskID, labelID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DetachLabel", taskID, map[string]any{
		"labelID": labelID,
	})
	return m.DetachLabelErr
}

// TaskCommenter methods

func (m *MockTaskService) CreateComment(_ context.Context, req task.CreateCommentRequest) (*models.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("CreateComment", req.TaskID, map[string]any{
		"message": req.Message,
		"author":  req.Author,
	})
	return m.CreateCommentResult, m.CreateCommentErr
}

func (m *MockTaskService) UpdateComment(_ context.Context, req task.UpdateCommentRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("UpdateComment", 0, map[string]any{
		"commentID": req.CommentID,
		"message":   req.Message,
	})
	return m.UpdateCommentErr
}

func (m *MockTaskService) DeleteComment(_ context.Context, commentID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DeleteComment", 0, map[string]any{
		"commentID": commentID,
	})
	return m.DeleteCommentErr
}

func (m *MockTaskService) GetCommentsByTask(_ context.Context, taskID int) ([]*models.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("GetCommentsByTask", taskID, nil)
	return m.GetCommentsByTaskResult, m.GetCommentsByTaskErr
}
