package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// setupTestModel creates a Model with test data for handlers testing.
// No database connection needed for pure state transformations.
func setupTestModel(columns []*models.Column, tasks map[int][]*models.TaskSummary) Model {
	cfg := &config.Config{
		KeyMappings: config.DefaultKeyMappings(),
	}

	return Model{
		Ctx:               context.Background(),
		App:               nil, // No app needed for navigation handlers
		Config:            cfg,
		AppState:          state.NewAppState(nil, 0, columns, tasks, nil),
		UiState:           state.NewUIState(),
		InputState:        state.NewInputState(),
		FormState:         state.NewFormState(),
		LabelPickerState:  state.NewLabelPickerState(),
		ParentPickerState: state.NewTaskPickerState(),
		ChildPickerState:  state.NewTaskPickerState(),
		NotificationState: state.NewNotificationState(),
		SearchState:       state.NewSearchState(),
		ListViewState:     state.NewListViewState(),
		StatusPickerState: state.NewStatusPickerState(),
	}
}

// TestHandleNavigateLeft_FirstColumn ensures left navigation at column 0 is safe.
// Edge case: User presses 'h' or left arrow when already at first column.
// Security value: No change, no panic (selection stays at 0).
func TestHandleNavigateLeft_FirstColumn(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
	}
	m := setupTestModel(columns, nil)
	m.UiState.SetSelectedColumn(0) // Already at first column
	m.UiState.SetSelectedTask(5)   // Some task selected

	newModel, _ := m.handleNavigateLeft()
	m = newModel.(Model)

	// Should not move left
	if m.UiState.SelectedColumn() != 0 {
		t.Errorf("SelectedColumn after navigate left from 0 = %d, want 0", m.UiState.SelectedColumn())
	}

	// Task selection should remain unchanged (only resets when actually moving)
	if m.UiState.SelectedTask() != 5 {
		t.Errorf("SelectedTask after no-op navigate left = %d, want 5 (unchanged)", m.UiState.SelectedTask())
	}
}

// TestHandleNavigateRight_LastColumn ensures right navigation at last column is safe.
// Edge case: User presses 'l' or right arrow when already at final column.
// Security value: No change, no panic.
func TestHandleNavigateRight_LastColumn(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
		{ID: 3, Name: "Col3"},
	}
	m := setupTestModel(columns, nil)
	m.UiState.SetSelectedColumn(2) // At last column (index 2 of 3 columns)
	m.UiState.SetSelectedTask(3)

	newModel, _ := m.handleNavigateRight()
	m = newModel.(Model)

	// Should not move right
	if m.UiState.SelectedColumn() != 2 {
		t.Errorf("SelectedColumn after navigate right from last = %d, want 2", m.UiState.SelectedColumn())
	}

	// Task selection should remain unchanged
	if m.UiState.SelectedTask() != 3 {
		t.Errorf("SelectedTask after no-op navigate right = %d, want 3 (unchanged)", m.UiState.SelectedTask())
	}
}

// TestHandleNavigateUp_FirstTask ensures up navigation at task 0 is safe.
// Edge case: User presses 'k' or up arrow when already at first task.
// Security value: No change, no panic.
func TestHandleNavigateUp_FirstTask(t *testing.T) {
	m := setupTestModel(nil, nil)
	m.UiState.SetSelectedTask(0) // Already at first task

	newModel, _ := m.handleNavigateUp()
	m = newModel.(Model)

	// Should not move up
	if m.UiState.SelectedTask() != 0 {
		t.Errorf("SelectedTask after navigate up from 0 = %d, want 0", m.UiState.SelectedTask())
	}
}

// TestHandleNavigateDown_LastTask ensures down navigation at final task is safe.
// Edge case: User presses 'j' or down arrow when at last task in column.
// Security value: No change, no panic.
func TestHandleNavigateDown_LastTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
			{ID: 3, Title: "Task 3"},
		},
	}
	m := setupTestModel(columns, tasks)
	m.UiState.SetSelectedColumn(0)
	m.UiState.SetSelectedTask(2) // At last task (index 2 of 3 tasks)

	newModel, _ := m.handleNavigateDown()
	m = newModel.(Model)

	// Should not move down
	if m.UiState.SelectedTask() != 2 {
		t.Errorf("SelectedTask after navigate down from last = %d, want 2", m.UiState.SelectedTask())
	}
}

// TestHandleNavigateRight_ResetsTaskSelection ensures column change resets task to 0.
// Edge case: User navigates to different column while task 5 is selected.
// Security value: Prevents stale task index (new column may have fewer tasks).
func TestHandleNavigateRight_ResetsTaskSelection(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
	}
	m := setupTestModel(columns, nil)
	m.UiState.SetSelectedColumn(0)
	m.UiState.SetSelectedTask(5) // Some task selected

	newModel, _ := m.handleNavigateRight()
	m = newModel.(Model)

	// Should move to column 1
	if m.UiState.SelectedColumn() != 1 {
		t.Errorf("SelectedColumn after navigate right = %d, want 1", m.UiState.SelectedColumn())
	}

	// Task selection should reset to 0
	if m.UiState.SelectedTask() != 0 {
		t.Errorf("SelectedTask after navigate right = %d, want 0 (reset)", m.UiState.SelectedTask())
	}
}

