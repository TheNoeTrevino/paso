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
		Ctx:      context.Background(),
		App:      nil, // No app needed for navigation handlers
		Config:   cfg,
		AppState: state.NewAppState(nil, 0, columns, tasks, nil),
		UIState:  state.NewUIState(),
		Pickers:  state.NewPickerStates(),
		Forms:    state.NewFormStates(),
		UI:       state.NewUIElements(),
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
	m.UIState.SelectedColumn = 0 // Already at first column
	m.UIState.SelectedTask = 5   // Some task selected

	newModel, _ := m.handleNavigateLeft()
	m = newModel.(Model)

	// Should not move left
	if m.UIState.SelectedColumn != 0 {
		t.Errorf("SelectedColumn after navigate left from 0 = %d, want 0", m.UIState.SelectedColumn)
	}

	// Task selection should remain unchanged (only resets when actually moving)
	if m.UIState.SelectedTask != 5 {
		t.Errorf("SelectedTask after no-op navigate left = %d, want 5 (unchanged)", m.UIState.SelectedTask)
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
	m.UIState.SelectedColumn = 2 // At last column (index 2 of 3 columns)
	m.UIState.SelectedTask = 3

	newModel, _ := m.handleNavigateRight()
	m = newModel.(Model)

	// Should not move right
	if m.UIState.SelectedColumn != 2 {
		t.Errorf("SelectedColumn after navigate right from last = %d, want 2", m.UIState.SelectedColumn)
	}

	// Task selection should remain unchanged
	if m.UIState.SelectedTask != 3 {
		t.Errorf("SelectedTask after no-op navigate right = %d, want 3 (unchanged)", m.UIState.SelectedTask)
	}
}

