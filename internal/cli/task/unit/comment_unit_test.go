package task_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func setupCommentMocks() (*mocks.MockTaskService, cli.MockServices) {
	mockTask := mocks.NewMockTaskService()
	services := cli.MockServices{
		TaskService: mockTask,
	}
	return mockTask, services
}

func setDefaultCommentResults(mockTask *mocks.MockTaskService) {
	mockTask.GetTaskDetailResult = &models.TaskDetail{
		ID:           1,
		Title:        "Test Task",
		TicketNumber: 42,
		ProjectName:  "Test Project",
	}
	mockTask.CreateCommentResult = &models.Comment{
		ID:        5,
		TaskID:    1,
		Message:   "test comment",
		Author:    "user",
		CreatedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}
}

func TestCommentTask_BasicCreate(t *testing.T) {
	t.Parallel()

	mockTask, services := setupCommentMocks()
	setDefaultCommentResults(mockTask)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"comment", "-i", "1", "-m", "test comment"})

	require.NoError(t, err)
	assert.Contains(t, output, "Comment added")
	assert.Contains(t, output, "#42")
	assert.Contains(t, output, "Test Task")
	assert.True(t, mockTask.HasCall("GetTaskDetail", 1))
	assert.Equal(t, 1, mockTask.CallCount("CreateComment"))

	calls := mockTask.GetCalls()
	var createCall *mocks.MockCall
	for i := range calls {
		if calls[i].Method == "CreateComment" {
			createCall = &calls[i]
			break
		}
	}
	require.NotNil(t, createCall, "CreateComment should have been called")
	assert.Equal(t, "test comment", createCall.Args["message"])
}

func TestCommentTask_Quiet(t *testing.T) {
	t.Parallel()

	mockTask, services := setupCommentMocks()
	setDefaultCommentResults(mockTask)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"comment", "-i", "1", "-m", "test comment", "-q"})

	require.NoError(t, err)
	assert.Equal(t, "5", strings.TrimSpace(output))
}

func TestCommentTask_JSON(t *testing.T) {
	t.Parallel()

	mockTask, services := setupCommentMocks()
	setDefaultCommentResults(mockTask)

	output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"comment", "-i", "1", "-m", "test comment", "-j"})

	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result),
		"output should be valid JSON: %s", output)

	assert.Equal(t, true, result["success"])

	comment, ok := result["comment"].(map[string]any)
	require.True(t, ok, "comment should be a map")
	assert.Equal(t, float64(5), comment["id"])
	assert.Equal(t, float64(1), comment["task_id"])
	assert.Equal(t, "test comment", comment["message"])
	assert.Equal(t, "user", comment["author"])

	taskData, ok := result["task"].(map[string]any)
	require.True(t, ok, "task should be a map")
	assert.Equal(t, float64(1), taskData["id"])
	assert.Equal(t, "Test Task", taskData["title"])
	assert.Equal(t, float64(42), taskData["ticket_number"])
	assert.Equal(t, "Test Project", taskData["project"])
}

func TestCommentTask_CustomAuthor(t *testing.T) {
	t.Parallel()

	mockTask, services := setupCommentMocks()
	setDefaultCommentResults(mockTask)

	_, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"comment", "-i", "1", "-m", "test comment", "-a", "noe"})

	require.NoError(t, err)

	calls := mockTask.GetCalls()
	var createCall *mocks.MockCall
	for i := range calls {
		if calls[i].Method == "CreateComment" {
			createCall = &calls[i]
			break
		}
	}
	require.NotNil(t, createCall, "CreateComment should have been called")
	assert.Equal(t, "noe", createCall.Args["author"])
}

func TestCommentTask_TaskNotFound(t *testing.T) {
	t.Parallel()

	mockTask, services := setupCommentMocks()
	mockTask.GetTaskDetailErr = fmt.Errorf("not found")

	_, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"comment", "-i", "999", "-m", "test comment"})

	require.Error(t, err)
	cli.AssertExitError(t, err, rootcli.ExitError)
	assert.Contains(t, err.Error(), "not found")
	assert.Equal(t, 0, mockTask.CallCount("CreateComment"))
}

func TestCommentTask_MissingTaskID(t *testing.T) {
	t.Parallel()

	_, services := setupCommentMocks()

	_, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"comment", "-m", "test comment"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestCommentTask_MissingMessage(t *testing.T) {
	t.Parallel()

	_, services := setupCommentMocks()

	_, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(),
		[]string{"comment", "-i", "1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}
