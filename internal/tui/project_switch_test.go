package tui

import (
	"context"
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestHandleNextProject_ClearsCacheAndTriggersPrefetch verifies that switching to the next
// project clears the detail cache and triggers a prefetch for the new project's tasks.
//
// Bug context: When switching projects with }, the detail cache should be cleared and
// the detail panel should load the task at the new cursor position (column 0, task 0).
func TestHandleNextProject_ClearsCacheAndTriggersPrefetch(t *testing.T) {
	m, db := SetupTestModelWithDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a second project so we have 2 projects to switch between
	project2ID := testutil.CreateTestProject(t, db, "Second Project")

	// Load both projects into the model
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to load projects: %v", err)
	}
	if len(projects) < 2 {
		t.Fatalf("Expected at least 2 projects, got %d", len(projects))
	}
	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(0) // Start at first project

	// Load columns for the second project (CreateTestProject creates default columns)
	columns2, err := m.App.ColumnService.GetColumnsByProject(ctx, project2ID)
	if err != nil || len(columns2) == 0 {
		t.Fatalf("Failed to load columns for project 2: %v", err)
	}

	// Create a task in the first column of the second project
	firstColID := columns2[0].ID
	taskID := testutil.CreateTestTask(t, db, firstColID, "Task in Project 2")

	// Set terminal width large enough to show detail panel
	m.UIState.SetWidth(200) // Large enough for detail panel to be visible

	// Initialize detail cache
	m.DetailCache = state.NewTaskDetailCache(10)

	// Pre-populate the cache with a fake task detail
	m.DetailCache.Set(999, &models.TaskDetail{
		ID:          999,
		Title:       "Cached Task",
		Description: "This should be cleared",
	})

	// Verify cache has the entry
	if !m.DetailCache.Has(999) {
		t.Fatal("Cache should have task 999 before switching projects")
	}

	// Call handleNextProject
	updatedModel, cmd := m.handleNextProject()
	m = updatedModel.(Model)

	// Assert: Selected project changed from 0 to 1
	if m.AppState.SelectedProject() != 1 {
		t.Errorf("SelectedProject after handleNextProject = %d, want 1", m.AppState.SelectedProject())
	}

	// Assert: Detail cache was cleared
	if m.DetailCache.Has(999) {
		t.Error("Detail cache should be cleared after switching projects")
	}

	// Assert: UIState selection was reset to 0, 0
	if m.UIState.SelectedColumn != 0 {
		t.Errorf("SelectedColumn after project switch = %d, want 0 (reset)", m.UIState.SelectedColumn)
	}
	if m.UIState.SelectedTask != 0 {
		t.Errorf("SelectedTask after project switch = %d, want 0 (reset)", m.UIState.SelectedTask)
	}

	// Verify the new project's data was loaded
	currentProject := m.AppState.GetCurrentProject()
	if currentProject == nil {
		t.Fatal("Current project should not be nil after switch")
	}
	if currentProject.ID != project2ID {
		t.Errorf("Current project ID = %d, want %d", currentProject.ID, project2ID)
	}

	// Verify the columns and tasks were loaded for the new project
	columns := m.AppState.Columns()
	if len(columns) == 0 {
		t.Error("Columns should be loaded for the new project")
	}

	// Verify task exists in the new project
	loadedTasks := m.AppState.Tasks()
	if len(loadedTasks[firstColID]) == 0 {
		t.Errorf("Tasks should be loaded for column %d in new project", firstColID)
	}
	if loadedTasks[firstColID][0].ID != taskID {
		t.Errorf("First task ID = %d, want %d", loadedTasks[firstColID][0].ID, taskID)
	}

	// Assert: Returned cmd is non-nil (the prefetch command)
	if cmd == nil {
		t.Error("handleNextProject should return a non-nil tea.Cmd for prefetch when tasks exist and detail panel is visible")
	}

	t.Logf("✓ Cache cleared: task 999 no longer cached")
	t.Logf("✓ Prefetch command returned: non-nil")
	t.Logf("✓ Selection reset: column=%d, task=%d", m.UIState.SelectedColumn, m.UIState.SelectedTask)
	t.Logf("✓ Switched to project: %s (ID=%d)", currentProject.Name, currentProject.ID)
}