// TestHandleNavigateUp_FirstTask ensures up navigation at task 0 is safe.
// Edge case: User presses 'k' or up arrow when already at first task.
// Security value: No change, no panic.
func TestHandleNavigateUp_FirstTask(t *testing.T) {
	m := setupTestModel(nil, nil)
	m.UIState.SelectedTask = 0 // Already at first task

	newModel, _ := m.handleNavigateUp()
	m = newModel.(Model)

	// Should not move up
	if m.UIState.SelectedTask != 0 {
		t.Errorf("SelectedTask after navigate up from 0 = %d, want 0", m.UIState.SelectedTask)
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
	m.UIState.SelectedColumn = 0
	m.UIState.SelectedTask = 2 // At last task (index 2 of 3 tasks)

	newModel, _ := m.handleNavigateDown()
	m = newModel.(Model)

	// Should not move down
	if m.UIState.SelectedTask != 2 {
		t.Errorf("SelectedTask after navigate down from last = %d, want 2", m.UIState.SelectedTask)
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
	m.UIState.SelectedColumn = 0
	m.UIState.SelectedTask = 5 // Some task selected

	newModel, _ := m.handleNavigateRight()
	m = newModel.(Model)

	// Should move to column 1
	if m.UIState.SelectedColumn != 1 {
		t.Errorf("SelectedColumn after navigate right = %d, want 1", m.UIState.SelectedColumn)
	}

	// Task selection should reset to 0
	if m.UIState.SelectedTask != 0 {
		t.Errorf("SelectedTask after navigate right = %d, want 0 (reset)", m.UIState.SelectedTask)
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
	if !m.UI.Notification.HasAny() {
		t.Error("handleAddTask with no columns should set notification, but HasAny() = false")
	}

	expectedError := "Cannot add task: No columns exist"
	notifications := m.UI.Notification.All()
	if len(notifications) == 0 || !strings.Contains(notifications[0].Message, expectedError) {
		t.Errorf("Notification message = %q, want to contain %q", notifications[0].Message, expectedError)
	}

	// Should not change mode
	if m.UIState.Mode != state.NormalMode {
		t.Errorf("Mode after add task with no columns = %v, want NormalMode", m.UIState.Mode)
	}
}

// TestHandleEditTask_NoTask ensures edit task with no task selected shows error.
// Edge case: User presses 'e' when column is empty or no task selected.
// Security value: Shows error, doesn't crash.
func TestHandleEditTask_NoTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Empty Column"}}
	tasks := map[int][]*models.TaskSummary{1: {}} // Empty tasks
	m := setupTestModel(columns, tasks)
	m.UIState.SelectedColumn = 0
	m.UIState.SelectedTask = 0

	newModel, _ := m.handleEditTask()
	m = newModel.(Model)

	// Should set notification (getCurrentTask returns nil, so notification is set)
	if !m.UI.Notification.HasAny() {
		t.Error("handleEditTask with no task should set notification, but HasAny() = false")
	}

	expectedError := "No task selected to edit"
	notifications := m.UI.Notification.All()
	if len(notifications) == 0 || notifications[0].Message != expectedError {
		t.Errorf("Notification message = %q, want %q", notifications[0].Message, expectedError)
	}

	// Should not change mode
	if m.UIState.Mode != state.NormalMode {
		t.Errorf("Mode after edit with no task = %v, want NormalMode", m.UIState.Mode)
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
	if !m.UI.Notification.HasAny() {
		t.Error("handleDeleteTask with no task should set notification, but HasAny() = false")
	}

	expectedError := "No task selected to delete"
	notifications := m.UI.Notification.All()
	if len(notifications) == 0 || notifications[0].Message != expectedError {
		t.Errorf("Notification message = %q, want %q", notifications[0].Message, expectedError)
	}

	// Should not enter delete confirm mode
	if m.UIState.Mode != state.NormalMode {
		t.Errorf("Mode after delete with no task = %v, want NormalMode", m.UIState.Mode)
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
	m.UIState.SetWidth(100) // Viewport size will be 2 columns
	m.UIState.SelectedColumn = 0
	m.UIState.ViewportOffset = 0 // Showing columns 0-1

	// Scroll right - viewport becomes 1-2, selection at 0 is now out of view
	newModel, _ := m.handleScrollRight()
	m = newModel.(Model)

	// Viewport should have scrolled
	if m.UIState.ViewportOffset != 1 {
		t.Errorf("ViewportOffset after scroll right = %d, want 1", m.UIState.ViewportOffset)
	}

	// Selection should follow viewport (adjust to 1, the new leftmost visible column)
	if m.UIState.SelectedColumn != 1 {
		t.Errorf("SelectedColumn after scroll right = %d, want 1 (adjusted to viewport)", m.UIState.SelectedColumn)
	}

	// Task should reset to 0 when selection changes
	if m.UIState.SelectedTask != 0 {
		t.Errorf("SelectedTask after scroll adjustment = %d, want 0", m.UIState.SelectedTask)
	}
}

// TestHandleNavigateUp_MovesUp ensures up navigation moves from task 1 to task 0.
// Edge case: User presses 'k' or up arrow at task index 1.
// Security value: Selection decrements correctly (board view).
func TestHandleNavigateUp_MovesUp(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
			{ID: 3, Title: "Task 3"},
		},
	}
	m := setupTestModel(columns, tasks)
	m.UIState.SelectedColumn = 0
	m.UIState.SelectedTask = 1 // At second task

	newModel, _ := m.handleNavigateUp()
	m = newModel.(Model)

	// Should move up to task 0
	if m.UIState.SelectedTask != 0 {
		t.Errorf("SelectedTask after navigate up = %d, want 0", m.UIState.SelectedTask)
	}
}

// TestHandleNavigateDown_MovesDown ensures down navigation moves from task 0 to task 1.
// Edge case: User presses 'j' or down arrow at task index 0.
// Security value: Selection increments correctly (board view).
func TestHandleNavigateDown_MovesDown(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
			{ID: 3, Title: "Task 3"},
		},
	}
	m := setupTestModel(columns, tasks)
	m.UIState.SelectedColumn = 0
	m.UIState.SelectedTask = 0 // At first task

	newModel, _ := m.handleNavigateDown()
	m = newModel.(Model)

	// Should move down to task 1
	if m.UIState.SelectedTask != 1 {
		t.Errorf("SelectedTask after navigate down = %d, want 1", m.UIState.SelectedTask)
	}
}

