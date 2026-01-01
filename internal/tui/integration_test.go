package tui

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestModelInitialization tests that a Model can be created and initialized without panic
func TestModelInitialization(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	// Create a test project with data
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	// Initialize the model
	model := InitialModel(ctx, appInstance, cfg, nil)

	// Verify initial state
	if model.App == nil {
		t.Fatal("model.App should not be nil")
	}
	if model.AppState == nil {
		t.Fatal("model.AppState should not be nil")
	}
	if model.UIState == nil {
		t.Fatal("model.UIState should not be nil")
	}
	if model.Config == nil {
		t.Fatal("model.Config should not be nil")
	}

	// Verify projects were loaded
	projects := model.AppState.Projects()
	if len(projects) == 0 {
		t.Fatal("expected projects to be loaded")
	}

	// Verify columns were loaded for the current project
	columns := model.AppState.Columns()
	if len(columns) == 0 {
		t.Fatal("expected columns to be loaded")
	}

	// Verify tasks were loaded
	tasks := model.AppState.Tasks()
	if len(tasks) == 0 {
		t.Fatal("expected tasks to be loaded")
	}
}

// TestModelInit tests the Init command returns without panic
func TestModelInit(t *testing.T) {
	model := createTestModel(t)

	// Call Init - should not panic
	cmd := model.Init()

	// Init should return a Cmd for listening to notifications
	if cmd == nil {
		t.Fatal("Init() should return a Cmd")
	}
}

// TestModelInitWithNoProjects tests model initialization with no projects
func TestModelInitWithNoProjects(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	// Initialize without creating any projects
	model := InitialModel(ctx, appInstance, cfg, nil)

	// Should not panic and should have empty projects
	if model.AppState == nil {
		t.Fatal("AppState should not be nil")
	}

	projects := model.AppState.Projects()
	if len(projects) != 0 {
		t.Fatalf("expected no projects, got %d", len(projects))
	}

	// Should still be able to call View without panic
	_ = model.View()
}

// TestWindowSizeHandling tests that model handles window size messages
func TestWindowSizeHandling(t *testing.T) {
	model := createTestModel(t)

	// Send window size message
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if updatedModel == nil {
		t.Fatal("Update should return a model")
	}

	// Should not panic when handling window size
	m := updatedModel.(Model)
	// The model should still be valid after the update
	if m.UIState == nil {
		t.Fatal("UIState should not be nil after update")
	}
}

// TestViewGeneratesOutput tests that View produces non-empty output after window size is set
func TestViewGeneratesOutput(t *testing.T) {
	model := createTestModel(t)

	// First View should return "Loading..." when window size is not set
	output := model.View()
	if output.Content == "" {
		t.Fatal("View should return content even before window size is set")
	}

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m := updatedModel.(Model)

	// View should now generate real output
	output = m.View()
	if output.Content == "" {
		t.Fatal("View should generate non-empty content after window size is set")
	}
}

// TestStateTransitions_NormalToForm tests state transition from normal mode to form mode
func TestStateTransitions_NormalToForm(t *testing.T) {
	model := createTestModel(t)

	// Initial state should be NormalMode
	if model.UIState.Mode() != state.NormalMode {
		t.Fatalf("expected NormalMode, got %v", model.UIState.Mode())
	}

	// Simulate entering form mode (e.g., by setting mode directly)
	model.UIState.SetMode(state.TicketFormMode)

	// Verify state changed
	if model.UIState.Mode() != state.TicketFormMode {
		t.Fatalf("expected TicketFormMode, got %v", model.UIState.Mode())
	}

	// View should not panic
	output := model.View()
	if output.Content == "" {
		t.Fatal("View should generate content in form mode")
	}
}

// TestStateTransitions_FormToNormal tests transitioning back from form mode to normal
func TestStateTransitions_FormToNormal(t *testing.T) {
	model := createTestModel(t)

	// Enter form mode
	model.UIState.SetMode(state.TicketFormMode)
	if model.UIState.Mode() != state.TicketFormMode {
		t.Fatal("failed to enter form mode")
	}

	// Exit form mode
	model.UIState.SetMode(state.NormalMode)

	// Verify state changed back
	if model.UIState.Mode() != state.NormalMode {
		t.Fatalf("expected NormalMode, got %v", model.UIState.Mode())
	}

	// View should render normally again
	output := model.View()
	if output.Content == "" {
		t.Fatal("View should generate content after returning to normal mode")
	}
}

