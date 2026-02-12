package task_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/app"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/task"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func TestListTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		setupMock    func(*mocks.MockTaskService)
		setupApp     func(*mocks.MockTaskService) *app.App
		wantErr      bool
		wantCode     int
		wantContains []string
		wantErrMsg   string
		assertOutput func(*testing.T, string)
		checkMock    func(*testing.T, *mocks.MockTaskService)
	}{
		{
			name: "list tasks with results",
			args: []string{"1"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
					1: {
						{ID: 1, Title: "Task 1", PriorityDescription: "medium", TypeDescription: "task"},
						{ID: 2, Title: "Task 2", PriorityDescription: "high", TypeDescription: "feature"},
					},
				}
			},
			wantContains: []string{"Found 2 tasks", "Task 1", "Task 2"},
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 1, m.CallCount("GetTaskSummariesByProject"))
			},
		},
		{
			name: "list empty project",
			args: []string{"1"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{}
			},
			wantContains: []string{"No tasks found"},
		},
		{
			name: "list with --json",
			args: []string{"1", "--json"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
					1: {
						{ID: 10, Title: "JSON Task", PriorityDescription: "low", TypeDescription: "bug"},
					},
				}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(output), &result))
				assert.True(t, result["success"].(bool))
				tasks := result["tasks"].([]any)
				assert.Len(t, tasks, 1)
				taskData := tasks[0].(map[string]any)
				assert.Equal(t, float64(10), taskData["ID"])
				assert.Equal(t, "JSON Task", taskData["Title"])
			},
		},
		{
			name: "list with --quiet",
			args: []string{"1", "--quiet"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
					1: {
						{ID: 5, Title: "Quiet Task 1"},
						{ID: 12, Title: "Quiet Task 2"},
					},
				}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				lines := strings.Split(strings.TrimSpace(output), "\n")
				require.Len(t, lines, 2)
				assert.Equal(t, "5", lines[0])
				assert.Equal(t, "12", lines[1])
			},
		},
		{
			name: "missing project ID errors",
			args: []string{},
			setupApp: func(m *mocks.MockTaskService) *app.App {
				mockGit := mocks.NewMockGitDetector()
				return &app.App{
					TaskService: m,
					GitDetector: mockGit,
				}
			},
			wantErr:  true,
			wantCode: rootcli.ExitUsage,
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 0, m.CallCount("GetTaskSummariesByProject"))
			},
		},
		{
			name:       "invalid project ID",
			args:       []string{"abc"},
			wantErr:    true,
			wantCode:   rootcli.ExitUsage,
			wantErrMsg: "Invalid project ID: abc",
			checkMock: func(t *testing.T, m *mocks.MockTaskService) {
				t.Helper()
				assert.Equal(t, 0, m.CallCount("GetTaskSummariesByProject"))
			},
		},
		{
			name: "service error propagates",
			args: []string{"1"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectErr = assert.AnError
			},
			wantErr:  true,
			wantCode: rootcli.ExitError,
		},
		{
			name: "project flag works as alternative to positional arg",
			args: []string{"--project", "7", "--json"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
					1: {{ID: 3, Title: "Flag Task"}},
				}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(output), &result))
				assert.True(t, result["success"].(bool))
			},
		},
		{
			name: "json output for empty project",
			args: []string{"1", "--json"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(output), &result))
				assert.True(t, result["success"].(bool))
				tasks := result["tasks"]
				if tasks != nil {
					assert.Empty(t, tasks.([]any))
				}
			},
		},
		{
			name: "quiet output for empty project",
			args: []string{"1", "--quiet"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				assert.Empty(t, strings.TrimSpace(output))
			},
		},
		{
			name: "tasks across multiple columns are flattened",
			args: []string{"1", "--quiet"},
			setupMock: func(m *mocks.MockTaskService) {
				m.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
					1: {{ID: 1, Title: "Todo Task"}},
					2: {{ID: 2, Title: "In Progress Task"}},
					3: {{ID: 3, Title: "Done Task"}},
				}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				lines := strings.Split(strings.TrimSpace(output), "\n")
				assert.Len(t, lines, 3)
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

			cmd := task.ListCmd()

			var output string
			var err error

			if tt.setupApp != nil {
				testApp := tt.setupApp(mockTask)
				output, err = cli.ExecuteCLICommand(t, testApp, cmd, tt.args)
			} else {
				output, err = cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
					TaskService: mockTask,
				}, cmd, tt.args)
			}

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantCode != 0 {
					cli.AssertExitError(t, err, tt.wantCode)
				}
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
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