// TestHandleAddTask_NoColumns ensures add task with no columns shows error.
// Edge case: User presses 'a' when no columns exist.
// Security value: Shows error, doesn't crash.
func TestHandleAddTask_NoColumns(t *testing.T) {
	m := setupTestModel([]*models.Column{}, nil) // No columns

	newModel, _ := m.handleAddTask()
	m = newModel.(Model)

	// Should set notification
	if !m.NotificationState.HasAny() {
		t.Error("handleAddTask with no columns should set notification, but HasAny() = false")
	}

	expectedError := "Cannot add task: No columns exist"
	notifications := m.NotificationState.All()
	if len(notifications) == 0 || !strings.Contains(notifications[0].Message, expectedError) {
		t.Errorf("Notification message = %q, want to contain %q", notifications[0].Message, expectedError)
	}

	// Should not change mode
	if m.UiState.Mode() != state.NormalMode {
		t.Errorf("Mode after add task with no columns = %v, want NormalMode", m.UiState.Mode())
	}
}

// TestHandleEditTask_NoTask ensures edit task with no task selected shows error.
// Edge case: User presses 'e' when column is empty or no task selected.
// Security value: Shows error, doesn't crash.
func TestHandleEditTask_NoTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Empty Column"}}
	tasks := map[int][]*models.TaskSummary{1: {}} // Empty tasks
	m := setupTestModel(columns, tasks)
	m.UiState.SetSelectedColumn(0)
	m.UiState.SetSelectedTask(0)

	newModel, _ := m.handleEditTask()
	m = newModel.(Model)

	// Should set notification (getCurrentTask returns nil, so notification is set)
	if !m.NotificationState.HasAny() {
		t.Error("handleEditTask with no task should set notification, but HasAny() = false")
	}

	expectedError := "No task selected to edit"
	notifications := m.NotificationState.All()
	if len(notifications) == 0 || notifications[0].Message != expectedError {
		t.Errorf("Notification message = %q, want %q", notifications[0].Message, expectedError)
	}

	// Should not change mode
	if m.UiState.Mode() != state.NormalMode {
		t.Errorf("Mode after edit with no task = %v, want NormalMode", m.UiState.Mode())
	}
}

// TestHandleDeleteTask_NoTask ensures delete task with no task selected shows error.
// Edge case: User presses 'd' when no task is selected.
// Security value: Shows error, doesn't crash.
func TestHandleDeleteTask_NoTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Empty"}}
	tasks := map[int][]*models.TaskSummary{1: {}}
	m := setupTestModel(columns, tasks)

	newModel, _ := m.handleDeleteTask()
	m = newModel.(Model)

	// Should set notification
	if !m.NotificationState.HasAny() {
		t.Error("handleDeleteTask with no task should set notification, but HasAny() = false")
	}

	expectedError := "No task selected to delete"
	notifications := m.NotificationState.All()
	if len(notifications) == 0 || notifications[0].Message != expectedError {
		t.Errorf("Notification message = %q, want %q", notifications[0].Message, expectedError)
	}

	// Should not enter delete confirm mode
	if m.UiState.Mode() != state.NormalMode {
		t.Errorf("Mode after delete with no task = %v, want NormalMode", m.UiState.Mode())
	}
}

// TestHandleScrollRight_SelectionFollows ensures scroll pushes selection into view.
// Edge case: Viewport scrolls right, pushing selected column out of view.
// Security value: Selection remains visible (auto-adjusts to viewport).
func TestHandleScrollRight_SelectionFollows(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
		{ID: 3, Name: "Col3"},
		{ID: 4, Name: "Col4"},
		{ID: 5, Name: "Col5"},
	}
	m := setupTestModel(columns, nil)
	m.UiState.SetWidth(100) // Viewport size will be 2 columns
	m.UiState.SetSelectedColumn(0)
	m.UiState.SetViewportOffset(0) // Showing columns 0-1

	// Scroll right - viewport becomes 1-2, selection at 0 is now out of view
	newModel, _ := m.handleScrollRight()
	m = newModel.(Model)

	// Viewport should have scrolled
	if m.UiState.ViewportOffset() != 1 {
		t.Errorf("ViewportOffset after scroll right = %d, want 1", m.UiState.ViewportOffset())
	}

	// Selection should follow viewport (adjust to 1, the new leftmost visible column)
	if m.UiState.SelectedColumn() != 1 {
		t.Errorf("SelectedColumn after scroll right = %d, want 1 (adjusted to viewport)", m.UiState.SelectedColumn())
	}

	// Task should reset to 0 when selection changes
	if m.UiState.SelectedTask() != 0 {
		t.Errorf("SelectedTask after scroll adjustment = %d, want 0", m.UiState.SelectedTask())
	}
}
