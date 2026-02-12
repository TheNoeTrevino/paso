package task_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func intPtr(v int) *int { return &v }

func setupMoveColumns() []*models.Column {
	return []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1, PrevID: nil, NextID: intPtr(11)},
		{ID: 11, Name: "In Progress", ProjectID: 1, PrevID: intPtr(10), NextID: intPtr(12)},
		{ID: 12, Name: "Done", ProjectID: 1, PrevID: intPtr(11), NextID: nil},
	}
}

func TestMoveTask(t *testing.T) {
	t.Parallel()

	t.Run("move to next column", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockColumn := mocks.NewMockColumnService()

		mockTask.GetTaskDetailResult = &models.TaskDetail{ID: 1, Title: "Test Task", ColumnID: 10}
		mockColumn.GetColumnByIDResult = &models.Column{ID: 10, Name: "Todo", ProjectID: 1}
		mockColumn.GetColumnsByProjectResult = setupMoveColumns()

		cmd := task.MoveCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:   mockTask,
			ColumnService: mockColumn,
		}, cmd, []string{"--id", "1", "next"})

		require.NoError(t, err)
		assert.Contains(t, output, "Task 1 moved to 'In Progress'")
		assert.True(t, mockTask.HasCall("MoveTaskToNextColumn", 1))
	})

	t.Run("move to specific column by name", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockColumn := mocks.NewMockColumnService()

		mockTask.GetTaskDetailResult = &models.TaskDetail{ID: 1, Title: "Test Task", ColumnID: 10}
		mockColumn.GetColumnByIDResult = &models.Column{ID: 10, Name: "Todo", ProjectID: 1}
		mockColumn.GetColumnsByProjectResult = setupMoveColumns()

		cmd := task.MoveCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:   mockTask,
			ColumnService: mockColumn,
		}, cmd, []string{"--id", "1", "Done"})

		require.NoError(t, err)
		assert.Contains(t, output, "Task 1 moved to 'Done'")
		assert.True(t, mockTask.HasCall("MoveTaskToColumn", 1))
	})

	t.Run("move with --quiet", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockColumn := mocks.NewMockColumnService()

		mockTask.GetTaskDetailResult = &models.TaskDetail{ID: 1, Title: "Test Task", ColumnID: 10}
		mockColumn.GetColumnByIDResult = &models.Column{ID: 10, Name: "Todo", ProjectID: 1}
		mockColumn.GetColumnsByProjectResult = setupMoveColumns()

		cmd := task.MoveCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:   mockTask,
			ColumnService: mockColumn,
		}, cmd, []string{"--id", "1", "next", "--quiet"})

		require.NoError(t, err)
		assert.Equal(t, "1\n", output)
	})

	t.Run("move with --json", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockColumn := mocks.NewMockColumnService()

		mockTask.GetTaskDetailResult = &models.TaskDetail{ID: 1, Title: "Test Task", ColumnID: 10}
		mockColumn.GetColumnByIDResult = &models.Column{ID: 10, Name: "Todo", ProjectID: 1}
		mockColumn.GetColumnsByProjectResult = setupMoveColumns()

		cmd := task.MoveCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:   mockTask,
			ColumnService: mockColumn,
		}, cmd, []string{"--id", "1", "next", "--json"})

		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(1), result["task_id"])
		assert.Equal(t, "Todo", result["from_column"])
		assert.Equal(t, "In Progress", result["to_column"])
	})

	t.Run("task not found", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockColumn := mocks.NewMockColumnService()

		mockTask.GetTaskDetailErr = fmt.Errorf("not found")

		cmd := task.MoveCmd()
		_, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:   mockTask,
			ColumnService: mockColumn,
		}, cmd, []string{"--id", "999", "next"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "task 999 not found")
	})
}