// TestHandlePrevProject_ClearsCacheAndTriggersPrefetch verifies that switching to the previous
// project clears the detail cache and triggers a prefetch for the new project's tasks.
//
// Bug context: When switching projects with {, the detail cache should be cleared and
// the detail panel should load the task at the new cursor position (column 0, task 0).
func TestHandlePrevProject_ClearsCacheAndTriggersPrefetch(t *testing.T) {
	m, db := SetupTestModelWithDB(t)
	defer db.Close()

	ctx := context.Background()

	// Get the first project that was created by SetupTestModelWithDB
	project1, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil || len(project1) == 0 {
		t.Fatalf("Failed to load project 1: %v", err)
	}
	project1ID := project1[0].ID

	// Create tasks in project 1 so there's something to prefetch when we switch back
	cols1, err := m.App.ColumnService.GetColumnsByProject(ctx, project1ID)
	if err != nil || len(cols1) == 0 {
		t.Fatalf("Failed to load columns for project 1: %v", err)
	}
	testutil.CreateTestTask(t, db, cols1[0].ID, "Task in Project 1")

	// Create a second project so we have 2 projects to switch between
	project2ID := testutil.CreateTestProject(t, db, "Second Project")

	// Load both projects into the model
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to load projects: %v", err)
	}
	if len(projects) < 2 {
		t.Fatalf("Expected at least 2 projects, got %d", len(projects))
	}
	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(1) // Start at second project (index 1)

	// Load columns for the second project (CreateTestProject creates default columns)
	columns2, err := m.App.ColumnService.GetColumnsByProject(ctx, project2ID)
	if err != nil || len(columns2) == 0 {
		t.Fatalf("Failed to load columns for project 2: %v", err)
	}

	// Create a task in the first column of the second project so there's something to prefetch
	firstColID := columns2[0].ID
	testutil.CreateTestTask(t, db, firstColID, "Task in Project 2")

	// Set terminal width large enough to show detail panel
	m.UIState.SetWidth(200)

	// Initialize detail cache
	m.DetailCache = state.NewTaskDetailCache(10)

	// Pre-populate the cache with a fake task detail
	m.DetailCache.Set(888, &models.TaskDetail{
		ID:          888,
		Title:       "Cached Task",
		Description: "This should be cleared",
	})

	// Verify cache has the entry
	if !m.DetailCache.Has(888) {
		t.Fatal("Cache should have task 888 before switching projects")
	}

	// Call handlePrevProject
	updatedModel, cmd := m.handlePrevProject()
	m = updatedModel.(Model)

	// Assert: Selected project changed from 1 to 0
	if m.AppState.SelectedProject() != 0 {
		t.Errorf("SelectedProject after handlePrevProject = %d, want 0", m.AppState.SelectedProject())
	}

	// Assert: Detail cache was cleared
	if m.DetailCache.Has(888) {
		t.Error("Detail cache should be cleared after switching projects")
	}

	// Assert: Returned cmd is non-nil (the prefetch command)
	if cmd == nil {
		t.Error("handlePrevProject should return a non-nil tea.Cmd for prefetch")
	}

	// Assert: UIState selection was reset to 0, 0
	if m.UIState.SelectedColumn != 0 {
		t.Errorf("SelectedColumn after project switch = %d, want 0 (reset)", m.UIState.SelectedColumn)
	}
	if m.UIState.SelectedTask != 0 {
		t.Errorf("SelectedTask after project switch = %d, want 0 (reset)", m.UIState.SelectedTask)
	}

	t.Logf("✓ Cache cleared: task 888 no longer cached")
	t.Logf("✓ Prefetch command returned: non-nil")
	t.Logf("✓ Selection reset: column=%d, task=%d", m.UIState.SelectedColumn, m.UIState.SelectedTask)
	t.Logf("✓ Switched back to first project (index 0)")
}

