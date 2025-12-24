package task

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
)

// mockDataStore implements database.DataStore for testing
type mockDataStore struct {
	// Task methods
	CreateTaskFn                        func(ctx context.Context, title, description string, columnID, position int) (*models.Task, error)
	UpdateTaskFn                        func(ctx context.Context, id int, title, description string) error
	UpdateTaskPriorityFn                func(ctx context.Context, taskID, priorityID int) error
	UpdateTaskTypeFn                    func(ctx context.Context, taskID, typeID int) error
	DeleteTaskFn                        func(ctx context.Context, id int) error
	GetTaskDetailFn                     func(ctx context.Context, id int) (*models.TaskDetail, error)
	GetTaskSummariesByProjectFn         func(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error)
	GetTaskSummariesByProjectFilteredFn func(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error)
	GetTaskReferencesForProjectFn       func(ctx context.Context, projectID int) ([]*models.TaskReference, error)
	MoveTaskToNextColumnFn              func(ctx context.Context, taskID int) error
	MoveTaskToPrevColumnFn              func(ctx context.Context, taskID int) error
	MoveTaskToColumnFn                  func(ctx context.Context, taskID int, targetColumnID int) error
	SwapTaskUpFn                        func(ctx context.Context, taskID int) error
	SwapTaskDownFn                      func(ctx context.Context, taskID int) error
	AddSubtaskWithRelationTypeFn        func(ctx context.Context, parentID, childID, relationTypeID int) error
	RemoveSubtaskFn                     func(ctx context.Context, parentID, childID int) error
	GetTaskSummariesByColumnFn          func(ctx context.Context, columnID int) ([]*models.TaskSummary, error)
	GetTasksByColumnFn                  func(ctx context.Context, columnID int) ([]*models.Task, error)
	GetTaskCountByColumnFn              func(ctx context.Context, columnID int) (int, error)
	GetParentTasksFn                    func(ctx context.Context, taskID int) ([]*models.TaskReference, error)
	GetChildTasksFn                     func(ctx context.Context, taskID int) ([]*models.TaskReference, error)
	GetAllRelationTypesFn               func(ctx context.Context) ([]*models.RelationType, error)
	AddSubtaskFn                        func(ctx context.Context, parentID, childID int) error

	// Label methods
	AddLabelToTaskFn      func(ctx context.Context, taskID, labelID int) error
	RemoveLabelFromTaskFn func(ctx context.Context, taskID, labelID int) error
	SetTaskLabelsFn       func(ctx context.Context, taskID int, labelIDs []int) error
	GetLabelsByProjectFn  func(ctx context.Context, projectID int) ([]*models.Label, error)
	GetLabelsForTaskFn    func(ctx context.Context, taskID int) ([]*models.Label, error)
	CreateLabelFn         func(ctx context.Context, projectID int, name, color string) (*models.Label, error)
	UpdateLabelFn         func(ctx context.Context, id int, name, color string) error
	DeleteLabelFn         func(ctx context.Context, id int) error

	// Project methods (stubs)
	GetAllProjectsFn func(ctx context.Context) ([]*models.Project, error)
	GetProjectByIDFn func(ctx context.Context, id int) (*models.Project, error)
	CreateProjectFn  func(ctx context.Context, name, description string) (*models.Project, error)
	UpdateProjectFn  func(ctx context.Context, id int, name, description string) error
	DeleteProjectFn  func(ctx context.Context, id int) error

	// Column methods (stubs)
	GetColumnsByProjectFn func(ctx context.Context, projectID int) ([]*models.Column, error)
	GetColumnByIDFn       func(ctx context.Context, id int) (*models.Column, error)
	CreateColumnFn        func(ctx context.Context, name string, projectID int, afterID *int) (*models.Column, error)
	UpdateColumnNameFn    func(ctx context.Context, id int, name string) error
	DeleteColumnFn        func(ctx context.Context, id int) error
}

// Task Reader methods
func (m *mockDataStore) GetTaskSummariesByColumn(ctx context.Context, columnID int) ([]*models.TaskSummary, error) {
	if m.GetTaskSummariesByColumnFn != nil {
		return m.GetTaskSummariesByColumnFn(ctx, columnID)
	}
	return nil, nil
}

func (m *mockDataStore) GetTaskSummariesByProject(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
	if m.GetTaskSummariesByProjectFn != nil {
		return m.GetTaskSummariesByProjectFn(ctx, projectID)
	}
	return nil, nil
}

func (m *mockDataStore) GetTaskSummariesByProjectFiltered(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error) {
	if m.GetTaskSummariesByProjectFilteredFn != nil {
		return m.GetTaskSummariesByProjectFilteredFn(ctx, projectID, searchQuery)
	}
	return nil, nil
}

func (m *mockDataStore) GetTasksByColumn(ctx context.Context, columnID int) ([]*models.Task, error) {
	if m.GetTasksByColumnFn != nil {
		return m.GetTasksByColumnFn(ctx, columnID)
	}
	return nil, nil
}

func (m *mockDataStore) GetTaskDetail(ctx context.Context, id int) (*models.TaskDetail, error) {
	if m.GetTaskDetailFn != nil {
		return m.GetTaskDetailFn(ctx, id)
	}
	return nil, nil
}

func (m *mockDataStore) GetTaskCountByColumn(ctx context.Context, columnID int) (int, error) {
	if m.GetTaskCountByColumnFn != nil {
		return m.GetTaskCountByColumnFn(ctx, columnID)
	}
	return 0, nil
}

// Task Writer methods
func (m *mockDataStore) CreateTask(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
	if m.CreateTaskFn != nil {
		return m.CreateTaskFn(ctx, title, description, columnID, position)
	}
	return nil, nil
}