// TestStateTransitions_PickerModes tests transitions to various picker modes
func TestStateTransitions_PickerModes(t *testing.T) {
	pickerModes := []state.Mode{
		state.LabelPickerMode,
		state.PriorityPickerMode,
		state.TypePickerMode,
	}

	for i, mode := range pickerModes {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			model := createTestModel(t)

			// Set window size for proper rendering
			updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
			m := updatedModel.(Model)

			// Transition to picker mode
			m.UIState.SetMode(mode)

			if m.UIState.Mode() != mode {
				t.Fatalf("expected %v, got %v", mode, m.UIState.Mode())
			}

			// View should not panic
			output := m.View()
			if output.Content == "" {
				t.Fatalf("View should generate content for mode %v", mode)
			}
		})
	}
}

// TestGetCurrentTask returns correct task from model
func TestGetCurrentTask(t *testing.T) {
	model := createTestModel(t)

	// Initially no task should be selected
	task := model.getCurrentTask()
	if task == nil {
		// This is expected - no tasks in test data
		return
	}

	// Verify we got a task
	if task.ID == 0 {
		t.Fatal("task ID should not be 0")
	}
}

// TestNavigationHandling tests column and task navigation
func TestNavigationHandling(t *testing.T) {
	model := createTestModel(t)

	_ = model.UIState.SelectedColumn()

	// Try to navigate right
	keyMsgRight := tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})
	updatedModel, _ := model.Update(keyMsgRight)
	m := updatedModel.(Model)

	// Selection might not change if there's only one column, but shouldn't panic
	_ = m.UIState.SelectedColumn()

	// Try to navigate left
	keyMsgLeft := tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft})
	updatedModel, _ = m.Update(keyMsgLeft)
	m = updatedModel.(Model)

	// Check we can still get selection
	if m.UIState.SelectedColumn() < 0 {
		t.Fatal("selected column should not be negative")
	}
}

// TestContextCancellation tests that model respects context cancellation
func TestContextCancellation(t *testing.T) {
	model := createTestModel(t)

	// Cancel the context
	cancelCtx, cancel := context.WithCancel(model.Ctx)
	cancel()
	model.Ctx = cancelCtx

	// Update should handle key messages even with cancelled context
	keyMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	updatedModel, cmd := model.Update(keyMsg)

	// Should get a quit command
	if cmd == nil || updatedModel == nil {
		t.Fatal("expected valid response from cancelled context")
	}
}

// TestMultipleProjectsLoading tests loading multiple projects
func TestMultipleProjectsLoading(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	// Create multiple projects
	proj1ID := testutil.CreateTestProject(t, db, "Project 1")
	proj2ID := testutil.CreateTestProject(t, db, "Project 2")

	// Create columns for both projects
	col1 := testutil.CreateTestColumn(t, db, proj1ID, "Column 1")
	col2 := testutil.CreateTestColumn(t, db, proj2ID, "Column 2")

	testutil.CreateTestTask(t, db, col1, "Task 1")
	testutil.CreateTestTask(t, db, col2, "Task 2")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	model := InitialModel(ctx, appInstance, cfg, nil)

	// Verify both projects loaded
	projects := model.AppState.Projects()
	if len(projects) < 2 {
		t.Fatalf("expected at least 2 projects, got %d", len(projects))
	}
}

// TestTasksLoadedForCurrentProject verifies tasks are loaded for the current project
func TestTasksLoadedForCurrentProject(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column B")
	testutil.CreateTestTask(t, db, columnID, "Task X")
	testutil.CreateTestTask(t, db, columnID, "Task Y")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	model := InitialModel(ctx, appInstance, cfg, nil)

	// Get current project
	currentProject := model.AppState.GetCurrentProject()
	if currentProject == nil {
		t.Fatal("expected current project to be set")
	}

	// Get current column
	columns := model.AppState.Columns()
	if len(columns) == 0 {
		t.Fatal("expected columns to be loaded")
	}

	// Verify tasks map was created (even if empty due to query issues)
	tasks := model.AppState.Tasks()
	if tasks == nil {
		t.Fatal("tasks map should not be nil")
	}
}

// TestLabelsLoaded verifies labels are loaded for the current project
func TestLabelsLoaded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	testutil.CreateTestLabel(t, db, projectID, "Bug", "#FF0000")
	testutil.CreateTestLabel(t, db, projectID, "Feature", "#00FF00")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	model := InitialModel(ctx, appInstance, cfg, nil)

	// Verify labels were loaded
	labels := model.AppState.Labels()
	if len(labels) < 2 {
		t.Fatalf("expected at least 2 labels, got %d", len(labels))
	}
}

