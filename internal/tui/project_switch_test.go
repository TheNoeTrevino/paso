package tui

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// detailPanelVisibleWidth is the minimum terminal width required for the detail panel
// to be visible during tests. This ensures prefetch behavior is triggered.
const detailPanelVisibleWidth = 200

// setupProjectSwitchTest creates a test model with initialized projects,
// detail cache, and UI state configured for project switching tests.
//
// Parameters:
//   - t: testing context
//   - selectedProject: which project index to start at (0-based)
func setupProjectSwitchTest(t *testing.T, selectedProject int) Model {
	t.Helper()

	m, _ := SetupTestModelWithDB(t)
	ctx := context.Background()

	// Load all available projects
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	require.NoError(t, err)

	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(selectedProject)

	// Initialize detail cache with reasonable size
	m.DetailCache = state.NewTaskDetailCache(10)

	// Set width to enable detail panel visibility
	m.UIState.SetWidth(detailPanelVisibleWidth)

	return m
}

// TestHandleNextProject_ClearsCacheAndTriggersPrefetch verifies that switching to the next
// project clears the detail cache and triggers a prefetch for the new project's tasks.
//
// Bug context: When switching projects with }, the detail cache should be cleared and
// the detail panel should load the task at the new cursor position (column 0, task 0).
func TestHandleNextProject_ClearsCacheAndTriggersPrefetch(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	ctx := context.Background()

	// Create a second project so we have 2 projects to switch between
	project2ID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Second Project")

	// Load both projects into the model
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(projects), 2)
	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(0) // Start at first project

	// Load columns for the second project (CreateTestProject creates default columns)
	columns2, err := m.App.ColumnService.GetColumnsByProject(ctx, project2ID)
	require.NoError(t, err)
	require.NotEmpty(t, columns2)

	// Create a task in the first column of the second project
	firstColID := columns2[0].ID
	taskID := fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), firstColID, "Task in Project 2")

	// Set terminal width to enable detail panel
	m.UIState.SetWidth(detailPanelVisibleWidth)

	// Initialize detail cache
	m.DetailCache = state.NewTaskDetailCache(10)

	// Pre-populate the cache with a fake task detail
	m.DetailCache.Set(999, &models.TaskDetail{
		ID:          999,
		Title:       "Cached Task",
		Description: "This should be cleared",
	})

	// Verify cache has the entry
	require.True(t, m.DetailCache.Has(999), "Cache should have task 999 before switching projects")

	// Call handleNextProject
	updatedModel, cmd := m.handleNextProject()
	m = updatedModel.(Model)

	// Assert: Selected project changed from 0 to 1
	assert.Equal(t, 1, m.AppState.SelectedProject())

	// Assert: Detail cache was cleared
	assert.False(t, m.DetailCache.Has(999), "Detail cache should be cleared after switching projects")

	// Assert: UIState selection was reset to 0, 0
	assert.Equal(t, 0, m.UIState.SelectedColumn)
	assert.Equal(t, 0, m.UIState.SelectedTask)

	// Verify the new project's data was loaded
	currentProject := m.AppState.GetCurrentProject()
	require.NotNil(t, currentProject)
	assert.Equal(t, project2ID, currentProject.ID)

	// Verify the columns and tasks were loaded for the new project
	columns := m.AppState.Columns()
	assert.NotEmpty(t, columns, "Columns should be loaded for the new project")

	// Verify task exists in the new project
	loadedTasks := m.AppState.Tasks()
	assert.NotEmpty(t, loadedTasks[firstColID], "Tasks should be loaded for column in new project")
	assert.Equal(t, taskID, loadedTasks[firstColID][0].ID)

	// Assert: Returned cmd is non-nil (the prefetch command)
	assert.NotNil(t, cmd, "handleNextProject should return a non-nil tea.Cmd for prefetch when tasks exist and detail panel is visible")

	t.Logf("Cache cleared: task 999 no longer cached")
	t.Logf("Prefetch command returned: non-nil")
	t.Logf("Selection reset: column=%d, task=%d", m.UIState.SelectedColumn, m.UIState.SelectedTask)
	t.Logf("Switched to project: %s (ID=%d)", currentProject.Name, currentProject.ID)
}

