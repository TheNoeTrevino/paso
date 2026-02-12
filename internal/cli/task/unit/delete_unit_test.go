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

func TestDeleteTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		setupMock    func(*mocks.MockTaskService)
		wantErr      bool
		wantCode     int
		wantContains []string
		wantEmpty    bool
		wantJSON     map[string]any
		checkMock    func(*testing.T, *mocks.MockTaskService)
	}{
		{
			name: "Delete with --force",
			args: []string{"delete", "1", "-f"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailResult = &models.TaskDetail{ID: 1, Title: "Test Task"}
			},
			wantContains: []string{"deleted successfully"},
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.True(t, m.HasCall("GetTaskDetail", 1))
				assert.True(t, m.HasCall("DeleteTask", 1))
			},
		},
		{
			name: "Delete with --quiet",
			args: []string{"delete", "2", "-q"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailResult = &models.TaskDetail{ID: 2, Title: "Quiet Task"}
			},
			wantEmpty: true,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.True(t, m.HasCall("GetTaskDetail", 2))
				assert.True(t, m.HasCall("DeleteTask", 2))
			},
		},
		{
			name: "Delete with --json",
			args: []string{"delete", "3", "-j", "-f"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailResult = &models.TaskDetail{ID: 3, Title: "JSON Task"}
			},
			wantJSON: map[string]any{
				"success": true,
				"task_id": float64(3),
			},
		},
		{
			name: "Task not found",
			args: []string{"delete", "999", "-f"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailErr = fmt.Errorf("not found")
			},
			wantErr:  true,
			wantCode: rootcli.ExitNotFound,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.True(t, m.HasCall("GetTaskDetail", 999))
				assert.Equal(t, 0, m.CallCount("DeleteTask"), "DeleteTask should not be called when task not found")
			},
		},
		{
			name: "Delete service error",
			args: []string{"delete", "5", "-f"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskDetailResult = &models.TaskDetail{ID: 5, Title: "Error Task"}
				m.DeleteTaskErr = fmt.Errorf("database connection lost")
			},
			wantErr:  true,
			wantCode: rootcli.ExitError,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.True(t, m.HasCall("DeleteTask", 5))
			},
		},
		{
			name:    "Missing task ID",
			args:    []string{"delete"},
			wantErr: true,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 0, m.CallCount("GetTaskDetail"), "no service calls should be made")
			},
		},
		{
			name:     "Invalid task ID",
			args:     []string{"delete", "abc", "-f"},
			wantErr:  true,
			wantCode: rootcli.ExitValidation,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 0, m.CallCount("GetTaskDetail"), "no service calls should be made for invalid ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Parallel()

			mockTask := mocks.NewMockTaskService()
			if tt.setupMock != nil {
				tt.setupMock(mockTask)
			}

			output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
				TaskService: mockTask,
			}, task.TaskCmd(), tt.args)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantCode != 0 {
					cli.AssertExitError(t, err, tt.wantCode)
				}
			} else {
				assert.NoError(t, err)
			}

			for _, s := range tt.wantContains {
				assert.Contains(t, output, s)
			}

			if tt.wantEmpty {
				assert.Empty(t, output, "quiet mode should produce no output")
			}

			if tt.wantJSON != nil {
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result), "output should be valid JSON")
				for k, v := range tt.wantJSON {
					assert.Equal(t, v, result[k])
				}
			}

			if tt.checkMock != nil {
				tt.checkMock(t, mockTask)
			}
		})
	}
}