// TestHandleNextProject_AtLastProject_NoOp verifies that attempting to switch to the next
// project when already at the last project shows a notification and returns nil command.
//
// Edge case: User presses } when already at the last project.
// Security value: No change, no panic, appropriate notification shown.
func TestHandleNextProject_AtLastProject_NoOp(t *testing.T) {
	m, db := SetupTestModelWithDB(t)
	defer db.Close()

	ctx := context.Background()

	// Load all projects (should be just 1 from setup)
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to load projects: %v", err)
	}
	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(len(projects) - 1) // Set to last project

	// Initialize detail cache
	m.DetailCache = state.NewTaskDetailCache(10)

	// Call handleNextProject when already at last project
	initialProjectIndex := m.AppState.SelectedProject()
	updatedModel, cmd := m.handleNextProject()
	m = updatedModel.(Model)

	// Assert: Selected project unchanged
	if m.AppState.SelectedProject() != initialProjectIndex {
		t.Errorf("SelectedProject should remain %d, got %d", initialProjectIndex, m.AppState.SelectedProject())
	}

	// Assert: Notification was set
	if !m.UI.Notification.HasAny() {
		t.Error("handleNextProject at last project should set notification")
	}

	notifications := m.UI.Notification.All()
	expectedMsg := "Already at the last project"
	if len(notifications) == 0 || notifications[0].Message != expectedMsg {
		t.Errorf("Notification message = %q, want %q", notifications[0].Message, expectedMsg)
	}

	// Assert: Returned cmd is nil (no prefetch needed)
	if cmd != nil {
		t.Error("handleNextProject at last project should return nil tea.Cmd")
	}

	t.Logf("✓ No project switch occurred")
	t.Logf("✓ Notification shown: %q", expectedMsg)
	t.Logf("✓ No prefetch command returned (nil)")
}

// TestHandlePrevProject_AtFirstProject_NoOp verifies that attempting to switch to the previous
// project when already at the first project shows a notification and returns nil command.
//
// Edge case: User presses { when already at the first project.
// Security value: No change, no panic, appropriate notification shown.
func TestHandlePrevProject_AtFirstProject_NoOp(t *testing.T) {
	m, db := SetupTestModelWithDB(t)
	defer db.Close()

	ctx := context.Background()

	// Load all projects (should be just 1 from setup)
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to load projects: %v", err)
	}
	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(0) // Already at first project

	// Initialize detail cache
	m.DetailCache = state.NewTaskDetailCache(10)

	// Call handlePrevProject when already at first project
	initialProjectIndex := m.AppState.SelectedProject()
	updatedModel, cmd := m.handlePrevProject()
	m = updatedModel.(Model)

	// Assert: Selected project unchanged
	if m.AppState.SelectedProject() != initialProjectIndex {
		t.Errorf("SelectedProject should remain %d, got %d", initialProjectIndex, m.AppState.SelectedProject())
	}

	// Assert: Notification was set
	if !m.UI.Notification.HasAny() {
		t.Error("handlePrevProject at first project should set notification")
	}

	notifications := m.UI.Notification.All()
	expectedMsg := "Already at the first project"
	if len(notifications) == 0 || notifications[0].Message != expectedMsg {
		t.Errorf("Notification message = %q, want %q", notifications[0].Message, expectedMsg)
	}

	// Assert: Returned cmd is nil (no prefetch needed)
	if cmd != nil {
		t.Error("handlePrevProject at first project should return nil tea.Cmd")
	}

	t.Logf("✓ No project switch occurred")
	t.Logf("✓ Notification shown: %q", expectedMsg)
	t.Logf("✓ No prefetch command returned (nil)")
}