func (m *mockDataStore) UpdateTask(ctx context.Context, id int, title, description string) error {
	if m.UpdateTaskFn != nil {
		return m.UpdateTaskFn(ctx, id, title, description)
	}
	return nil
}

func (m *mockDataStore) UpdateTaskPriority(ctx context.Context, taskID, priorityID int) error {
	if m.UpdateTaskPriorityFn != nil {
		return m.UpdateTaskPriorityFn(ctx, taskID, priorityID)
	}
	return nil
}

func (m *mockDataStore) UpdateTaskType(ctx context.Context, taskID, typeID int) error {
	if m.UpdateTaskTypeFn != nil {
		return m.UpdateTaskTypeFn(ctx, taskID, typeID)
	}
	return nil
}

func (m *mockDataStore) DeleteTask(ctx context.Context, id int) error {
	if m.DeleteTaskFn != nil {
		return m.DeleteTaskFn(ctx, id)
	}
	return nil
}

// Task Mover methods
func (m *mockDataStore) MoveTaskToNextColumn(ctx context.Context, taskID int) error {
	if m.MoveTaskToNextColumnFn != nil {
		return m.MoveTaskToNextColumnFn(ctx, taskID)
	}
	return nil
}

func (m *mockDataStore) MoveTaskToPrevColumn(ctx context.Context, taskID int) error {
	if m.MoveTaskToPrevColumnFn != nil {
		return m.MoveTaskToPrevColumnFn(ctx, taskID)
	}
	return nil
}

func (m *mockDataStore) MoveTaskToColumn(ctx context.Context, taskID int, targetColumnID int) error {
	if m.MoveTaskToColumnFn != nil {
		return m.MoveTaskToColumnFn(ctx, taskID, targetColumnID)
	}
	return nil
}

func (m *mockDataStore) SwapTaskUp(ctx context.Context, taskID int) error {
	if m.SwapTaskUpFn != nil {
		return m.SwapTaskUpFn(ctx, taskID)
	}
	return nil
}

func (m *mockDataStore) SwapTaskDown(ctx context.Context, taskID int) error {
	if m.SwapTaskDownFn != nil {
		return m.SwapTaskDownFn(ctx, taskID)
	}
	return nil
}

// Task Relationship Reader methods
func (m *mockDataStore) GetParentTasks(ctx context.Context, taskID int) ([]*models.TaskReference, error) {
	if m.GetParentTasksFn != nil {
		return m.GetParentTasksFn(ctx, taskID)
	}
	return nil, nil
}

func (m *mockDataStore) GetChildTasks(ctx context.Context, taskID int) ([]*models.TaskReference, error) {
	if m.GetChildTasksFn != nil {
		return m.GetChildTasksFn(ctx, taskID)
	}
	return nil, nil
}

func (m *mockDataStore) GetTaskReferencesForProject(ctx context.Context, projectID int) ([]*models.TaskReference, error) {
	if m.GetTaskReferencesForProjectFn != nil {
		return m.GetTaskReferencesForProjectFn(ctx, projectID)
	}
	return nil, nil
}

func (m *mockDataStore) GetAllRelationTypes(ctx context.Context) ([]*models.RelationType, error) {
	if m.GetAllRelationTypesFn != nil {
		return m.GetAllRelationTypesFn(ctx)
	}
	return nil, nil
}

// Task Relationship Writer methods
func (m *mockDataStore) AddSubtask(ctx context.Context, parentID, childID int) error {
	if m.AddSubtaskFn != nil {
		return m.AddSubtaskFn(ctx, parentID, childID)
	}
	return nil
}

func (m *mockDataStore) AddSubtaskWithRelationType(ctx context.Context, parentID, childID, relationTypeID int) error {
	if m.AddSubtaskWithRelationTypeFn != nil {
		return m.AddSubtaskWithRelationTypeFn(ctx, parentID, childID, relationTypeID)
	}
	return nil
}

func (m *mockDataStore) RemoveSubtask(ctx context.Context, parentID, childID int) error {
	if m.RemoveSubtaskFn != nil {
		return m.RemoveSubtaskFn(ctx, parentID, childID)
	}
	return nil
}

// Label Reader methods
func (m *mockDataStore) GetLabelsByProject(ctx context.Context, projectID int) ([]*models.Label, error) {
	if m.GetLabelsByProjectFn != nil {
		return m.GetLabelsByProjectFn(ctx, projectID)
	}
	return nil, nil
}

func (m *mockDataStore) GetLabelsForTask(ctx context.Context, taskID int) ([]*models.Label, error) {
	if m.GetLabelsForTaskFn != nil {
		return m.GetLabelsForTaskFn(ctx, taskID)
	}
	return nil, nil
}

// Label Writer methods
func (m *mockDataStore) CreateLabel(ctx context.Context, projectID int, name, color string) (*models.Label, error) {
	if m.CreateLabelFn != nil {
		return m.CreateLabelFn(ctx, projectID, name, color)
	}
	return nil, nil
}

func (m *mockDataStore) UpdateLabel(ctx context.Context, id int, name, color string) error {
	if m.UpdateLabelFn != nil {
		return m.UpdateLabelFn(ctx, id, name, color)
	}
	return nil
}

func (m *mockDataStore) DeleteLabel(ctx context.Context, id int) error {
	if m.DeleteLabelFn != nil {
		return m.DeleteLabelFn(ctx, id)
	}
	return nil
}

// Task Label Manager methods
func (m *mockDataStore) AddLabelToTask(ctx context.Context, taskID, labelID int) error {
	if m.AddLabelToTaskFn != nil {
		return m.AddLabelToTaskFn(ctx, taskID, labelID)
	}
	return nil
}