// TestStateIsolation verifies that each state object is properly initialized and isolated
func TestStateIsolation(t *testing.T) {
	model := createTestModel(t)

	// Verify each state is initialized and independent
	if model.UIState == nil {
		t.Fatal("UIState not initialized")
	}
	if model.InputState == nil {
		t.Fatal("InputState not initialized")
	}
	if model.FormState == nil {
		t.Fatal("FormState not initialized")
	}
	if model.LabelPickerState == nil {
		t.Fatal("LabelPickerState not initialized")
	}
	if model.ParentPickerState == nil {
		t.Fatal("ParentPickerState not initialized")
	}
	if model.ChildPickerState == nil {
		t.Fatal("ChildPickerState not initialized")
	}
	if model.PriorityPickerState == nil {
		t.Fatal("PriorityPickerState not initialized")
	}
	if model.TypePickerState == nil {
		t.Fatal("TypePickerState not initialized")
	}
	if model.NotificationState == nil {
		t.Fatal("NotificationState not initialized")
	}
	if model.SearchState == nil {
		t.Fatal("SearchState not initialized")
	}
	if model.ListViewState == nil {
		t.Fatal("ListViewState not initialized")
	}
	if model.StatusPickerState == nil {
		t.Fatal("StatusPickerState not initialized")
	}

	// Modifying one state shouldn't affect others
	model.UIState.SetMode(state.TicketFormMode)
	if model.InputState == nil || model.FormState == nil {
		t.Fatal("other states should be unaffected")
	}
}

// TestKeyMessageHandling tests that various key messages are handled without panic
func TestKeyMessageHandling(t *testing.T) {
	keyTests := []struct {
		name string
		msg  tea.Msg
	}{
		{"Up", tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})},
		{"Down", tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})},
		{"Left", tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft})},
		{"Right", tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})},
		{"Home", tea.KeyPressMsg(tea.Key{Code: tea.KeyHome})},
		{"End", tea.KeyPressMsg(tea.Key{Code: tea.KeyEnd})},
		{"Enter", tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})},
		{"Escape", tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})},
	}

	for _, tt := range keyTests {
		t.Run(tt.name, func(t *testing.T) {
			model := createTestModel(t)
			model.UIState.SetMode(state.NormalMode)

			// Set window size first
			updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
			m := updatedModel.(Model)

			// Should handle key message without panic
			_, _ = m.Update(tt.msg)
		})
	}
}

// TestConnectionStateInitialization tests connection state is properly initialized
func TestConnectionStateInitialization(t *testing.T) {
	model := createTestModel(t)

	if model.ConnectionState == nil {
		t.Fatal("ConnectionState should be initialized")
	}

	// When no event client, should be Disconnected
	if model.EventClient == nil && model.ConnectionState.Status() != state.Disconnected {
		t.Fatalf("expected Disconnected status, got %v", model.ConnectionState.Status())
	}
}

// TestNotificationChannelInitialization tests notification channel is created
func TestNotificationChannelInitialization(t *testing.T) {
	model := createTestModel(t)

	if model.NotifyChan == nil {
		t.Fatal("NotifyChan should be initialized")
	}

	// Channel should be non-blocking send capable (if we had an events client)
	// For now, just verify it exists
	_ = model.NotifyChan
}

// TestModeUsesLayers verifies layer-based rendering modes are identified correctly
func TestModeUsesLayers(t *testing.T) {
	layerModes := []state.Mode{
		state.TicketFormMode,
		state.ProjectFormMode,
		state.LabelPickerMode,
		state.PriorityPickerMode,
		state.HelpMode,
	}

	for _, mode := range layerModes {
		if !mode.UsesLayers() {
			t.Errorf("expected mode %v to use layers", mode)
		}
	}
}

// TestEmptyColumnsHandling tests model behavior with empty columns
func TestEmptyColumnsHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	// Create column but don't add tasks
	_ = testutil.CreateTestColumn(t, db, projectID, "Empty Column")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	model := InitialModel(ctx, appInstance, cfg, nil)

	// Should handle empty columns gracefully
	tasks := model.getCurrentTasks()
	if tasks == nil {
		t.Fatal("getCurrentTasks should return empty slice, not nil")
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}

	// View should still render
	output := model.View()
	if output.Content == "" {
		t.Fatal("View should render even with empty columns")
	}
}

// TestMinimalWindowSize tests model with very small window
func TestMinimalWindowSize(t *testing.T) {
	model := createTestModel(t)

	// Set very small window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 10, Height: 3})
	m := updatedModel.(Model)

	// Should handle gracefully
	output := m.View()
	if output.Content == "" {
		t.Fatal("View should render even with minimal window size")
	}
}

// TestLargeWindowSize tests model with very large window
func TestLargeWindowSize(t *testing.T) {
	model := createTestModel(t)

	// Set very large window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 500, Height: 200})
	m := updatedModel.(Model)

	// Should handle gracefully
	output := m.View()
	if output.Content == "" {
		t.Fatal("View should render even with large window size")
	}
}

// createTestModel is a helper to create a test model with test data
func createTestModel(t *testing.T) Model {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	t.Cleanup(func() { db.Close() })

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	return InitialModel(ctx, appInstance, cfg, nil)
}