// TestProjectSwitch_MultipleProjectsRoundTrip verifies that switching back and forth
// between projects works correctly with cache clearing and prefetch on each switch.
//
// Bug context: Ensures that rapidly switching between projects always clears the cache
// and triggers prefetch, preventing stale task details from appearing.
func TestProjectSwitch_MultipleProjectsRoundTrip(t *testing.T) {
	m, db := SetupTestModelWithDB(t)
	defer db.Close()

	ctx := context.Background()

	// Get the first project that was created by SetupTestModelWithDB
	project1, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil || len(project1) == 0 {
		t.Fatalf("Failed to load project 1: %v", err)
	}
	project1ID := project1[0].ID

	// Create two more projects (total of 3)
	project2ID := testutil.CreateTestProject(t, db, "Project Two")
	project3ID := testutil.CreateTestProject(t, db, "Project Three")

	// Load all projects
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to load projects: %v", err)
	}
	if len(projects) < 3 {
		t.Fatalf("Expected at least 3 projects, got %d", len(projects))
	}
	m.AppState.SetProjects(projects)
	m.AppState.SetSelectedProject(0) // Start at first project

	// Create tasks in the first column of each project so there's something to prefetch
	cols1, _ := m.App.ColumnService.GetColumnsByProject(ctx, project1ID)
	cols2, _ := m.App.ColumnService.GetColumnsByProject(ctx, project2ID)
	cols3, _ := m.App.ColumnService.GetColumnsByProject(ctx, project3ID)
	if len(cols1) > 0 {
		testutil.CreateTestTask(t, db, cols1[0].ID, "Task in Project 1")
	}
	if len(cols2) > 0 {
		testutil.CreateTestTask(t, db, cols2[0].ID, "Task in Project 2")
	}
	if len(cols3) > 0 {
		testutil.CreateTestTask(t, db, cols3[0].ID, "Task in Project 3")
	}

	// Set terminal width large enough to show detail panel
	m.UIState.SetWidth(200)

	// Initialize detail cache
	m.DetailCache = state.NewTaskDetailCache(10)

	// Populate cache with task from project 1
	m.DetailCache.Set(100, &models.TaskDetail{ID: 100, Title: "Project 1 Task"})
	if !m.DetailCache.Has(100) {
		t.Fatal("Cache should have task 100")
	}

	// Switch to project 2 (index 0 -> 1)
	updatedModel, cmd := m.handleNextProject()
	m = updatedModel.(Model)

	if m.AppState.SelectedProject() != 1 {
		t.Errorf("After first next: SelectedProject = %d, want 1", m.AppState.SelectedProject())
	}
	if m.DetailCache.Has(100) {
		t.Error("Cache should be cleared after first switch")
	}
	if cmd == nil {
		t.Error("First switch should return prefetch cmd")
	}

	// Populate cache with task from project 2
	m.DetailCache.Set(200, &models.TaskDetail{ID: 200, Title: "Project 2 Task"})

	// Switch to project 3 (index 1 -> 2)
	updatedModel, cmd = m.handleNextProject()
	m = updatedModel.(Model)

	if m.AppState.SelectedProject() != 2 {
		t.Errorf("After second next: SelectedProject = %d, want 2", m.AppState.SelectedProject())
	}
	if m.DetailCache.Has(200) {
		t.Error("Cache should be cleared after second switch")
	}
	if cmd == nil {
		t.Error("Second switch should return prefetch cmd")
	}

	// Populate cache with task from project 3
	m.DetailCache.Set(300, &models.TaskDetail{ID: 300, Title: "Project 3 Task"})

	// Switch back to project 2 (index 2 -> 1)
	updatedModel, cmd = m.handlePrevProject()
	m = updatedModel.(Model)

	if m.AppState.SelectedProject() != 1 {
		t.Errorf("After first prev: SelectedProject = %d, want 1", m.AppState.SelectedProject())
	}
	if m.DetailCache.Has(300) {
		t.Error("Cache should be cleared after third switch")
	}
	if cmd == nil {
		t.Error("Third switch should return prefetch cmd")
	}

	// Populate cache again
	m.DetailCache.Set(400, &models.TaskDetail{ID: 400, Title: "Back to Project 2"})

	// Switch back to project 1 (index 1 -> 0)
	updatedModel, cmd = m.handlePrevProject()
	m = updatedModel.(Model)

	if m.AppState.SelectedProject() != 0 {
		t.Errorf("After second prev: SelectedProject = %d, want 0", m.AppState.SelectedProject())
	}
	if m.DetailCache.Has(400) {
		t.Error("Cache should be cleared after fourth switch")
	}
	if cmd == nil {
		t.Error("Fourth switch should return prefetch cmd")
	}

	t.Logf("✓ Round-trip project switching works correctly")
	t.Logf("✓ Cache cleared on each switch: 4/4 switches")
	t.Logf("✓ Prefetch cmd returned on each switch: 4/4 switches")
	t.Logf("✓ Final project: %d (back to start)", m.AppState.SelectedProject())
}