func (m *mockDataStore) RemoveLabelFromTask(ctx context.Context, taskID, labelID int) error {
	if m.RemoveLabelFromTaskFn != nil {
		return m.RemoveLabelFromTaskFn(ctx, taskID, labelID)
	}
	return nil
}

func (m *mockDataStore) SetTaskLabels(ctx context.Context, taskID int, labelIDs []int) error {
	if m.SetTaskLabelsFn != nil {
		return m.SetTaskLabelsFn(ctx, taskID, labelIDs)
	}
	return nil
}

// Project Reader methods (stubs for interface satisfaction)
func (m *mockDataStore) GetAllProjects(ctx context.Context) ([]*models.Project, error) {
	if m.GetAllProjectsFn != nil {
		return m.GetAllProjectsFn(ctx)
	}
	return nil, nil
}

func (m *mockDataStore) GetProjectByID(ctx context.Context, id int) (*models.Project, error) {
	if m.GetProjectByIDFn != nil {
		return m.GetProjectByIDFn(ctx, id)
	}
	return nil, nil
}

// Project Writer methods (stubs for interface satisfaction)
func (m *mockDataStore) CreateProject(ctx context.Context, name, description string) (*models.Project, error) {
	if m.CreateProjectFn != nil {
		return m.CreateProjectFn(ctx, name, description)
	}
	return nil, nil
}

func (m *mockDataStore) UpdateProject(ctx context.Context, id int, name, description string) error {
	if m.UpdateProjectFn != nil {
		return m.UpdateProjectFn(ctx, id, name, description)
	}
	return nil
}

func (m *mockDataStore) DeleteProject(ctx context.Context, id int) error {
	if m.DeleteProjectFn != nil {
		return m.DeleteProjectFn(ctx, id)
	}
	return nil
}

// Column Reader methods (stubs for interface satisfaction)
func (m *mockDataStore) GetColumnsByProject(ctx context.Context, projectID int) ([]*models.Column, error) {
	if m.GetColumnsByProjectFn != nil {
		return m.GetColumnsByProjectFn(ctx, projectID)
	}
	return nil, nil
}

func (m *mockDataStore) GetColumnByID(ctx context.Context, id int) (*models.Column, error) {
	if m.GetColumnByIDFn != nil {
		return m.GetColumnByIDFn(ctx, id)
	}
	return nil, nil
}

// Column Writer methods (stubs for interface satisfaction)
func (m *mockDataStore) CreateColumn(ctx context.Context, name string, projectID int, afterID *int) (*models.Column, error) {
	if m.CreateColumnFn != nil {
		return m.CreateColumnFn(ctx, name, projectID, afterID)
	}
	return nil, nil
}

func (m *mockDataStore) UpdateColumnName(ctx context.Context, id int, name string) error {
	if m.UpdateColumnNameFn != nil {
		return m.UpdateColumnNameFn(ctx, id, name)
	}
	return nil
}

func (m *mockDataStore) DeleteColumn(ctx context.Context, id int) error {
	if m.DeleteColumnFn != nil {
		return m.DeleteColumnFn(ctx, id)
	}
	return nil
}

// ============================================================================
// TEST: CreateTask
// ============================================================================

func TestCreateTask_Success(t *testing.T) {
	mock := &mockDataStore{
		CreateTaskFn: func(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
			if title != "Test Task" {
				t.Errorf("Expected title 'Test Task', got '%s'", title)
			}
			if description != "Test Description" {
				t.Errorf("Expected description 'Test Description', got '%s'", description)
			}
			if columnID != 1 {
				t.Errorf("Expected columnID 1, got %d", columnID)
			}
			if position != 0 {
				t.Errorf("Expected position 0, got %d", position)
			}
			return &models.Task{ID: 42, Title: title, Description: description, ColumnID: columnID}, nil
		},
	}

	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		ColumnID:    1,
		Position:    0,
	}

	task, err := svc.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if task == nil {
		t.Fatal("Expected task, got nil")
	}
	if task.ID != 42 {
		t.Errorf("Expected task ID 42, got %d", task.ID)
	}
}

func TestCreateTask_WithPriorityAndType(t *testing.T) {
	createCalled := false
	priorityCalled := false
	typeCalled := false

	mock := &mockDataStore{
		CreateTaskFn: func(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
			createCalled = true
			return &models.Task{ID: 1, Title: title, Description: description, ColumnID: columnID}, nil
		},
		UpdateTaskPriorityFn: func(ctx context.Context, taskID, priorityID int) error {
			priorityCalled = true
			if taskID != 1 {
				t.Errorf("Expected taskID 1, got %d", taskID)
			}
			if priorityID != 2 {
				t.Errorf("Expected priorityID 2, got %d", priorityID)
			}
			return nil
		},
		UpdateTaskTypeFn: func(ctx context.Context, taskID, typeID int) error {
			typeCalled = true
			if taskID != 1 {
				t.Errorf("Expected taskID 1, got %d", taskID)
			}
			if typeID != 3 {
				t.Errorf("Expected typeID 3, got %d", typeID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:      "Test Task",
		ColumnID:   1,
		Position:   0,
		PriorityID: 2,
		TypeID:     3,
	}

	_, err := svc.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !createCalled {
		t.Error("Expected CreateTask to be called")
	}
	if !priorityCalled {
		t.Error("Expected UpdateTaskPriority to be called")
	}
	if !typeCalled {
		t.Error("Expected UpdateTaskType to be called")
	}
}

func TestCreateTask_WithLabels(t *testing.T) {
	var addedLabels []int

	mock := &mockDataStore{
		CreateTaskFn: func(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
			return &models.Task{ID: 1, Title: title, Description: description, ColumnID: columnID}, nil
		},
		AddLabelToTaskFn: func(ctx context.Context, taskID, labelID int) error {
			if taskID != 1 {
				t.Errorf("Expected taskID 1, got %d", taskID)
			}
			addedLabels = append(addedLabels, labelID)
			return nil
		},
	}

	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: 1,
		Position: 0,
		LabelIDs: []int{10, 20, 30},
	}

	_, err := svc.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(addedLabels) != 3 {
		t.Fatalf("Expected 3 labels added, got %d", len(addedLabels))
	}
	if addedLabels[0] != 10 || addedLabels[1] != 20 || addedLabels[2] != 30 {
		t.Errorf("Labels not added in correct order: %v", addedLabels)
	}
}

func TestCreateTask_EmptyTitle(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:    "",
		ColumnID: 1,
		Position: 0,
	}

	_, err := svc.CreateTask(context.Background(), req)
	if err != ErrEmptyTitle {
		t.Errorf("Expected ErrEmptyTitle, got %v", err)
	}
}

func TestCreateTask_TitleTooLong(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:    strings.Repeat("a", 256), // 256 characters
		ColumnID: 1,
		Position: 0,
	}

	_, err := svc.CreateTask(context.Background(), req)
	if err != ErrTitleTooLong {
		t.Errorf("Expected ErrTitleTooLong, got %v", err)
	}
}

