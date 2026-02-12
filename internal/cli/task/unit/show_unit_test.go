package task_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func setupShowMocks() (*mocks.MockTaskService, cli.MockServices) {
	mockTask := mocks.NewMockTaskService()
	services := cli.MockServices{
		TaskService: mockTask,
	}
	return mockTask, services
}

func defaultShowTask() *models.TaskDetail {
	return &models.TaskDetail{
		ID:                  1,
		Title:               "Test Task",
		Description:         "desc",
		ColumnID:            10,
		ColumnName:          "Todo",
		ProjectName:         "Test Project",
		TicketNumber:        42,
		TypeDescription:     "task",
		PriorityDescription: "medium",
		PriorityColor:       "#F59E0B",
	}
}

func TestShowTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		setupMock    func(*mocks.MockTaskService)
		wantErr      bool
		wantCode     int
		wantContains []string
		wantErrMsg   string
		assertOutput func(*testing.T, string)
		checkMock    func(*testing.T, *mocks.MockTaskService)
	}{
		{
			name: "positional arg",
			args: []string{"show", "1"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailResult = defaultShowTask()
			},
			wantContains: []string{"Test Task"},
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.True(t, m.HasCall("GetTaskDetail", 1))
			},
		},
		{
			name: "ID flag",
			args: []string{"show", "-i", "1"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailResult = defaultShowTask()
			},
			wantContains: []string{"Test Task"},
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.True(t, m.HasCall("GetTaskDetail", 1))
			},
		},
		{
			name: "JSON output",
			args: []string{"show", "1", "-j"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailResult = defaultShowTask()
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result),
					"output should be valid JSON: %s", output)

				assert.True(t, result["success"].(bool))
				taskData := result["task"].(map[string]any)
				assert.Equal(t, float64(1), taskData["id"])
				assert.Equal(t, "Test Task", taskData["title"])
				assert.Equal(t, "desc", taskData["description"])
				assert.Equal(t, "Test Project", taskData["project_name"])
				assert.Equal(t, float64(42), taskData["ticket_number"])

				column := taskData["column"].(map[string]any)
				assert.Equal(t, float64(10), column["id"])
				assert.Equal(t, "Todo", column["name"])

				priority := taskData["priority"].(map[string]any)
				assert.Equal(t, "medium", priority["name"])
				assert.Equal(t, "#F59E0B", priority["color"])
			},
		},
		{
			name: "quiet output",
			args: []string{"show", "1", "-q"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailResult = defaultShowTask()
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				assert.Equal(t, "1", strings.TrimSpace(output))
			},
		},
		{
			name: "task not found",
			args: []string{"show", "999"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailErr = fmt.Errorf("not found")
			},
			wantErr:    true,
			wantCode:   rootcli.ExitNotFound,
			wantErrMsg: "task 999 not found",
		},
		{
			name:       "missing task ID",
			args:       []string{"show"},
			wantErr:    true,
			wantCode:   rootcli.ExitUsage,
			wantErrMsg: "task ID must be a positive integer",
		},
		{
			name:       "invalid task ID",
			args:       []string{"show", "abc"},
			wantErr:    true,
			wantCode:   rootcli.ExitUsage,
			wantErrMsg: "task ID must be a positive integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
			t.Parallel()

			mockTask, services := setupShowMocks()
			if tt.setupMock != nil {
				tt.setupMock(mockTask)
			}

			output, err := cli.ExecuteCLICommandWithMocks(t, services, task.TaskCmd(), tt.args)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantCode != 0 {
					cli.AssertExitError(t, err, tt.wantCode)
				}
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
			}

			for _, s := range tt.wantContains {
				assert.Contains(t, output, s)
			}

			if tt.assertOutput != nil {
				tt.assertOutput(t, output)
			}

			if tt.checkMock != nil {
				tt.checkMock(t, mockTask)
			}
		})
	}
}
