package tui

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, 0, m.UIState.SelectedColumn)

	// Task selection should remain unchanged (only resets when actually moving)
	assert.Equal(t, 5, m.UIState.SelectedTask)
}

// TestHandleNavigateRight_LastColumn ensures right navigation at last column is safe.
// Edge case: User presses 'l' or right arrow when already at final column.
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
	assert.Equal(t, 2, m.UIState.SelectedColumn)

	// Task selection should remain unchanged
	assert.Equal(t, 3, m.UIState.SelectedTask)
}

// TestHandleNavigateUp_FirstTask ensures up navigation at task 0 is safe.
// Edge case: User presses 'k' or up arrow when already at first task.
func TestHandleNavigateUp_FirstTask(t *testing.T) {
	m := setupTestModel(nil, nil)
	m.UIState.SelectedTask = 0 // Already at first task

	newModel, _ := m.handleNavigateUp()
	m = newModel.(Model)

	// Should not move up
	assert.Equal(t, 0, m.UIState.SelectedTask)
}

// TestHandleNavigateDown_LastTask ensures down navigation at final task is safe.
// Edge case: User presses 'j' or down arrow when at last task in column.
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
	assert.Equal(t, 2, m.UIState.SelectedTask)
}

// TestHandleNavigateRight_ResetsTaskSelection ensures column change resets task to 0.
// Edge case: User navigates to different column while task 5 is selected.
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
	assert.Equal(t, 1, m.UIState.SelectedColumn)

	// Task selection should reset to 0
	assert.Equal(t, 0, m.UIState.SelectedTask)
}

// TestHandleAddTask_NoColumns ensures add task with no columns shows error.
// Edge case: User presses 'a' when no columns exist.
func TestHandleAddTask_NoColumns(t *testing.T) {
	m := setupTestModel([]*models.Column{}, nil) // No columns

	newModel, _ := m.handleAddTask()
	m = newModel.(Model)

	// Should set notification
	assert.True(t, m.UI.Notification.HasAny(), "handleAddTask with no columns should set notification, but HasAny() = false")

	expectedError := "Cannot add task: No columns exist"
	notifications := m.UI.Notification.All()
	require.NotEmpty(t, notifications)
	assert.Contains(t, notifications[0].Message, expectedError)

	// Should not change mode
	assert.Equal(t, state.NormalMode, m.UIState.Mode)
}

// TestHandleEditTask_NoTask ensures edit task with no task selected shows error.
// Edge case: User presses 'e' when column is empty or no task selected.
func TestHandleEditTask_NoTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Empty Column"}}
	tasks := map[int][]*models.TaskSummary{1: {}} // Empty tasks
	m := setupTestModel(columns, tasks)
	m.UIState.SelectedColumn = 0
	m.UIState.SelectedTask = 0

	newModel, _ := m.handleEditTask()
	m = newModel.(Model)

	// Should set notification (getCurrentTask returns nil, so notification is set)
	assert.True(t, m.UI.Notification.HasAny(), "handleEditTask with no task should set notification, but HasAny() = false")

	expectedError := "No task selected to edit"
	notifications := m.UI.Notification.All()
	require.NotEmpty(t, notifications)
	assert.Equal(t, expectedError, notifications[0].Message)

	// Should not change mode
	assert.Equal(t, state.NormalMode, m.UIState.Mode)
}

// TestHandleDeleteTask_NoTask ensures delete task with no task selected shows error.
// Edge case: User presses 'd' when no task is selected.
func TestHandleDeleteTask_NoTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Empty"}}
	tasks := map[int][]*models.TaskSummary{1: {}}
	m := setupTestModel(columns, tasks)

	newModel, _ := m.handleDeleteTask()
	m = newModel.(Model)

	// Should set notification
	assert.True(t, m.UI.Notification.HasAny(), "handleDeleteTask with no task should set notification, but HasAny() = false")

	expectedError := "No task selected to delete"
	notifications := m.UI.Notification.All()
	require.NotEmpty(t, notifications)
	assert.Equal(t, expectedError, notifications[0].Message)

	// Should not enter delete confirm mode
	assert.Equal(t, state.NormalMode, m.UIState.Mode)
}

// TestHandleScrollRight_SelectionFollows ensures scroll pushes selection into view.
// Edge case: Viewport scrolls right, pushing selected column out of view.
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
	assert.Equal(t, 1, m.UIState.ViewportOffset)

	// Selection should follow viewport (adjust to 1, the new leftmost visible column)
	assert.Equal(t, 1, m.UIState.SelectedColumn)

	// Task should reset to 0 when selection changes
	assert.Equal(t, 0, m.UIState.SelectedTask)
}

// TestHandleNavigateUp_MovesUp ensures up navigation moves from task 1 to task 0.
// Edge case: User presses 'k' or up arrow at task index 1.
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
	assert.Equal(t, 0, m.UIState.SelectedTask)
}

// TestHandleNavigateDown_MovesDown ensures down navigation moves from task 0 to task 1.
// Edge case: User presses 'j' or down arrow at task index 0.
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
	assert.Equal(t, 1, m.UIState.SelectedTask)
}

// TestHandleNavigateLeft_MovesLeft ensures left navigation moves from column 1 to column 0.
// Edge case: User presses 'h' or left arrow at column index 1.
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
	assert.Equal(t, 0, m.UIState.SelectedColumn)

	// Task selection should reset to 0
	assert.Equal(t, 0, m.UIState.SelectedTask)
}

// TestHandleScrollRight_AtRightmostView_NoOp ensures scroll right at rightmost view is safe.
// Edge case: User presses scroll right when viewport is already at the rightmost position.
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
	assert.Equal(t, 1, m.UIState.ViewportOffset)

	// Should set notification
	assert.True(t, m.UI.Notification.HasAny(), "handleScrollRight at rightmost should set notification, but HasAny() = false")
}

// TestHandleScrollLeft_MovesLeft ensures scroll left moves viewport offset.
// Edge case: User presses scroll left when viewport offset is 1.
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
	assert.Equal(t, 0, m.UIState.ViewportOffset)
}

// TestHandleScrollLeft_AtLeftmostView_NoOp ensures scroll left at leftmost view is safe.
// Edge case: User presses scroll left when viewport offset is already 0.
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
	assert.Equal(t, 0, m.UIState.ViewportOffset)

	// Should set notification
	assert.True(t, m.UI.Notification.HasAny(), "handleScrollLeft at leftmost should set notification, but HasAny() = false")
}