func TestCreateTask_InvalidColumnID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name     string
		columnID int
	}{
		{"zero column ID", 0},
		{"negative column ID", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateTaskRequest{
				Title:    "Test Task",
				ColumnID: tc.columnID,
				Position: 0,
			}

			_, err := svc.CreateTask(context.Background(), req)
			if err != ErrInvalidColumnID {
				t.Errorf("Expected ErrInvalidColumnID, got %v", err)
			}
		})
	}
}

func TestCreateTask_InvalidPosition(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: 1,
		Position: -1,
	}

	_, err := svc.CreateTask(context.Background(), req)
	if err != ErrInvalidPosition {
		t.Errorf("Expected ErrInvalidPosition, got %v", err)
	}
}

func TestCreateTask_InvalidPriority(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:      "Test Task",
		ColumnID:   1,
		Position:   0,
		PriorityID: -1,
	}

	_, err := svc.CreateTask(context.Background(), req)
	if err != ErrInvalidPriority {
		t.Errorf("Expected ErrInvalidPriority, got %v", err)
	}
}

func TestCreateTask_InvalidType(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: 1,
		Position: 0,
		TypeID:   -1,
	}

	_, err := svc.CreateTask(context.Background(), req)
	if err != ErrInvalidType {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}

// ============================================================================
// TEST: UpdateTask
// ============================================================================