// TestHandlePrevProject_ClearsCacheAndTriggersPrefetch verifies that switching to the previous
// project clears the detail cache and triggers a prefetch for the new project's tasks.
//
// Bug context: When switching projects with {, the detail cache should be cleared and
// the detail panel should load the task at the new cursor position (column 0, task 0).
func TestHandlePrevProject_ClearsCacheAndTriggersPrefetch(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	ctx := context.Background()

	// Get the first project that was created by SetupTestModelWithDB
	project1, err := m.App.ProjectService.GetAllProjects(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, project1)
	project1ID := project1[0].ID

	// Create tasks in project 1 so there's something to prefetch when we switch back
	cols1, err := m.App.ColumnService.GetColumnsByProject(ctx, project1ID)
	require.NoError(t, err)
	require.NotEmpty(t, cols1)
	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), cols1[0].ID, "Task in Project 1")

	// Create a second project so we have 2 projects to switch between
	project2ID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Second Project")

	// Load both projects into the model
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(projects), 2)
	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(1) // Start at second project (index 1)

	// Load columns for the second project (CreateTestProject creates default columns)
	columns2, err := m.App.ColumnService.GetColumnsByProject(ctx, project2ID)
	require.NoError(t, err)
	require.NotEmpty(t, columns2)

	// Create a task in the first column of the second project so there's something to prefetch
	firstColID := columns2[0].ID
	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), firstColID, "Task in Project 2")

	// Set terminal width to enable detail panel
	m.UIState.SetWidth(detailPanelVisibleWidth)

	// Initialize detail cache
	m.DetailCache = state.NewTaskDetailCache(10)

	// Pre-populate the cache with a fake task detail
	m.DetailCache.Set(888, &models.TaskDetail{
		ID:          888,
		Title:       "Cached Task",
		Description: "This should be cleared",
	})

	// Verify cache has the entry
	require.True(t, m.DetailCache.Has(888), "Cache should have task 888 before switching projects")

	// Call handlePrevProject
	updatedModel, cmd := m.handlePrevProject()
	m = updatedModel.(Model)

	// Assert: Selected project changed from 1 to 0
	assert.Equal(t, 0, m.AppState.SelectedProject())

	// Assert: Detail cache was cleared
	assert.False(t, m.DetailCache.Has(888), "Detail cache should be cleared after switching projects")

	// Assert: Returned cmd is non-nil (the prefetch command)
	assert.NotNil(t, cmd, "handlePrevProject should return a non-nil tea.Cmd for prefetch")

	// Assert: UIState selection was reset to 0, 0
	assert.Equal(t, 0, m.UIState.SelectedColumn)
	assert.Equal(t, 0, m.UIState.SelectedTask)

	t.Logf("Cache cleared: task 888 no longer cached")
	t.Logf("Prefetch command returned: non-nil")
	t.Logf("Selection reset: column=%d, task=%d", m.UIState.SelectedColumn, m.UIState.SelectedTask)
	t.Logf("Switched back to first project (index 0)")
}

// TestHandleNextProject_AtLastProject_NoOp verifies that attempting to switch to the next
// project when already at the last project shows a notification and returns nil command.
//
// Edge case: User presses } when already at the last project.
func TestHandleNextProject_AtLastProject_NoOp(t *testing.T) {
	t.Parallel()
	m := setupProjectSwitchTest(t, 0)

	// Position at last project for this edge case test
	m.AppState.SetSelectedProject(len(m.AppState.Projects()) - 1)

	// Call handleNextProject when already at last project
	initialProjectIndex := m.AppState.SelectedProject()
	updatedModel, cmd := m.handleNextProject()
	m = updatedModel.(Model)

	// Assert: Selected project unchanged
	assert.Equal(t, initialProjectIndex, m.AppState.SelectedProject())

	// Assert: Notification was set
	assert.True(t, m.UI.Notification.HasAny(), "handleNextProject at last project should set notification")

	notifications := m.UI.Notification.All()
	expectedMsg := "Already at the last project"
	require.NotEmpty(t, notifications)
	assert.Equal(t, expectedMsg, notifications[0].Message)

	// Assert: Returned cmd is nil (no prefetch needed)
	assert.Nil(t, cmd, "handleNextProject at last project should return nil tea.Cmd")

	t.Logf("No project switch occurred")
	t.Logf("Notification shown: %q", expectedMsg)
	t.Logf("No prefetch command returned (nil)")
}

// TestHandlePrevProject_AtFirstProject_NoOp verifies that attempting to switch to the previous
// project when already at the first project shows a notification and returns nil command.
//
// Edge case: User presses { when already at the first project.
func TestHandlePrevProject_AtFirstProject_NoOp(t *testing.T) {
	t.Parallel()
	m := setupProjectSwitchTest(t, 0)

	// Call handlePrevProject when already at first project
	initialProjectIndex := m.AppState.SelectedProject()
	updatedModel, cmd := m.handlePrevProject()
	m = updatedModel.(Model)

	// Assert: Selected project unchanged
	assert.Equal(t, initialProjectIndex, m.AppState.SelectedProject())

	// Assert: Notification was set
	assert.True(t, m.UI.Notification.HasAny(), "handlePrevProject at first project should set notification")

	notifications := m.UI.Notification.All()
	expectedMsg := "Already at the first project"
	require.NotEmpty(t, notifications)
	assert.Equal(t, expectedMsg, notifications[0].Message)

	// Assert: Returned cmd is nil (no prefetch needed)
	assert.Nil(t, cmd, "handlePrevProject at first project should return nil tea.Cmd")

	t.Logf("No project switch occurred")
	t.Logf("Notification shown: %q", expectedMsg)
	t.Logf("No prefetch command returned (nil)")
}

