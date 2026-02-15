package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
)

// TestUnit_SearchPersistence_RefreshMsgUsesFilteredQuery verifies that when
// a search filter is active and a RefreshMsg arrives (daemon event), the
// reload uses GetTaskSummariesByProjectFiltered instead of the unfiltered variant.
func TestUnit_SearchPersistence_RefreshMsgUsesFilteredQuery(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = make(map[int][]*models.TaskSummary)
	svc.Task.GetTaskSummariesByProjectFilteredResult = map[int][]*models.TaskSummary{
		10: {{ID: 1, Title: "Bug fix", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"}},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40
	m.SubscriptionStarted = true

	// Simulate an active search filter
	m.UI.Search.Query = "bug"
	m.UI.Search.Activate()

	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesByProjectFiltered")
	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	// Send a RefreshMsg (simulating a daemon event)
	refreshMsg := RefreshMsg{
		Event: events.Event{
			Type:       events.EventDatabaseChanged,
			ProjectID:  0,
			Timestamp:  time.Now(),
			SequenceID: 1,
		},
	}
	updated := UpdateModelWithMessage(m, refreshMsg)

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProjectFiltered"), initialFilteredCalls,
		"RefreshMsg should use filtered query when search is active")
	assert.Equal(t, initialUnfilteredCalls, svc.Task.CallCount("GetTaskSummariesByProject"),
		"RefreshMsg should NOT call unfiltered query when search is active")

	// Verify the search state is preserved
	assert.True(t, updated.UI.Search.IsActive, "search should remain active after RefreshMsg")
	assert.Equal(t, "bug", updated.UI.Search.Query, "search query should be preserved after RefreshMsg")
}

// TestUnit_SearchPersistence_RefreshMsgWithoutSearchUsesUnfiltered verifies
// that when no search is active, RefreshMsg uses the normal unfiltered query.
func TestUnit_SearchPersistence_RefreshMsgWithoutSearchUsesUnfiltered(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40
	m.SubscriptionStarted = true

	// No active search
	require.False(t, m.UI.Search.IsActive)
	require.Equal(t, "", m.UI.Search.Query)

	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	refreshMsg := RefreshMsg{
		Event: events.Event{
			Type:       events.EventDatabaseChanged,
			ProjectID:  0,
			Timestamp:  time.Now(),
			SequenceID: 1,
		},
	}
	UpdateModelWithMessage(m, refreshMsg)

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProject"), initialUnfilteredCalls,
		"RefreshMsg should use unfiltered query when no search is active")
	assert.Equal(t, 0, svc.Task.CallCount("GetTaskSummariesByProjectFiltered"),
		"RefreshMsg should NOT call filtered query when no search is active")
}

// TestUnit_SearchPersistence_ProjectSwitchCarriesSearchFilter verifies that
// switching projects with an active search carries the filter to the new project.
func TestUnit_SearchPersistence_ProjectSwitchCarriesSearchFilter(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
		{ID: 2, Name: "Beta"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectFilteredResult = map[int][]*models.TaskSummary{
		20: {{ID: 100, Title: "Beta bug", ColumnID: 20, Position: 0, Labels: []*models.Label{}, PriorityColor: "#00FF00"}},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Activate a search filter
	m.UI.Search.Query = "bug"
	m.UI.Search.Activate()

	// Reset call counts after init
	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesByProjectFiltered")

	// Set up what the second project should return
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 20, Name: "Backlog", ProjectID: 2},
	}

	// Switch to next project with '}'
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "}", Code: '}'})
	updated := UpdateModelWithMessage(m, keyMsg)

	// Verify filtered query was called for the new project
	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProjectFiltered"), initialFilteredCalls,
		"project switch should use filtered query when search is active")

	// Verify the search state is still active
	assert.True(t, updated.UI.Search.IsActive, "search should remain active after project switch")
	assert.Equal(t, "bug", updated.UI.Search.Query, "search query should persist across project switch")

	// Verify the filtered call included the right search query
	calls := svc.Task.GetCalls()
	foundFilteredCall := false
	for _, call := range calls {
		if call.Method == "GetTaskSummariesByProjectFiltered" {
			if query, ok := call.Args["searchQuery"].(string); ok && query == "bug" {
				foundFilteredCall = true
			}
		}
	}
	assert.True(t, foundFilteredCall,
		"project switch should pass the search query 'bug' to GetTaskSummariesByProjectFiltered")
}

// TestUnit_SearchPersistence_ProjectSwitchWithoutSearchUsesUnfiltered verifies
// that switching projects without a search uses the normal unfiltered query.
func TestUnit_SearchPersistence_ProjectSwitchWithoutSearchUsesUnfiltered(t *testing.T) {
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

	// No active search
	require.False(t, m.UI.Search.IsActive)

	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	// Switch to next project
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "}", Code: '}'})
	UpdateModelWithMessage(m, keyMsg)

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProject"), initialUnfilteredCalls,
		"project switch should use unfiltered query when no search is active")
	assert.Equal(t, 0, svc.Task.CallCount("GetTaskSummariesByProjectFiltered"),
		"project switch should NOT call filtered query when no search is active")
}