func TestUpdateTask_TitleOnly(t *testing.T) {
	mock := &mockDataStore{
		GetTaskDetailFn: func(ctx context.Context, id int) (*models.TaskDetail, error) {
			return &models.TaskDetail{
				ID:          1,
				Title:       "Old Title",
				Description: "Old Description",
			}, nil
		},
		UpdateTaskFn: func(ctx context.Context, id int, title, description string) error {
			if id != 1 {
				t.Errorf("Expected taskID 1, got %d", id)
			}
			if title != "New Title" {
				t.Errorf("Expected title 'New Title', got '%s'", title)
			}
			if description != "Old Description" {
				t.Errorf("Expected description 'Old Description', got '%s'", description)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	newTitle := "New Title"
	req := UpdateTaskRequest{
		TaskID: 1,
		Title:  &newTitle,
	}

	err := svc.UpdateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUpdateTask_DescriptionOnly(t *testing.T) {
	mock := &mockDataStore{
		GetTaskDetailFn: func(ctx context.Context, id int) (*models.TaskDetail, error) {
			return &models.TaskDetail{
				ID:          1,
				Title:       "Old Title",
				Description: "Old Description",
			}, nil
		},
		UpdateTaskFn: func(ctx context.Context, id int, title, description string) error {
			if id != 1 {
				t.Errorf("Expected taskID 1, got %d", id)
			}
			if title != "Old Title" {
				t.Errorf("Expected title 'Old Title', got '%s'", title)
			}
			if description != "New Description" {
				t.Errorf("Expected description 'New Description', got '%s'", description)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	newDesc := "New Description"
	req := UpdateTaskRequest{
		TaskID:      1,
		Description: &newDesc,
	}

	err := svc.UpdateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUpdateTask_PriorityAndType(t *testing.T) {
	priorityCalled := false
	typeCalled := false

	mock := &mockDataStore{
		UpdateTaskPriorityFn: func(ctx context.Context, taskID, priorityID int) error {
			priorityCalled = true
			if taskID != 1 {
				t.Errorf("Expected taskID 1, got %d", taskID)
			}
			if priorityID != 5 {
				t.Errorf("Expected priorityID 5, got %d", priorityID)
			}
			return nil
		},
		UpdateTaskTypeFn: func(ctx context.Context, taskID, typeID int) error {
			typeCalled = true
			if taskID != 1 {
				t.Errorf("Expected taskID 1, got %d", taskID)
			}
			if typeID != 7 {
				t.Errorf("Expected typeID 7, got %d", typeID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	priority := 5
	taskType := 7
	req := UpdateTaskRequest{
		TaskID:     1,
		PriorityID: &priority,
		TypeID:     &taskType,
	}

	err := svc.UpdateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !priorityCalled {
		t.Error("Expected UpdateTaskPriority to be called")
	}
	if !typeCalled {
		t.Error("Expected UpdateTaskType to be called")
	}
}

func TestUpdateTask_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name   string
		taskID int
	}{
		{"zero task ID", 0},
		{"negative task ID", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := UpdateTaskRequest{
				TaskID: tc.taskID,
			}

			err := svc.UpdateTask(context.Background(), req)
			if err != ErrInvalidTaskID {
				t.Errorf("Expected ErrInvalidTaskID, got %v", err)
			}
		})
	}
}

func TestUpdateTask_EmptyTitle(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	emptyTitle := ""
	req := UpdateTaskRequest{
		TaskID: 1,
		Title:  &emptyTitle,
	}

	err := svc.UpdateTask(context.Background(), req)
	if err != ErrEmptyTitle {
		t.Errorf("Expected ErrEmptyTitle, got %v", err)
	}
}

func TestUpdateTask_TitleTooLong(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	longTitle := strings.Repeat("a", 256)
	req := UpdateTaskRequest{
		TaskID: 1,
		Title:  &longTitle,
	}

	err := svc.UpdateTask(context.Background(), req)
	if err != ErrTitleTooLong {
		t.Errorf("Expected ErrTitleTooLong, got %v", err)
	}
}

func TestUpdateTask_InvalidPriority(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	invalidPriority := -1
	req := UpdateTaskRequest{
		TaskID:     1,
		PriorityID: &invalidPriority,
	}

	err := svc.UpdateTask(context.Background(), req)
	if err != ErrInvalidPriority {
		t.Errorf("Expected ErrInvalidPriority, got %v", err)
	}
}

func TestUpdateTask_InvalidType(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	invalidType := 0
	req := UpdateTaskRequest{
		TaskID: 1,
		TypeID: &invalidType,
	}

	err := svc.UpdateTask(context.Background(), req)
	if err != ErrInvalidType {
		t.Errorf("Expected ErrInvalidType, got %v", err)
	}
}

// ============================================================================
// TEST: DeleteTask
// ============================================================================

func TestDeleteTask_Success(t *testing.T) {
	deleteCalled := false
	mock := &mockDataStore{
		DeleteTaskFn: func(ctx context.Context, id int) error {
			deleteCalled = true
			if id != 42 {
				t.Errorf("Expected taskID 42, got %d", id)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.DeleteTask(context.Background(), 42)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !deleteCalled {
		t.Error("Expected DeleteTask to be called")
	}
}

func TestDeleteTask_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name   string
		taskID int
	}{
		{"zero task ID", 0},
		{"negative task ID", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.DeleteTask(context.Background(), tc.taskID)
			if err != ErrInvalidTaskID {
				t.Errorf("Expected ErrInvalidTaskID, got %v", err)
			}
		})
	}
}

// ============================================================================
// TEST: GetTaskDetail
// ============================================================================

func TestGetTaskDetail_Success(t *testing.T) {
	expectedDetail := &models.TaskDetail{
		ID:    1,
		Title: "Test Task",
	}

	mock := &mockDataStore{
		GetTaskDetailFn: func(ctx context.Context, id int) (*models.TaskDetail, error) {
			if id != 1 {
				t.Errorf("Expected taskID 1, got %d", id)
			}
			return expectedDetail, nil
		},
	}

	svc := NewService(mock, nil)
	detail, err := svc.GetTaskDetail(context.Background(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if detail != expectedDetail {
		t.Error("Expected returned detail to match")
	}
}

func TestGetTaskDetail_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name   string
		taskID int
	}{
		{"zero task ID", 0},
		{"negative task ID", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.GetTaskDetail(context.Background(), tc.taskID)
			if err != ErrInvalidTaskID {
				t.Errorf("Expected ErrInvalidTaskID, got %v", err)
			}
		})
	}
}

// ============================================================================
// TEST: Movement Operations
// ============================================================================

func TestMoveTaskToNextColumn_Success(t *testing.T) {
	moveCalled := false
	mock := &mockDataStore{
		MoveTaskToNextColumnFn: func(ctx context.Context, taskID int) error {
			moveCalled = true
			if taskID != 5 {
				t.Errorf("Expected taskID 5, got %d", taskID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskToNextColumn(context.Background(), 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !moveCalled {
		t.Error("Expected MoveTaskToNextColumn to be called")
	}
}

func TestMoveTaskToNextColumn_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	err := svc.MoveTaskToNextColumn(context.Background(), 0)
	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

func TestMoveTaskToPrevColumn_Success(t *testing.T) {
	moveCalled := false
	mock := &mockDataStore{
		MoveTaskToPrevColumnFn: func(ctx context.Context, taskID int) error {
			moveCalled = true
			if taskID != 5 {
				t.Errorf("Expected taskID 5, got %d", taskID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskToPrevColumn(context.Background(), 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !moveCalled {
		t.Error("Expected MoveTaskToPrevColumn to be called")
	}
}

func TestMoveTaskToPrevColumn_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	err := svc.MoveTaskToPrevColumn(context.Background(), -1)
	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

func TestMoveTaskToColumn_Success(t *testing.T) {
	moveCalled := false
	mock := &mockDataStore{
		MoveTaskToColumnFn: func(ctx context.Context, taskID, columnID int) error {
			moveCalled = true
			if taskID != 5 {
				t.Errorf("Expected taskID 5, got %d", taskID)
			}
			if columnID != 3 {
				t.Errorf("Expected columnID 3, got %d", columnID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskToColumn(context.Background(), 5, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !moveCalled {
		t.Error("Expected MoveTaskToColumn to be called")
	}
}

func TestMoveTaskToColumn_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	err := svc.MoveTaskToColumn(context.Background(), 0, 1)
	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

func TestMoveTaskToColumn_InvalidColumnID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	err := svc.MoveTaskToColumn(context.Background(), 1, 0)
	if err != ErrInvalidColumnID {
		t.Errorf("Expected ErrInvalidColumnID, got %v", err)
	}
}

func TestMoveTaskUp_Success(t *testing.T) {
	swapCalled := false
	mock := &mockDataStore{
		SwapTaskUpFn: func(ctx context.Context, taskID int) error {
			swapCalled = true
			if taskID != 7 {
				t.Errorf("Expected taskID 7, got %d", taskID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskUp(context.Background(), 7)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !swapCalled {
		t.Error("Expected SwapTaskUp to be called")
	}
}

func TestMoveTaskUp_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	err := svc.MoveTaskUp(context.Background(), -5)
	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

func TestMoveTaskDown_Success(t *testing.T) {
	swapCalled := false
	mock := &mockDataStore{
		SwapTaskDownFn: func(ctx context.Context, taskID int) error {
			swapCalled = true
			if taskID != 8 {
				t.Errorf("Expected taskID 8, got %d", taskID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskDown(context.Background(), 8)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !swapCalled {
		t.Error("Expected SwapTaskDown to be called")
	}
}

func TestMoveTaskDown_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	err := svc.MoveTaskDown(context.Background(), 0)
	if err != ErrInvalidTaskID {
		t.Errorf("Expected ErrInvalidTaskID, got %v", err)
	}
}

// ============================================================================
// TEST: Relationship Operations
// ============================================================================

func TestAddParentRelation_Success(t *testing.T) {
	addCalled := false
	mock := &mockDataStore{
		AddSubtaskWithRelationTypeFn: func(ctx context.Context, parentID, childID, relationTypeID int) error {
			addCalled = true
			if parentID != 10 {
				t.Errorf("Expected parentID 10, got %d", parentID)
			}
			if childID != 5 {
				t.Errorf("Expected childID 5, got %d", childID)
			}
			if relationTypeID != 1 {
				t.Errorf("Expected relationTypeID 1, got %d", relationTypeID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.AddParentRelation(context.Background(), 5, 10, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !addCalled {
		t.Error("Expected AddSubtaskWithRelationType to be called")
	}
}

func TestAddParentRelation_SelfRelation(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	err := svc.AddParentRelation(context.Background(), 5, 5, 1)
	if err != ErrSelfRelation {
		t.Errorf("Expected ErrSelfRelation, got %v", err)
	}
}

func TestAddParentRelation_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name     string
		taskID   int
		parentID int
	}{
		{"zero task ID", 0, 10},
		{"negative task ID", -1, 10},
		{"zero parent ID", 5, 0},
		{"negative parent ID", 5, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.AddParentRelation(context.Background(), tc.taskID, tc.parentID, 1)
			if err != ErrInvalidTaskID {
				t.Errorf("Expected ErrInvalidTaskID, got %v", err)
			}
		})
	}
}

func TestAddChildRelation_Success(t *testing.T) {
	addCalled := false
	mock := &mockDataStore{
		AddSubtaskWithRelationTypeFn: func(ctx context.Context, parentID, childID, relationTypeID int) error {
			addCalled = true
			if parentID != 5 {
				t.Errorf("Expected parentID 5, got %d", parentID)
			}
			if childID != 10 {
				t.Errorf("Expected childID 10, got %d", childID)
			}
			if relationTypeID != 1 {
				t.Errorf("Expected relationTypeID 1, got %d", relationTypeID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.AddChildRelation(context.Background(), 5, 10, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !addCalled {
		t.Error("Expected AddSubtaskWithRelationType to be called")
	}
}

func TestAddChildRelation_SelfRelation(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)
	err := svc.AddChildRelation(context.Background(), 5, 5, 1)
	if err != ErrSelfRelation {
		t.Errorf("Expected ErrSelfRelation, got %v", err)
	}
}

func TestAddChildRelation_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name    string
		taskID  int
		childID int
	}{
		{"zero task ID", 0, 10},
		{"negative task ID", -1, 10},
		{"zero child ID", 5, 0},
		{"negative child ID", 5, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.AddChildRelation(context.Background(), tc.taskID, tc.childID, 1)
			if err != ErrInvalidTaskID {
				t.Errorf("Expected ErrInvalidTaskID, got %v", err)
			}
		})
	}
}

func TestRemoveParentRelation_Success(t *testing.T) {
	removeCalled := false
	mock := &mockDataStore{
		RemoveSubtaskFn: func(ctx context.Context, parentID, childID int) error {
			removeCalled = true
			if parentID != 10 {
				t.Errorf("Expected parentID 10, got %d", parentID)
			}
			if childID != 5 {
				t.Errorf("Expected childID 5, got %d", childID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.RemoveParentRelation(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !removeCalled {
		t.Error("Expected RemoveSubtask to be called")
	}
}

func TestRemoveParentRelation_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name     string
		taskID   int
		parentID int
	}{
		{"zero task ID", 0, 10},
		{"negative task ID", -1, 10},
		{"zero parent ID", 5, 0},
		{"negative parent ID", 5, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.RemoveParentRelation(context.Background(), tc.taskID, tc.parentID)
			if err != ErrInvalidTaskID {
				t.Errorf("Expected ErrInvalidTaskID, got %v", err)
			}
		})
	}
}

func TestRemoveChildRelation_Success(t *testing.T) {
	removeCalled := false
	mock := &mockDataStore{
		RemoveSubtaskFn: func(ctx context.Context, parentID, childID int) error {
			removeCalled = true
			if parentID != 5 {
				t.Errorf("Expected parentID 5, got %d", parentID)
			}
			if childID != 10 {
				t.Errorf("Expected childID 10, got %d", childID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.RemoveChildRelation(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !removeCalled {
		t.Error("Expected RemoveSubtask to be called")
	}
}

func TestRemoveChildRelation_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name    string
		taskID  int
		childID int
	}{
		{"zero task ID", 0, 10},
		{"negative task ID", -1, 10},
		{"zero child ID", 5, 0},
		{"negative child ID", 5, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.RemoveChildRelation(context.Background(), tc.taskID, tc.childID)
			if err != ErrInvalidTaskID {
				t.Errorf("Expected ErrInvalidTaskID, got %v", err)
			}
		})
	}
}

// ============================================================================
// TEST: Label Operations
// ============================================================================

func TestAttachLabel_Success(t *testing.T) {
	attachCalled := false
	mock := &mockDataStore{
		AddLabelToTaskFn: func(ctx context.Context, taskID, labelID int) error {
			attachCalled = true
			if taskID != 5 {
				t.Errorf("Expected taskID 5, got %d", taskID)
			}
			if labelID != 10 {
				t.Errorf("Expected labelID 10, got %d", labelID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.AttachLabel(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !attachCalled {
		t.Error("Expected AddLabelToTask to be called")
	}
}

func TestAttachLabel_InvalidTaskID(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name    string
		taskID  int
		labelID int
		wantErr error
	}{
		{"zero task ID", 0, 10, ErrInvalidTaskID},
		{"negative task ID", -1, 10, ErrInvalidTaskID},
		{"zero label ID", 5, 0, ErrInvalidLabelID},
		{"negative label ID", 5, -1, ErrInvalidLabelID},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.AttachLabel(context.Background(), tc.taskID, tc.labelID)
			if err != tc.wantErr {
				t.Errorf("Expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestDetachLabel_Success(t *testing.T) {
	detachCalled := false
	mock := &mockDataStore{
		RemoveLabelFromTaskFn: func(ctx context.Context, taskID, labelID int) error {
			detachCalled = true
			if taskID != 5 {
				t.Errorf("Expected taskID 5, got %d", taskID)
			}
			if labelID != 10 {
				t.Errorf("Expected labelID 10, got %d", labelID)
			}
			return nil
		},
	}

	svc := NewService(mock, nil)
	err := svc.DetachLabel(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !detachCalled {
		t.Error("Expected RemoveLabelFromTask to be called")
	}
}

func TestDetachLabel_InvalidIDs(t *testing.T) {
	mock := &mockDataStore{}
	svc := NewService(mock, nil)

	testCases := []struct {
		name    string
		taskID  int
		labelID int
		wantErr error
	}{
		{"zero task ID", 0, 10, ErrInvalidTaskID},
		{"negative task ID", -1, 10, ErrInvalidTaskID},
		{"zero label ID", 5, 0, ErrInvalidLabelID},
		{"negative label ID", 5, -1, ErrInvalidLabelID},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.DetachLabel(context.Background(), tc.taskID, tc.labelID)
			if err != tc.wantErr {
				t.Errorf("Expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// ============================================================================
// TEST: Error Propagation
// ============================================================================

func TestCreateTask_RepositoryError(t *testing.T) {
	repoErr := errors.New("database connection failed")
	mock := &mockDataStore{
		CreateTaskFn: func(ctx context.Context, title, description string, columnID, position int) (*models.Task, error) {
			return nil, repoErr
		},
	}

	svc := NewService(mock, nil)
	req := CreateTaskRequest{
		Title:    "Test Task",
		ColumnID: 1,
		Position: 0,
	}

	_, err := svc.CreateTask(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	// Error should be wrapped
	if !errors.Is(err, repoErr) {
		// Check if error message contains the wrapped error
		if !strings.Contains(err.Error(), "failed to create task") {
			t.Errorf("Expected wrapped error, got %v", err)
		}
	}
}

func TestUpdateTask_RepositoryError(t *testing.T) {
	repoErr := errors.New("database locked")
	mock := &mockDataStore{
		GetTaskDetailFn: func(ctx context.Context, id int) (*models.TaskDetail, error) {
			return &models.TaskDetail{
				ID:          1,
				Title:       "Test",
				Description: "Test",
			}, nil
		},
		UpdateTaskFn: func(ctx context.Context, id int, title, description string) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	newTitle := "Updated Title"
	req := UpdateTaskRequest{
		TaskID: 1,
		Title:  &newTitle,
	}

	err := svc.UpdateTask(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update task") {
		t.Errorf("Expected wrapped error, got %v", err)
	}
}

func TestDeleteTask_RepositoryError(t *testing.T) {
	repoErr := errors.New("task has dependencies")
	mock := &mockDataStore{
		DeleteTaskFn: func(ctx context.Context, id int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.DeleteTask(context.Background(), 1)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to delete task") {
		t.Errorf("Expected wrapped error, got %v", err)
	}
}

// ============================================================================
// TEST: Read Operations (Passthroughs)
// ============================================================================

func TestGetTaskSummariesByProject_Success(t *testing.T) {
	expectedSummaries := map[int][]*models.TaskSummary{
		1: {{ID: 1, Title: "Task 1"}},
		2: {{ID: 2, Title: "Task 2"}},
	}

	mock := &mockDataStore{
		GetTaskSummariesByProjectFn: func(ctx context.Context, projectID int) (map[int][]*models.TaskSummary, error) {
			if projectID != 5 {
				t.Errorf("Expected projectID 5, got %d", projectID)
			}
			return expectedSummaries, nil
		},
	}

	svc := NewService(mock, nil)
	result, err := svc.GetTaskSummariesByProject(context.Background(), 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(result))
	}
}

func TestGetTaskSummariesByProjectFiltered_Success(t *testing.T) {
	expectedSummaries := map[int][]*models.TaskSummary{
		1: {{ID: 1, Title: "Filtered Task"}},
	}

	mock := &mockDataStore{
		GetTaskSummariesByProjectFilteredFn: func(ctx context.Context, projectID int, searchQuery string) (map[int][]*models.TaskSummary, error) {
			if projectID != 5 {
				t.Errorf("Expected projectID 5, got %d", projectID)
			}
			if searchQuery != "test query" {
				t.Errorf("Expected search query 'test query', got '%s'", searchQuery)
			}
			return expectedSummaries, nil
		},
	}

	svc := NewService(mock, nil)
	result, err := svc.GetTaskSummariesByProjectFiltered(context.Background(), 5, "test query")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 column, got %d", len(result))
	}
}

func TestGetTaskReferencesForProject_Success(t *testing.T) {
	expectedRefs := []*models.TaskReference{
		{ID: 1, Title: "Task 1"},
		{ID: 2, Title: "Task 2"},
	}

	mock := &mockDataStore{
		GetTaskReferencesForProjectFn: func(ctx context.Context, projectID int) ([]*models.TaskReference, error) {
			if projectID != 10 {
				t.Errorf("Expected projectID 10, got %d", projectID)
			}
			return expectedRefs, nil
		},
	}

	svc := NewService(mock, nil)
	result, err := svc.GetTaskReferencesForProject(context.Background(), 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 references, got %d", len(result))
	}
}

// ============================================================================
// TEST: Movement Error Propagation
// ============================================================================

func TestMoveTaskToNextColumn_RepositoryError(t *testing.T) {
	repoErr := errors.New("cannot move beyond last column")
	mock := &mockDataStore{
		MoveTaskToNextColumnFn: func(ctx context.Context, taskID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskToNextColumn(context.Background(), 1)
	if err != repoErr {
		t.Errorf("Expected repository error to be returned, got %v", err)
	}
}

func TestMoveTaskToPrevColumn_RepositoryError(t *testing.T) {
	repoErr := errors.New("cannot move before first column")
	mock := &mockDataStore{
		MoveTaskToPrevColumnFn: func(ctx context.Context, taskID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskToPrevColumn(context.Background(), 1)
	if err != repoErr {
		t.Errorf("Expected repository error to be returned, got %v", err)
	}
}

func TestMoveTaskToColumn_RepositoryError(t *testing.T) {
	repoErr := errors.New("column does not exist")
	mock := &mockDataStore{
		MoveTaskToColumnFn: func(ctx context.Context, taskID, columnID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskToColumn(context.Background(), 1, 99)
	if err != repoErr {
		t.Errorf("Expected repository error to be returned, got %v", err)
	}
}

func TestMoveTaskUp_RepositoryError(t *testing.T) {
	repoErr := errors.New("already at top")
	mock := &mockDataStore{
		SwapTaskUpFn: func(ctx context.Context, taskID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskUp(context.Background(), 1)
	if err != repoErr {
		t.Errorf("Expected repository error to be returned, got %v", err)
	}
}

func TestMoveTaskDown_RepositoryError(t *testing.T) {
	repoErr := errors.New("already at bottom")
	mock := &mockDataStore{
		SwapTaskDownFn: func(ctx context.Context, taskID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.MoveTaskDown(context.Background(), 1)
	if err != repoErr {
		t.Errorf("Expected repository error to be returned, got %v", err)
	}
}

// ============================================================================
// TEST: Relationship Error Propagation
// ============================================================================

func TestAddParentRelation_RepositoryError(t *testing.T) {
	repoErr := errors.New("circular dependency detected")
	mock := &mockDataStore{
		AddSubtaskWithRelationTypeFn: func(ctx context.Context, parentID, childID, relationTypeID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.AddParentRelation(context.Background(), 1, 2, 1)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to add parent relation") {
		t.Errorf("Expected wrapped error, got %v", err)
	}
}

func TestAddChildRelation_RepositoryError(t *testing.T) {
	repoErr := errors.New("circular dependency detected")
	mock := &mockDataStore{
		AddSubtaskWithRelationTypeFn: func(ctx context.Context, parentID, childID, relationTypeID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.AddChildRelation(context.Background(), 1, 2, 1)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to add child relation") {
		t.Errorf("Expected wrapped error, got %v", err)
	}
}

func TestRemoveParentRelation_RepositoryError(t *testing.T) {
	repoErr := errors.New("relationship not found")
	mock := &mockDataStore{
		RemoveSubtaskFn: func(ctx context.Context, parentID, childID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.RemoveParentRelation(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to remove parent relation") {
		t.Errorf("Expected wrapped error, got %v", err)
	}
}

func TestRemoveChildRelation_RepositoryError(t *testing.T) {
	repoErr := errors.New("relationship not found")
	mock := &mockDataStore{
		RemoveSubtaskFn: func(ctx context.Context, parentID, childID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.RemoveChildRelation(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to remove child relation") {
		t.Errorf("Expected wrapped error, got %v", err)
	}
}

// ============================================================================
// TEST: Label Error Propagation
// ============================================================================

func TestAttachLabel_RepositoryError(t *testing.T) {
	repoErr := errors.New("label already attached")
	mock := &mockDataStore{
		AddLabelToTaskFn: func(ctx context.Context, taskID, labelID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.AttachLabel(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to attach label") {
		t.Errorf("Expected wrapped error, got %v", err)
	}
}

func TestDetachLabel_RepositoryError(t *testing.T) {
	repoErr := errors.New("label not attached")
	mock := &mockDataStore{
		RemoveLabelFromTaskFn: func(ctx context.Context, taskID, labelID int) error {
			return repoErr
		},
	}

	svc := NewService(mock, nil)
	err := svc.DetachLabel(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to detach label") {
		t.Errorf("Expected wrapped error, got %v", err)
	}
}
