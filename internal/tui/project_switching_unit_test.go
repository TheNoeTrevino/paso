package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestUnit_ProjectSwitch_NextProjectUpdatesActiveProject verifies that
// pressing the NextProject key ('}') switches to the next project and
// updates the selected project index.
func TestUnit_ProjectSwitch_NextProjectUpdatesActiveProject(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
		{ID: 3, Name: "Gamma"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Verify initial state: first project is active
	require.Equal(t, 0, m.AppState.SelectedProject())
	currentProject := m.AppState.GetCurrentProject()
	require.NotNil(t, currentProject)
	require.Equal(t, 1, currentProject.ID)

	// Press '}' (default NextProject key)
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "}", Code: '}'})
	updated := UpdateModelWithMessage(m, keyMsg)

	// Should have moved to the second project
	assert.Equal(t, 1, updated.AppState.SelectedProject(),
		"SelectedProject index should be 1 after pressing NextProject")
	newProject := updated.AppState.GetCurrentProject()
	require.NotNil(t, newProject)
	assert.Equal(t, 2, newProject.ID,
		"active project ID should be 2 after switching to Beta")
}

// TestUnit_ProjectSwitch_ReloadsColumnsAndTasks verifies that switching
// projects calls GetColumnsByProject and GetTaskSummariesByProject for
// the new project.
func TestUnit_ProjectSwitch_ReloadsColumnsAndTasks(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Record initial call counts (init already called these)
	initialColumnCalls := svc.Column.CallCount("GetColumnsByProject")
	initialTaskCalls := svc.Task.CallCount("GetTaskSummariesByProject")
	initialLabelCalls := svc.Label.CallCount("GetLabelsByProject")

	// Now set up what the second project should return
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 20, Name: "Backlog", ProjectID: 2},
		{ID: 21, Name: "Done", ProjectID: 2},
	}
	svc.Task.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
		20: {{ID: 100, Title: "Beta task", ColumnID: 20, Position: 0, Labels: []*models.Label{}, PriorityColor: "#00FF00"}},
		21: {},
	}
	svc.Label.GetLabelsByProjectResult = []*models.Label{
		{ID: 5, Name: "urgent", Color: "#FF0000"},
	}

	// Press '}' to switch to Beta
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "}", Code: '}'})
	updated := UpdateModelWithMessage(m, keyMsg)

	// Verify service calls were made for the new project
	assert.Greater(t, svc.Column.CallCount("GetColumnsByProject"), initialColumnCalls,
		"GetColumnsByProject should be called again after project switch")
	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProject"), initialTaskCalls,
		"GetTaskSummariesByProject should be called again after project switch")
	assert.Greater(t, svc.Label.CallCount("GetLabelsByProject"), initialLabelCalls,
		"GetLabelsByProject should be called again after project switch")

	// Verify the model state was updated with new project's data
	assert.Len(t, updated.AppState.Columns(), 2,
		"columns should be updated to Beta's columns")
	assert.Equal(t, "Backlog", updated.AppState.Columns()[0].Name)
}

// TestUnit_ProjectSwitch_ResetsSelection verifies that switching projects
// resets column and task selection to zero.
func TestUnit_ProjectSwitch_ResetsSelection(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Col1", ProjectID: 1},
		{ID: 11, Name: "Col2", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
		10: {
			{ID: 1, Title: "Task A", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"},
			{ID: 2, Title: "Task B", ColumnID: 10, Position: 1, Labels: []*models.Label{}, PriorityColor: "#FF0000"},
		},
		11: {},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Simulate the user having navigated to column 1, task 1
	m.UIState.SelectedColumn = 1
	m.UIState.SelectedTask = 1

	// Switch to next project
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "}", Code: '}'})
	updated := UpdateModelWithMessage(m, keyMsg)

	assert.Equal(t, 0, updated.UIState.SelectedColumn,
		"SelectedColumn should reset to 0 after project switch")
	assert.Equal(t, 0, updated.UIState.SelectedTask,
		"SelectedTask should reset to 0 after project switch")
}

// TestUnit_ProjectSwitch_PrevProjectAtFirstStaysInPlace verifies that
// pressing PrevProject ('{') when already at the first project does not
// change the selection.
func TestUnit_ProjectSwitch_PrevProjectAtFirstStaysInPlace(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	require.Equal(t, 0, m.AppState.SelectedProject())

	// Press '{' (default PrevProject key) when already at first
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "{", Code: '{'})
	updated := UpdateModelWithMessage(m, keyMsg)

	assert.Equal(t, 0, updated.AppState.SelectedProject(),
		"SelectedProject should remain 0 when already at first project")
	assert.Equal(t, state.NormalMode, updated.UIState.Mode,
		"mode should remain NormalMode")
}

// TestUnit_ProjectSwitch_NextProjectAtLastStaysInPlace verifies that
// pressing NextProject ('}') when already at the last project does not
// change the selection.
func TestUnit_ProjectSwitch_NextProjectAtLastStaysInPlace(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Only Project"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	require.Equal(t, 0, m.AppState.SelectedProject())

	keyMsg := tea.KeyPressMsg(tea.Key{Text: "}", Code: '}'})
	updated := UpdateModelWithMessage(m, keyMsg)

	assert.Equal(t, 0, updated.AppState.SelectedProject(),
		"SelectedProject should remain 0 when there is only one project")
}