// TestProjectSwitch_MultipleProjectsRoundTrip verifies that switching back and forth
// between projects works correctly with cache clearing and prefetch on each switch.
//
// Bug context: Ensures that rapidly switching between projects always clears the cache
// and triggers prefetch, preventing stale task details from appearing.
func TestProjectSwitch_MultipleProjectsRoundTrip(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	ctx := context.Background()

	// Get the first project that was created by SetupTestModelWithDB
	project1, err := m.App.ProjectService.GetAllProjects(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, project1)
	project1ID := project1[0].ID

	// Create two more projects (total of 3)
	project2ID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Project Two")
	project3ID := fixtures.CreateTestProject(t, db, fixtures.SQLiteDialect(), "Project Three")

	// Load all projects
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(projects), 3)
	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(0) // Start at first project

	// Create tasks in the first column of each project so there's something to prefetch
	cols1, _ := m.App.ColumnService.GetColumnsByProject(ctx, project1ID)
	cols2, _ := m.App.ColumnService.GetColumnsByProject(ctx, project2ID)
	cols3, _ := m.App.ColumnService.GetColumnsByProject(ctx, project3ID)
	if len(cols1) > 0 {
		fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), cols1[0].ID, "Task in Project 1")
	}
	if len(cols2) > 0 {
		fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), cols2[0].ID, "Task in Project 2")
	}
	if len(cols3) > 0 {
		fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), cols3[0].ID, "Task in Project 3")
	}

	// Set terminal width to enable detail panel
	m.UIState.SetWidth(detailPanelVisibleWidth)

	// Initialize detail cache
	m.DetailCache = state.NewTaskDetailCache(10)

	// Populate cache with task from project 1
	m.DetailCache.Set(100, &models.TaskDetail{ID: 100, Title: "Project 1 Task"})
	require.True(t, m.DetailCache.Has(100), "Cache should have task 100")

	// Switch to project 2 (index 0 -> 1)
	updatedModel, cmd := m.handleNextProject()
	m = updatedModel.(Model)

	assert.Equal(t, 1, m.AppState.SelectedProject())
	assert.False(t, m.DetailCache.Has(100), "Cache should be cleared after first switch")
	assert.NotNil(t, cmd, "First switch should return prefetch cmd")

	// Populate cache with task from project 2
	m.DetailCache.Set(200, &models.TaskDetail{ID: 200, Title: "Project 2 Task"})

	// Switch to project 3 (index 1 -> 2)
	updatedModel, cmd = m.handleNextProject()
	m = updatedModel.(Model)

	assert.Equal(t, 2, m.AppState.SelectedProject())
	assert.False(t, m.DetailCache.Has(200), "Cache should be cleared after second switch")
	assert.NotNil(t, cmd, "Second switch should return prefetch cmd")

	// Populate cache with task from project 3
	m.DetailCache.Set(300, &models.TaskDetail{ID: 300, Title: "Project 3 Task"})

	// Switch back to project 2 (index 2 -> 1)
	updatedModel, cmd = m.handlePrevProject()
	m = updatedModel.(Model)

	assert.Equal(t, 1, m.AppState.SelectedProject())
	assert.False(t, m.DetailCache.Has(300), "Cache should be cleared after third switch")
	assert.NotNil(t, cmd, "Third switch should return prefetch cmd")

	// Populate cache again
	m.DetailCache.Set(400, &models.TaskDetail{ID: 400, Title: "Back to Project 2"})

	// Switch back to project 1 (index 1 -> 0)
	updatedModel, cmd = m.handlePrevProject()
	m = updatedModel.(Model)

	assert.Equal(t, 0, m.AppState.SelectedProject())
	assert.False(t, m.DetailCache.Has(400), "Cache should be cleared after fourth switch")
	assert.NotNil(t, cmd, "Fourth switch should return prefetch cmd")

	t.Logf("Round-trip project switching works correctly")
	t.Logf("Cache cleared on each switch: 4/4 switches")
	t.Logf("Prefetch cmd returned on each switch: 4/4 switches")
	t.Logf("Final project: %d (back to start)", m.AppState.SelectedProject())
}
