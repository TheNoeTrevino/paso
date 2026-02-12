package column_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rootcli "github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/column"
	"github.com/thenoetrevino/paso/internal/models"
	cli "github.com/thenoetrevino/paso/internal/testing/cli"
	"github.com/thenoetrevino/paso/internal/testing/mocks"
)

func TestCreateColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		setupMocks   func(*mocks.MockProjectService, *mocks.MockColumnService)
		wantErr      bool
		wantContains []string
		wantErrMsg   string
		wantJSON     map[string]any
		assertOutput func(*testing.T, string)
		checkMock    func(*testing.T, *mocks.MockProjectService, *mocks.MockColumnService)
	}{
		{
			name: "Create column",
			args: []string{"create", "-n", "Review", "-p", "1"},
			setupMocks: func(mp *mocks.MockProjectService, mc *mocks.MockColumnService) {
				mp.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
				mc.CreateColumnResult = &models.Column{ID: 10, Name: "Review", ProjectID: 1}
			},
			wantContains: []string{"10", "Review"},
			checkMock: func(t *testing.T, mp *mocks.MockProjectService, mc *mocks.MockColumnService) {
				t.Helper()
				assert.True(t, mp.HasCall("GetProjectByID"))
				assert.Equal(t, 1, mc.CallCount("CreateColumn"))
			},
		},
		{
			name: "Create column with --json",
			args: []string{"create", "-n", "Done", "-p", "1", "-j"},
			setupMocks: func(mp *mocks.MockProjectService, mc *mocks.MockColumnService) {
				mp.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
				mc.CreateColumnResult = &models.Column{ID: 15, Name: "Done", ProjectID: 1}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result), "output should be valid JSON: %s", output)
				assert.Equal(t, true, result["success"])
				col, ok := result["column"].(map[string]any)
				require.True(t, ok, "column should be a map")
				assert.Equal(t, float64(15), col["id"])
				assert.Equal(t, "Done", col["name"])
				assert.Equal(t, float64(1), col["project_id"])
			},
		},
		{
			name:       "Create column missing name",
			args:       []string{"create", "-p", "1"},
			wantErr:    true,
			wantErrMsg: "required",
			checkMock: func(t *testing.T, mp *mocks.MockProjectService, mc *mocks.MockColumnService) {
				t.Helper()
				assert.Equal(t, 0, mc.CallCount("CreateColumn"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockProject := mocks.NewMockProjectService()
			mockColumn := mocks.NewMockColumnService()
			if tt.setupMocks != nil {
				tt.setupMocks(mockProject, mockColumn)
			}

			output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
				ProjectService: mockProject,
				ColumnService:  mockColumn,
			}, column.ColumnCmd(), tt.args)

			if tt.wantErr {
				require.Error(t, err)
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
				tt.checkMock(t, mockProject, mockColumn)
			}
		})
	}
}

func TestListColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		setupMocks   func(*mocks.MockProjectService, *mocks.MockColumnService)
		wantContains []string
		assertOutput func(*testing.T, string)
		checkMock    func(*testing.T, *mocks.MockProjectService, *mocks.MockColumnService)
	}{
		{
			name: "List columns",
			args: []string{"list", "1"},
			setupMocks: func(mp *mocks.MockProjectService, mc *mocks.MockColumnService) {
				mp.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
				mc.GetColumnsByProjectResult = []*models.Column{
					{ID: 1, Name: "Todo", ProjectID: 1},
					{ID: 2, Name: "In Progress", ProjectID: 1},
					{ID: 3, Name: "Done", ProjectID: 1},
				}
			},
			wantContains: []string{"Todo", "In Progress", "Done"},
			checkMock: func(t *testing.T, mp *mocks.MockProjectService, mc *mocks.MockColumnService) {
				t.Helper()
				assert.True(t, mp.HasCall("GetProjectByID"))
				assert.Equal(t, 1, mc.CallCount("GetColumnsByProject"))
			},
		},
		{
			name: "List columns --json",
			args: []string{"list", "1", "-j"},
			setupMocks: func(mp *mocks.MockProjectService, mc *mocks.MockColumnService) {
				mp.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
				mc.GetColumnsByProjectResult = []*models.Column{
					{ID: 10, Name: "Backlog", ProjectID: 1, HoldsReadyTasks: true},
					{ID: 11, Name: "Active", ProjectID: 1, HoldsInProgressTasks: true},
				}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result))
				assert.Equal(t, true, result["success"])

				columns, ok := result["columns"].([]any)
				require.True(t, ok)
				require.Len(t, columns, 2)

				first := columns[0].(map[string]any)
				assert.Equal(t, float64(10), first["id"])
				assert.Equal(t, "Backlog", first["name"])
				assert.Equal(t, true, first["holds_ready_tasks"])

				second := columns[1].(map[string]any)
				assert.Equal(t, float64(11), second["id"])
				assert.Equal(t, "Active", second["name"])
				assert.Equal(t, true, second["holds_in_progress_tasks"])
			},
		},
		{
			name: "List columns --quiet",
			args: []string{"list", "1", "-q"},
			setupMocks: func(mp *mocks.MockProjectService, mc *mocks.MockColumnService) {
				mp.GetProjectByIDResult = &models.Project{ID: 1, Name: "Test Project"}
				mc.GetColumnsByProjectResult = []*models.Column{
					{ID: 5, Name: "Todo", ProjectID: 1},
					{ID: 12, Name: "Done", ProjectID: 1},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockProject := mocks.NewMockProjectService()
			mockColumn := mocks.NewMockColumnService()
			if tt.setupMocks != nil {
				tt.setupMocks(mockProject, mockColumn)
			}

			output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
				ProjectService: mockProject,
				ColumnService:  mockColumn,
			}, column.ColumnCmd(), tt.args)

			require.NoError(t, err)

			for _, s := range tt.wantContains {
				assert.Contains(t, output, s)
			}

			if tt.assertOutput != nil {
				tt.assertOutput(t, output)
			}

			if tt.checkMock != nil {
				tt.checkMock(t, mockProject, mockColumn)
			}
		})
	}
}

func TestUpdateColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		setupMock    func(*mocks.MockColumnService)
		wantErr      bool
		wantCode     int
		wantContains []string
		wantJSON     map[string]any
		assertOutput func(*testing.T, string)
		checkMock    func(*testing.T, *mocks.MockColumnService)
	}{
		{
			name: "Update column name",
			args: []string{"update", "1", "-n", "New Name"},
			setupMock: func(m *mocks.MockColumnService) {
				m.GetColumnByIDResult = &models.Column{ID: 1, Name: "Old Name", ProjectID: 1}
			},
			wantContains: []string{"updated successfully"},
			checkMock: func(t *testing.T, m *mocks.MockColumnService) {
				t.Helper()
				assert.True(t, m.HasCall("GetColumnByID"))
				assert.Equal(t, 1, m.CallCount("UpdateColumnName"))
			},
		},
		{
			name: "Update non-existent column",
			args: []string{"update", "999", "-n", "Nope"},
			setupMock: func(m *mocks.MockColumnService) {
				m.GetColumnByIDErr = fmt.Errorf("not found")
			},
			wantErr:  true,
			wantCode: rootcli.ExitNotFound,
			checkMock: func(t *testing.T, m *mocks.MockColumnService) {
				t.Helper()
				assert.True(t, m.HasCall("GetColumnByID"))
				assert.Equal(t, 0, m.CallCount("UpdateColumnName"))
			},
		},
		{
			name: "Update column with --json",
			args: []string{"update", "1", "-n", "Renamed", "-j"},
			setupMock: func(m *mocks.MockColumnService) {
				m.GetColumnByIDResult = &models.Column{ID: 1, Name: "Old Name", ProjectID: 1}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result))
				assert.Equal(t, true, result["success"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockColumn := mocks.NewMockColumnService()
			if tt.setupMock != nil {
				tt.setupMock(mockColumn)
			}

			output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
				ColumnService: mockColumn,
			}, column.ColumnCmd(), tt.args)

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

			if tt.assertOutput != nil {
				tt.assertOutput(t, output)
			}

			if tt.checkMock != nil {
				tt.checkMock(t, mockColumn)
			}
		})
	}
}

func TestDeleteColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		setupMocks   func(*mocks.MockColumnService, *mocks.MockTaskService)
		wantErr      bool
		wantCode     int
		wantContains []string
		assertOutput func(*testing.T, string)
		checkMock    func(*testing.T, *mocks.MockColumnService)
	}{
		{
			name: "Delete column with --force",
			args: []string{"delete", "7", "-f"},
			setupMocks: func(mc *mocks.MockColumnService, mt *mocks.MockTaskService) {
				mc.GetColumnByIDResult = &models.Column{ID: 7, Name: "Archive", ProjectID: 1}
				mt.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{}
			},
			wantContains: []string{"deleted successfully"},
			checkMock: func(t *testing.T, mc *mocks.MockColumnService) {
				t.Helper()
				assert.True(t, mc.HasCall("GetColumnByID"))
				assert.Equal(t, 1, mc.CallCount("DeleteColumn"))
			},
		},
		{
			name: "Delete column --json",
			args: []string{"delete", "3", "-f", "-j"},
			setupMocks: func(mc *mocks.MockColumnService, mt *mocks.MockTaskService) {
				mc.GetColumnByIDResult = &models.Column{ID: 3, Name: "Review", ProjectID: 1}
				mt.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{}
			},
			assertOutput: func(t *testing.T, output string) {
				t.Helper()
				var result map[string]any
				require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &result))
				assert.Equal(t, true, result["success"])
				assert.Equal(t, float64(3), result["column_id"])
			},
		},
		{
			name: "Delete non-existent column",
			args: []string{"delete", "999", "-f"},
			setupMocks: func(mc *mocks.MockColumnService, mt *mocks.MockTaskService) {
				mc.GetColumnByIDErr = fmt.Errorf("not found")
			},
			wantErr:  true,
			wantCode: rootcli.ExitNotFound,
			checkMock: func(t *testing.T, mc *mocks.MockColumnService) {
				t.Helper()
				assert.True(t, mc.HasCall("GetColumnByID"))
				assert.Equal(t, 0, mc.CallCount("DeleteColumn"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockColumn := mocks.NewMockColumnService()
			mockTask := mocks.NewMockTaskService()
			if tt.setupMocks != nil {
				tt.setupMocks(mockColumn, mockTask)
			}

			output, err := cli.ExecuteCLICommandWithMocks(t, cli.MockServices{
				ColumnService: mockColumn,
				TaskService:   mockTask,
			}, column.ColumnCmd(), tt.args)

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

			if tt.assertOutput != nil {
				tt.assertOutput(t, output)
			}

			if tt.checkMock != nil {
				tt.checkMock(t, mockColumn)
			}
		})
	}
}