// TestHandleNavigateLeft_MovesLeft ensures left navigation moves from column 1 to column 0.
// Edge case: User presses 'h' or left arrow at column index 1.
// Security value: Selection decrements correctly and task resets.
func TestHandleNavigateLeft_MovesLeft(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
	}
	m := setupTestModel(columns, nil)
	m.UIState.SelectedColumn = 1 // At second column
	m.UIState.SelectedTask = 5   // Some task selected

	newModel, _ := m.handleNavigateLeft()
	m = newModel.(Model)

	// Should move to column 0
	if m.UIState.SelectedColumn != 0 {
		t.Errorf("SelectedColumn after navigate left = %d, want 0", m.UIState.SelectedColumn)
	}

	// Task selection should reset to 0
	if m.UIState.SelectedTask != 0 {
		t.Errorf("SelectedTask after navigate left = %d, want 0 (reset)", m.UIState.SelectedTask)
	}
}

// TestHandleScrollRight_AtRightmostView_NoOp ensures scroll right at rightmost view is safe.
// Edge case: User presses scroll right when viewport is already at the rightmost position.
// Security value: No change, no panic, notification shown.
func TestHandleScrollRight_AtRightmostView_NoOp(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
		{ID: 3, Name: "Col3"},
	}
	m := setupTestModel(columns, nil)
	m.UIState.SetWidth(100)      // Viewport size will be 2 columns
	m.UIState.ViewportOffset = 1 // Already showing columns 1-2 (rightmost)
	m.UIState.SelectedColumn = 1

	newModel, _ := m.handleScrollRight()
	m = newModel.(Model)

	// Viewport should not have scrolled
	if m.UIState.ViewportOffset != 1 {
		t.Errorf("ViewportOffset after scroll right at rightmost = %d, want 1 (no change)", m.UIState.ViewportOffset)
	}

	// Should set notification
	if !m.UI.Notification.HasAny() {
		t.Error("handleScrollRight at rightmost should set notification, but HasAny() = false")
	}
}

// TestHandleScrollLeft_MovesLeft ensures scroll left moves viewport offset.
// Edge case: User presses scroll left when viewport offset is 1.
// Security value: Viewport offset decrements correctly.
func TestHandleScrollLeft_MovesLeft(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
		{ID: 3, Name: "Col3"},
		{ID: 4, Name: "Col4"},
	}
	m := setupTestModel(columns, nil)
	m.UIState.SetWidth(100)      // Viewport size will be 2 columns
	m.UIState.ViewportOffset = 1 // Showing columns 1-2
	m.UIState.SelectedColumn = 1

	newModel, _ := m.handleScrollLeft()
	m = newModel.(Model)

	// Viewport should have scrolled left
	if m.UIState.ViewportOffset != 0 {
		t.Errorf("ViewportOffset after scroll left = %d, want 0", m.UIState.ViewportOffset)
	}
}

// TestHandleScrollLeft_AtLeftmostView_NoOp ensures scroll left at leftmost view is safe.
// Edge case: User presses scroll left when viewport offset is already 0.
// Security value: No change, no panic, notification shown.
func TestHandleScrollLeft_AtLeftmostView_NoOp(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
		{ID: 3, Name: "Col3"},
	}
	m := setupTestModel(columns, nil)
	m.UIState.SetWidth(100)      // Viewport size will be 2 columns
	m.UIState.ViewportOffset = 0 // Already at leftmost
	m.UIState.SelectedColumn = 0

	newModel, _ := m.handleScrollLeft()
	m = newModel.(Model)

	// Viewport should not have scrolled
	if m.UIState.ViewportOffset != 0 {
		t.Errorf("ViewportOffset after scroll left at leftmost = %d, want 0 (no change)", m.UIState.ViewportOffset)
	}

	// Should set notification
	if !m.UI.Notification.HasAny() {
		t.Error("handleScrollLeft at leftmost should set notification, but HasAny() = false")
	}
}
