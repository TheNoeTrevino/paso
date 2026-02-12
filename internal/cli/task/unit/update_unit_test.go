package task_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/task"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func TestUpdateTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		setupMock    func(*mocks.MockTaskService)
		wantErr      bool
		wantCode     int
		wantContains []string
		wantExact    string
		wantJSON     map[string]any
		checkMock    func(*testing.T, *mocks.MockTaskService)
	}{
		{
			name:         "Update title",
			args:         []string{"update", "1", "-t", "Updated Title"},
			wantContains: []string{"updated successfully"},
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 1, m.CallCount("UpdateTask"))
				assert.True(t, m.HasCall("UpdateTask", 1))
			},
		},
		{
			name:         "Update description",
			args:         []string{"update", "42", "-d", "New description text"},
			wantContains: []string{"updated successfully"},
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 1, m.CallCount("UpdateTask"))
				assert.True(t, m.HasCall("UpdateTask", 42))
			},
		},
		{
			name:         "Update priority",
			args:         []string{"update", "5", "-r", "high"},
			wantContains: []string{"updated successfully"},
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 1, m.CallCount("UpdateTask"))
				assert.True(t, m.HasCall("UpdateTask", 5))
			},
		},
		{
			name:      "Update with quiet flag",
			args:      []string{"update", "7", "-t", "Quiet Update", "-q"},
			wantExact: "7\n",
		},
		{
			name: "Update with json flag",
			args: []string{"update", "3", "-t", "JSON Update", "-j"},
			wantJSON: map[string]any{
				"success": true,
				"task_id": float64(3),
			},
		},
		{
			name:    "Missing task ID",
			args:    []string{"update"},
			wantErr: true,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 0, m.CallCount("UpdateTask"))
			},
		},
		{
			name:     "Invalid task ID",
			args:     []string{"update", "abc", "-t", "Title"},
			wantErr:  true,
			wantCode: rootcli.ExitValidation,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 0, m.CallCount("UpdateTask"))
			},
		},
		{
			name: "Service error on UpdateTask",
			args: []string{"update", "1", "-t", "Fail"},
			setupMock: func(m *mocks.MockTaskService) {
				m.UpdateTaskErr = errors.New("database connection lost")
			},
			wantErr:  true,
			wantCode: rootcli.ExitError,
		},
		{
			name:     "No flags provided",
			args:     []string{"update", "1"},
			wantErr:  true,
			wantCode: rootcli.ExitUsage,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 0, m.CallCount("UpdateTask"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				require.NoError(t, err)
			}

			for _, s := range tt.wantContains {
				assert.Contains(t, output, s)
			}

			if tt.wantExact != "" {
				assert.Equal(t, tt.wantExact, output)
			}

			if tt.wantJSON != nil {
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(output), &result))
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