func TestDoneTask(t *testing.T) {
	t.Parallel()

	t.Run("mark task done", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		callCount := 0
		mockTask.GetTaskDetailFunc = func(_ context.Context, _ int) (*models.TaskDetail, error) {
			callCount++
			if callCount == 1 {
				return &models.TaskDetail{ID: 42, Title: "Test Task", ColumnID: 10, ColumnName: "In Progress"}, nil
			}
			return &models.TaskDetail{ID: 42, Title: "Test Task", ColumnID: 12, ColumnName: "Done"}, nil
		}

		cmd := task.DoneCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService: mockTask,
		}, cmd, []string{"42"})

		require.NoError(t, err)
		assert.Contains(t, output, "Task 42 moved to 'Done'")
		assert.True(t, mockTask.HasCall("MoveTaskToCompletedColumn", 42))
	})

	t.Run("done with --quiet", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		callCount := 0
		mockTask.GetTaskDetailFunc = func(_ context.Context, _ int) (*models.TaskDetail, error) {
			callCount++
			if callCount == 1 {
				return &models.TaskDetail{ID: 42, Title: "Test Task", ColumnID: 10, ColumnName: "In Progress"}, nil
			}
			return &models.TaskDetail{ID: 42, Title: "Test Task", ColumnID: 12, ColumnName: "Done"}, nil
		}

		cmd := task.DoneCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService: mockTask,
		}, cmd, []string{"42", "--quiet"})

		require.NoError(t, err)
		assert.Equal(t, "42\n", output)
	})

	t.Run("done with --json", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		callCount := 0
		mockTask.GetTaskDetailFunc = func(_ context.Context, _ int) (*models.TaskDetail, error) {
			callCount++
			if callCount == 1 {
				return &models.TaskDetail{ID: 42, Title: "Test Task", ColumnID: 10, ColumnName: "In Progress"}, nil
			}
			return &models.TaskDetail{ID: 42, Title: "Test Task", ColumnID: 12, ColumnName: "Done"}, nil
		}

		cmd := task.DoneCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService: mockTask,
		}, cmd, []string{"42", "--json"})

		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(42), result["task_id"])
		assert.Equal(t, "In Progress", result["from_column"])
		assert.Equal(t, "Done", result["to_column"])
	})

	t.Run("task not found", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockTask.GetTaskDetailErr = fmt.Errorf("not found")

		cmd := task.DoneCmd()
		_, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService: mockTask,
		}, cmd, []string{"999"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "task 999 not found")
	})
}

func TestReadyTask(t *testing.T) {
	t.Parallel()

	t.Run("list ready tasks", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockProject := mocks.NewMockProjectService()

		mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
		mockTask.GetReadyTaskSummariesByProjectResult = []*models.TaskSummary{
			{ID: 10, Title: "Ready Task 1"},
			{ID: 20, Title: "Ready Task 2"},
		}

		cmd := task.ReadyCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:    mockTask,
			ProjectService: mockProject,
		}, cmd, []string{"--project", "1"})

		require.NoError(t, err)
		assert.Contains(t, output, "Found 2 ready tasks")
		assert.Contains(t, output, "[10] Ready Task 1")
		assert.Contains(t, output, "[20] Ready Task 2")
	})

	t.Run("no ready tasks", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockProject := mocks.NewMockProjectService()

		mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
		mockTask.GetReadyTaskSummariesByProjectResult = []*models.TaskSummary{}

		cmd := task.ReadyCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:    mockTask,
			ProjectService: mockProject,
		}, cmd, []string{"--project", "1"})

		require.NoError(t, err)
		assert.Contains(t, output, "No ready tasks found")
	})

	t.Run("ready with --json", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockProject := mocks.NewMockProjectService()

		mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
		mockTask.GetReadyTaskSummariesByProjectResult = []*models.TaskSummary{
			{ID: 10, Title: "Ready Task 1"},
		}

		cmd := task.ReadyCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:    mockTask,
			ProjectService: mockProject,
		}, cmd, []string{"--project", "1", "--json"})

		require.NoError(t, err)

		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.Equal(t, true, result["success"])
		assert.Equal(t, float64(1), result["count"])

		tasks, ok := result["tasks"].([]any)
		require.True(t, ok)
		require.Len(t, tasks, 1)
		taskMap := tasks[0].(map[string]any)
		assert.Equal(t, float64(10), taskMap["ID"])
		assert.Equal(t, "Ready Task 1", taskMap["Title"])
	})

	t.Run("ready with --quiet", func(t *testing.T) {
		t.Parallel()
		t.Parallel()

		mockTask := mocks.NewMockTaskService()
		mockProject := mocks.NewMockProjectService()

		mockProject.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
		mockTask.GetReadyTaskSummariesByProjectResult = []*models.TaskSummary{
			{ID: 10, Title: "Ready Task 1"},
			{ID: 20, Title: "Ready Task 2"},
		}

		cmd := task.ReadyCmd()
		output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
			TaskService:    mockTask,
			ProjectService: mockProject,
		}, cmd, []string{"--project", "1", "--quiet"})

		require.NoError(t, err)
		assert.Equal(t, "10\n20\n", output)
	})
}
