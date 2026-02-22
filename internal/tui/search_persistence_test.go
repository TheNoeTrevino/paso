package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
	tasksvc "github.com/thenoetrevino/paso/internal/services/task"
)

// TestUnit_SearchPersistence_RefreshMsgUsesFilteredQuery verifies that when
// a search filter is active and a RefreshMsg arrives (daemon event), the
// reload uses GetTaskSummariesWithFilters instead of the unfiltered variant.
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
	svc.Task.GetTaskSummariesWithFiltersResult = map[int][]*models.TaskSummary{
		10: {{ID: 1, Title: "Bug fix", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"}},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40
	m.SubscriptionStarted = true

	// Simulate an active search filter via the filter bar
	m.UI.Filter.SearchQuery = "bug"

	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesWithFilters")
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

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesWithFilters"), initialFilteredCalls,
		"RefreshMsg should use filtered query when search is active")
	assert.Equal(t, initialUnfilteredCalls, svc.Task.CallCount("GetTaskSummariesByProject"),
		"RefreshMsg should NOT call unfiltered query when search is active")

	// Verify the search state is preserved
	assert.Equal(t, "bug", updated.UI.Filter.SearchQuery, "search query should be preserved after RefreshMsg")
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
	require.Equal(t, "", m.UI.Filter.SearchQuery)

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
	assert.Equal(t, 0, svc.Task.CallCount("GetTaskSummariesWithFilters"),
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
	svc.Task.GetTaskSummariesWithFiltersResult = map[int][]*models.TaskSummary{
		20: {{ID: 100, Title: "Beta bug", ColumnID: 20, Position: 0, Labels: []*models.Label{}, PriorityColor: "#00FF00"}},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Activate a search filter
	m.UI.Filter.SearchQuery = "bug"

	// Reset call counts after init
	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesWithFilters")

	// Set up what the second project should return
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 20, Name: "Backlog", ProjectID: 2},
	}

	// Switch to next project with '}'
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "}", Code: '}'})
	updated := UpdateModelWithMessage(m, keyMsg)

	// Verify filtered query was called for the new project
	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesWithFilters"), initialFilteredCalls,
		"project switch should use filtered query when search is active")

	// Verify the search state is still active
	assert.Equal(t, "bug", updated.UI.Filter.SearchQuery, "search query should persist across project switch")

	// Verify the filtered call included the right search query
	calls := svc.Task.GetCalls()
	foundFilteredCall := false
	for _, call := range calls {
		if call.Method == "GetTaskSummariesWithFilters" {
			if params, ok := call.Args["params"].(tasksvc.TaskFilterParams); ok && params.Title != nil && *params.Title == "bug" {
				foundFilteredCall = true
			}
		}
	}
	assert.True(t, foundFilteredCall,
		"project switch should pass the search query 'bug' to GetTaskSummariesWithFilters")
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
	require.Equal(t, "", m.UI.Filter.SearchQuery)

	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	// Switch to next project
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "}", Code: '}'})
	UpdateModelWithMessage(m, keyMsg)

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProject"), initialUnfilteredCalls,
		"project switch should use unfiltered query when no search is active")
	assert.Equal(t, 0, svc.Task.CallCount("GetTaskSummariesWithFilters"),
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
	m.UI.Filter.SearchQuery = "important"
	require.Equal(t, "important", m.UI.Filter.SearchQuery)

	// Simulate a database switch completing by sending a dataReloaded message
	reloadedMsg := dataReloaded{
		projects: []*models.Project{{ID: 1, Name: "Alpha"}},
		columns:  []*models.Column{{ID: 10, Name: "Todo", ProjectID: 1}},
		tasks:    map[int][]*models.TaskSummary{10: {}},
		labels:   []*models.Label{},
	}
	updated := UpdateModelWithMessage(m, reloadedMsg)

	assert.Equal(t, "", updated.UI.Filter.SearchQuery, "search query should be cleared after database switch")
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
	svc.Task.GetTaskSummariesWithFiltersResult = map[int][]*models.TaskSummary{
		10: {{ID: 1, Title: "Bug task", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"}},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Activate a search filter
	m.UI.Filter.SearchQuery = "bug"

	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesWithFilters")
	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	// Call reloadCurrentColumnTasks directly (simulates what pickers do)
	m.reloadCurrentColumnTasks()

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesWithFilters"), initialFilteredCalls,
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
	svc.Task.GetTaskSummariesWithFiltersResult = make(map[int][]*models.TaskSummary)

	m := SetupTestModelWithMocks(t, svc)

	// Set up an active search with a specific query
	m.UI.Filter.SearchQuery = "ci pipeline"

	// Call reloadCurrentProject to trigger the helper
	m.reloadCurrentProject()

	// Find the filtered call and verify the exact query
	calls := svc.Task.GetCalls()
	var matchingQueries []string
	for _, call := range calls {
		if call.Method == "GetTaskSummariesWithFilters" {
			if params, ok := call.Args["params"].(tasksvc.TaskFilterParams); ok && params.Title != nil {
				matchingQueries = append(matchingQueries, *params.Title)
			}
		}
	}
	require.NotEmpty(t, matchingQueries, "expected at least one filtered call")
	assert.Equal(t, "ci pipeline", matchingQueries[len(matchingQueries)-1],
		"search query should be passed exactly as-is to the filtered service call")
}

// TestUnit_SearchPersistence_EmptySearchQueryDoesNotFilter verifies
// that when the search query is empty, it uses the unfiltered path.
func TestUnit_SearchPersistence_EmptySearchQueryDoesNotFilter(t *testing.T) {
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

	// Empty search query should not trigger filtered path
	m.UI.Filter.SearchQuery = ""

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
		"should use unfiltered query when search query is empty")
	assert.Equal(t, 0, svc.Task.CallCount("GetTaskSummariesWithFilters"),
		"should NOT use filtered query when search query is empty")
}

// TestUnit_SearchPersistence_CreateTaskRespectsActiveFilters verifies that
// creating a new task reloads with GetTaskSummariesWithFilters when filters are active.
func TestUnit_SearchPersistence_CreateTaskRespectsActiveFilters(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = make(map[int][]*models.TaskSummary)
	svc.Task.GetTaskSummariesWithFiltersResult = map[int][]*models.TaskSummary{
		10: {{ID: 99, Title: "Filtered task", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"}},
	}
	svc.Task.CreateTaskResult = &models.Task{ID: 50, Title: "New task"}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Activate a search filter before creating a task
	m.UI.Filter.SearchQuery = "bug"

	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesWithFilters")
	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	// Create a task
	m.createNewTaskWithLabelsAndRelationships(taskFormValues{
		title:       "New task",
		description: "A new task",
		confirm:     true,
	})

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesWithFilters"), initialFilteredCalls,
		"createNewTask should use filtered query when search is active")
	assert.Equal(t, initialUnfilteredCalls, svc.Task.CallCount("GetTaskSummariesByProject"),
		"createNewTask should NOT call unfiltered query when search is active")
}

// TestUnit_SearchPersistence_CreateTaskWithoutFiltersUsesUnfiltered verifies that
// creating a task without active filters uses the unfiltered query.
func TestUnit_SearchPersistence_CreateTaskWithoutFiltersUsesUnfiltered(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = make(map[int][]*models.TaskSummary)
	svc.Task.CreateTaskResult = &models.Task{ID: 50, Title: "New task"}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// No filters active
	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	m.createNewTaskWithLabelsAndRelationships(taskFormValues{
		title:       "New task",
		description: "A new task",
		confirm:     true,
	})

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesByProject"), initialUnfilteredCalls,
		"createNewTask should use unfiltered query when no filters are active")
	assert.Equal(t, 0, svc.Task.CallCount("GetTaskSummariesWithFilters"),
		"createNewTask should NOT call filtered query when no filters are active")
}

// TestUnit_SearchPersistence_UpdateTaskRespectsActiveFilters verifies that
// updating an existing task reloads with GetTaskSummariesWithFilters when filters are active.
func TestUnit_SearchPersistence_UpdateTaskRespectsActiveFilters(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = make(map[int][]*models.TaskSummary)
	svc.Task.GetTaskSummariesWithFiltersResult = map[int][]*models.TaskSummary{
		10: {{ID: 1, Title: "Filtered task", ColumnID: 10, Position: 0, Labels: []*models.Label{}, PriorityColor: "#FF0000"}},
	}
	svc.Task.GetTaskDetailResult = &models.TaskDetail{
		ID:     1,
		Title:  "Existing task",
		Labels: []*models.Label{},
	}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40
	m.Forms.Form.EditingTaskID = 1

	// Activate a search filter before updating a task
	m.UI.Filter.SearchQuery = "bug"

	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesWithFilters")
	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	m.updateExistingTaskWithLabelsAndRelationships(taskFormValues{
		title:       "Updated task",
		description: "Updated description",
		confirm:     true,
	})

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesWithFilters"), initialFilteredCalls,
		"updateExistingTask should use filtered query when search is active")
	assert.Equal(t, initialUnfilteredCalls, svc.Task.CallCount("GetTaskSummariesByProject"),
		"updateExistingTask should NOT call unfiltered query when search is active")
}

// TestUnit_SearchPersistence_CreateTaskRespectsFieldFilters verifies that
// creating a task respects non-search filters (e.g. priority, label filters).
func TestUnit_SearchPersistence_CreateTaskRespectsFieldFilters(t *testing.T) {
	t.Parallel()
	svc := NewDefaultMockServices()

	svc.Project.GetAllProjectsResult = []*models.Project{
		{ID: 1, Name: "Alpha"},
	}
	svc.Column.GetColumnsByProjectResult = []*models.Column{
		{ID: 10, Name: "Todo", ProjectID: 1},
	}
	svc.Task.GetTaskSummariesByProjectResult = make(map[int][]*models.TaskSummary)
	svc.Task.GetTaskSummariesWithFiltersResult = make(map[int][]*models.TaskSummary)
	svc.Task.CreateTaskResult = &models.Task{ID: 50, Title: "New task"}

	m := SetupTestModelWithMocks(t, svc)
	m.UIState.SetWidth(120)
	m.UIState.Height = 40

	// Activate a priority filter (no search query)
	priorityID := 2
	m.UI.Filter.PriorityID = &priorityID

	initialFilteredCalls := svc.Task.CallCount("GetTaskSummariesWithFilters")
	initialUnfilteredCalls := svc.Task.CallCount("GetTaskSummariesByProject")

	m.createNewTaskWithLabelsAndRelationships(taskFormValues{
		title:       "New task",
		description: "A new task",
		confirm:     true,
	})

	assert.Greater(t, svc.Task.CallCount("GetTaskSummariesWithFilters"), initialFilteredCalls,
		"createNewTask should use filtered query when priority filter is active")
	assert.Equal(t, initialUnfilteredCalls, svc.Task.CallCount("GetTaskSummariesByProject"),
		"createNewTask should NOT call unfiltered query when priority filter is active")

	// Verify the filter params included the priority
	calls := svc.Task.GetCalls()
	for i := len(calls) - 1; i >= 0; i-- {
		if calls[i].Method == "GetTaskSummariesWithFilters" {
			params, ok := calls[i].Args["params"].(tasksvc.TaskFilterParams)
			require.True(t, ok, "expected TaskFilterParams in call args")
			require.NotNil(t, params.PriorityID, "expected PriorityID to be set")
			assert.Equal(t, 2, *params.PriorityID, "PriorityID should match active filter")
			break
		}
	}
}