// TestUnit_SearchPersistence_DatabaseSwitchClearsSearch verifies that
// switching databases clears the active search filter.
func TestUnit_SearchPersistence_DatabaseSwitchClearsSearch(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40
	m.SubscriptionStarted = true

	// Activate a search filter
	m.UI.Search.Query = "important"
	m.UI.Search.Activate()
	require.True(t, m.UI.Search.IsActive)

	// Simulate a database switch completing by sending a dataReloaded message
	reloadedMsg := dataReloaded{
		projects: []*models.Project{{ID: 1, Name: "Alpha"}},
		columns:  []*models.Column{{ID: 10, Name: "Todo", ProjectID: 1}},
		tasks:    map[int][]*models.TaskSummary{10: {}},
		labels:   []*models.Label{},
	}
	updated := UpdateModelWithMessage(m, reloadedMsg)

	assert.False(t, updated.UI.Search.IsActive, "search should be deactivated after database switch")
	assert.Equal(t, "", updated.UI.Search.Query, "search query should be cleared after database switch")
}

// TestUnit_SearchPersistence_ReloadCurrentColumnTasksUsesFilter verifies that
// reloadCurrentColumnTasks (called from pickers) respects the active search.
func TestUnit_SearchPersistence_ReloadCurrentColumnTasksUsesFilter(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = map[int][]*models.TaskSummary{
		10: {{ID: 1, Title: "Task 1", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"}},
	}
	svc.Task.GetTaskSummariesByProjectFilteredResult = map[int][]*models.TaskSummary{
		10: {{ID: 1, Title: "Bug task", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"}},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Activate a search filter
	m.UI.Search.Query = "bug"
	m.UI.Search.Activate()

	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesByProjectFiltered")
	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	// Call reloadCurrentColumnTasks directly (simulates what pickers do)
	m.reloadCurrentColumnTasks()

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProjectFiltered"), initialFilteredCalls,
		"reloadCurrentColumnTasks should use filtered query when search is active")
	assert.Equal(t, initialUnfilteredCalls, svc.Task.CallCount("GetTaskSummariesByProject"),
		"reloadCurrentColumnTasks should NOT call unfiltered query when search is active")
}

// TestUnit_SearchPersistence_FetchTasksPassesCorrectQuery verifies that
// fetchTasksForCurrentProject passes the exact search query to the service.
func TestUnit_SearchPersistence_FetchTasksPassesCorrectQuery(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectFilteredResult = make(map[int][]*models.TaskSummary)

	m := SetupTestModelWithMocks(t, svc)

	// Set up an active search with a specific query
	m.UI.Search.Query = "ci pipeline"
	m.UI.Search.Activate()

	// Call reloadCurrentProject to trigger the helper
	m.reloadCurrentProject()

	// Find the filtered call and verify the exact query
	calls := svc.Task.GetCalls()
	var matchingCalls []string
	for _, call := range calls {
		if call.Method == "GetTaskSummariesByProjectFiltered" {
			if query, ok := call.Args["searchQuery"].(string); ok {
				matchingCalls = append(matchingCalls, query)
			}
		}
	}
	require.NotEmpty(t, matchingCalls, "expected at least one filtered call")
	assert.Equal(t, "ci pipeline", matchingCalls[len(matchingCalls)-1],
		"search query should be passed exactly as-is to the filtered service call")
}

// TestUnit_SearchPersistence_InactiveSearchWithQueryDoesNotFilter verifies
// that if the search has a query but IsActive is false, it uses the unfiltered path.
// This covers the case where the user typed a search but pressed Esc to cancel.
func TestUnit_SearchPersistence_InactiveSearchWithQueryDoesNotFilter(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40
	m.SubscriptionStarted = true

	// Set a query but don't activate it (user pressed Esc)
	m.UI.Search.Query = "test"
	m.UI.Search.Deactivate()
	require.False(t, m.UI.Search.IsActive)

	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	// Send a RefreshMsg
	refreshMsg := RefreshMsg{
		Event: events.Event{
			Type:       events.EventDatabaseChanged,
			ProjectID:  0,
			Timestamp:  time.Now(),
			SequenceID: 1,
		},
	}
	UpdateModelWithMessage(m, refreshMsg)

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProject"), initialUnfilteredCalls,
		"should use unfiltered query when search has a query but is not active")
	assert.Equal(t, 0, svc.Task.CallCount("GetTaskSummariesByProjectFiltered"),
		"should NOT use filtered query when search is not active")
}
