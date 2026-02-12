package task_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func setupCreateMocks() (*mocks.MockTaskService, *mocks.MockProjectService, *mocks.MockColumnService, *mocks.MockAssigneeService, cli.MockServices) {
	mockTask := mocks.NewMockTaskService()
	mockProject := mocks.NewMockProjectService()
	mockColumn := mocks.NewMockColumnService()
	mockAssignee := mocks.NewMockAssigneeService()

	services := cli.MockServices{
		TaskService:     mockTask,
		ProjectService:  mockProject,
		ColumnService:   mockColumn,
		AssigneeService: mockAssignee,
	}

	return mockTask, mockProject, mockColumn, mockAssignee, services
}

func setDefaultCreateResults(mockTask *mocks.MockTaskService, mockProject *mocks.MockProjectService, mockColumn *mocks.MockColumnService, mockAssignee *mocks.MockAssigneeService) {
	mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
	mockColumn.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	mockAssignee.GetOrCreateResult = &models.Assignee{ID: 5, Name: "testuser"}
	mockTask.CreateTaskResult = &models.Task{
		ID:        42,
		Title:     "My Task",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestCreateTask_BasicCreate(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockAssignee, services := setupCreateMocks()
	setDefaultCreateResults(mockTask, mockProject, mockColumn, mockAssignee)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"create", "-t", "My Task", "-p", "1"})

	require.NoError(t, err)
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "My Task")
	assert.True(t, mockProject.HasCall("GetProjectByID"))
	assert.True(t, mockColumn.HasCall("GetColumnsByProject"))
	assert.Equal(t, 1, mockTask.CallCount("CreateTask"))
}

func TestCreateTask_Quiet(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockAssignee, services := setupCreateMocks()
	setDefaultCreateResults(mockTask, mockProject, mockColumn, mockAssignee)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"create", "-t", "My Task", "-p", "1", "-q"})

	require.NoError(t, err)
	trimmed := strings.TrimSpace(output)
	assert.Equal(t, "42", trimmed)
}

func TestCreateTask_JSON(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockAssignee, services := setupCreateMocks()
	setDefaultCreateResults(mockTask, mockProject, mockColumn, mockAssignee)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"create", "-t", "My Task", "-p", "1", "-j"})

	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result), "output should be valid JSON: %s", output)
	assert.Equal(t, true, result["success"])
	data, ok := result["data"].(map[string]any)
	require.True(t, ok, "data should be a map")
	assert.Equal(t, float64(42), data["ID"])
	assert.Equal(t, "My Task", data["Title"])
}

func TestCreateTask_MissingTitle(t *testing.T) {
	t.Parallel()

	_, _, _, _, services := setupCreateMocks()

	_, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"create", "-p", "1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestCreateTask_ServiceError(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockAssignee, services := setupCreateMocks()
	setDefaultCreateResults(mockTask, mockProject, mockColumn, mockAssignee)
	mockTask.CreateTaskErr = fmt.Errorf("database connection lost")

	_, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"create", "-t", "My Task", "-p", "1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database connection lost")
}

func TestCreateTask_ProjectNotFound(t *testing.T) {
	t.Parallel()

	_, mockProject, _, _, services := setupCreateMocks()
	mockProject.GetProjectByIDErr = fmt.Errorf("not found")

	_, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"create", "-t", "My Task", "-p", "999"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "project")
}

func TestCreateTask_WithDescription(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockAssignee, services := setupCreateMocks()
	setDefaultCreateResults(mockTask, mockProject, mockColumn, mockAssignee)
	mockTask.CreateTaskResult = &models.Task{
		ID:          43,
		Title:       "Described Task",
		Description: "A detailed description",
		CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"create", "-t", "Described Task", "-p", "1", "-d", "A detailed description", "-q"})

	require.NoError(t, err)
	assert.Equal(t, "43", strings.TrimSpace(output))

	calls := mockTask.GetCalls()
	var createCall *mocks.MockCall
	for i := range calls {
		if calls[i].Method == "CreateTask" {
			createCall = &calls[i]
			break
		}
	}
	require.NotNil(t, createCall, "CreateTask should have been called")
	assert.Equal(t, "A detailed description", createCall.Args["description"])
}

func TestCreateTask_WithPriority(t *testing.T) {
	t.Parallel()

	mockTask, mockProject, mockColumn, mockAssignee, services := setupCreateMocks()
	setDefaultCreateResults(mockTask, mockProject, mockColumn, mockAssignee)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"create", "-t", "High Priority Task", "-p", "1", "-r", "high", "-q"})

	require.NoError(t, err)
	assert.Equal(t, "42", strings.TrimSpace(output))

	calls := mockTask.GetCalls()
	var createCall *mocks.MockCall
	for i := range calls {
		if calls[i].Method == "CreateTask" {
			createCall = &calls[i]
			break
		}
	}
	require.NotNil(t, createCall, "CreateTask should have been called")
	assert.Equal(t, models.PriorityHigh, createCall.Args["priorityID"])
}
