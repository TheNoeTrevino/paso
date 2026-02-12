package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
)

// TestUnit_BoardRendering_EmptyBoard verifies that View() does not panic
// when the model has no projects, no columns, and no tasks.
func TestUnit_BoardRendering_EmptyBoard(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()
	m := SetupTestModelWithMocks(t, svc)

	// Ensure no data
	require.Empty(t, m.AppState.Projects())
	require.Empty(t, m.AppState.Columns())

	// Give the model a non-zero width so View() progresses past the "Loading..." guard
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// View() must not panic
	view := m.View()
	assert.NotEmpty(t, view.Content, "View should produce some output even with empty state")
}

// TestUnit_BoardRendering_WithColumnsShowsHeaders verifies that column names
// appear in the rendered output when the model has one project and columns.
func TestUnit_BoardRendering_WithColumnsShowsHeaders(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	// Configure mocks to return one project and columns
	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Test Project"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Backlog", ProjectID: 1},
		{ID: 20, Name: "In Progress", ProjectID: 1},
		{ID: 30, Name: "Done", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)

	// Set terminal dimensions
	m.UIState.SetWidth(160)
	m.UIState.Height = 40

	view := m.View()

	assert.Contains(t, view.Content, "Backlog", "rendered board should contain column header 'Backlog'")
	assert.Contains(t, view.Content, "In Progress", "rendered board should contain column header 'In Progress'")
	assert.Contains(t, view.Content, "Done", "rendered board should contain column header 'Done'")
}

// TestUnit_BoardRendering_MultipleProjectsShowsActiveProjectName verifies
// that the active project's name appears in the tab bar rendering.
func TestUnit_BoardRendering_MultipleProjectsShowsActiveProjectName(t *testing.T) {
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

	m.UIState.SetWidth(160)
	m.UIState.Height = 40

	view := m.View()

	// The tab bar should contain the project names
	assert.Contains(t, view.Content, "Alpha", "rendered board should show project name 'Alpha'")
	assert.Contains(t, view.Content, "Beta", "rendered board should show project name 'Beta'")
	assert.Contains(t, view.Content, "Gamma", "rendered board should show project name 'Gamma'")
}

// TestUnit_BoardRendering_WithTasksShowsTaskTitles verifies that task titles
// appear in the rendered output when the model has tasks in columns.
func TestUnit_BoardRendering_WithTasksShowsTaskTitles(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Test Project"},
	}
	col := &models.Column{ID: 10, Name: "Todo", ProjectID: 1}
	svc.Column.GetColumnsByProjectResult = []*models.Column{col}
	svc.Task.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
		10: {
			{ID: 1, Title: "Fix login bug", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"},
			{ID: 2, Title: "Add dark mode", ColumnID: 10, Position: 1, Labels: []*models.Label{}, PriorityColor: "#00FF00"},
		},
	}

	m := SetupTestModelWithMocks(t, svc)

	m.UIState.SetWidth(160)
	m.UIState.Height = 40

	view := m.View()

	assert.Contains(t, view.Content, "Fix login bug", "rendered board should show task title")
	assert.Contains(t, view.Content, "Add dark mode", "rendered board should show task title")
}
