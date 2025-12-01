package tui

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestGetCurrentTasks_NoColumns ensures task access when columns slice is empty.
// Edge case: User deletes all columns or app starts with no columns in database.
// Security value: Prevents panic accessing columns[0] when columns slice is empty.
func TestGetCurrentTasks_NoColumns(t *testing.T) {
	m := Model{
		appState: state.NewAppState(nil, 0, []*models.Column{}, nil, nil),
		uiState:  state.NewUIState(),
	}

	tasks := m.getCurrentTasks()
	if tasks == nil {
		t.Error("getCurrentTasks() with no columns = nil, want empty slice")
	}
	if len(tasks) != 0 {
		t.Errorf("getCurrentTasks() with no columns length = %d, want 0", len(tasks))
	}
}

// TestGetCurrentTasks_SelectedColumnOutOfBounds ensures task access when selection is invalid.
// Edge case: Corrupted state where selectedColumn >= len(columns).
// Security value: Prevents index out of bounds panic.
func TestGetCurrentTasks_SelectedColumnOutOfBounds(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "Done"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {{ID: 1, Title: "Task 1"}},
		2: {{ID: 2, Title: "Task 2"}},
	}

	m := Model{
		appState: state.NewAppState(nil, 0, columns, tasks, nil),
		uiState:  state.NewUIState(),
	}

	// Set selected column out of bounds
	m.uiState.SetSelectedColumn(5)

	// Should return empty slice safely without panic
	result := m.getCurrentTasks()
	if result == nil {
		t.Error("getCurrentTasks() with out-of-bounds index = nil, want empty slice")
	}
	if len(result) != 0 {
		t.Errorf("getCurrentTasks() with invalid selectedColumn = %d tasks, want 0", len(result))
	}
}

// TestGetCurrentTask_NoTasks ensures task access when column has no tasks.
// Edge case: Selected column exists but has zero tasks.
// Security value: Returns nil safely instead of panicking.
func TestGetCurrentTask_NoTasks(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Empty Column"}}
	tasks := map[int][]*models.TaskSummary{
		1: {}, // Empty tasks for column 1
	}

	m := Model{
		appState: state.NewAppState(nil, 0, columns, tasks, nil),
		uiState:  state.NewUIState(),
	}
	m.uiState.SetSelectedColumn(0)
	m.uiState.SetSelectedTask(0)

	task := m.getCurrentTask()
	if task != nil {
		t.Errorf("getCurrentTask() with no tasks = %v, want nil", task)
	}
}

// TestGetCurrentTask_SelectedTaskOutOfBounds ensures task access when task index is invalid.
// Edge case: selectedTask >= len(tasks) due to corrupted state or race condition.
// Security value: Prevents index out of bounds panic.
func TestGetCurrentTask_SelectedTaskOutOfBounds(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
		},
	}

	m := Model{
		appState: state.NewAppState(nil, 0, columns, tasks, nil),
		uiState:  state.NewUIState(),
	}
	m.uiState.SetSelectedColumn(0)
	m.uiState.SetSelectedTask(5) // Out of bounds

	task := m.getCurrentTask()
	if task != nil {
		t.Errorf("getCurrentTask() with out-of-bounds index = %v, want nil", task)
	}

	// Valid task for comparison
	m.uiState.SetSelectedTask(1)
	task = m.getCurrentTask()
	if task == nil || task.ID != 2 {
		t.Errorf("getCurrentTask() with valid index, got %v, want task with ID=2", task)
	}
}

// TestGetCurrentColumn_EmptyColumns ensures column access with no columns.
// Edge case: User deletes all columns.
// Security value: Returns nil safely instead of panicking.
func TestGetCurrentColumn_EmptyColumns(t *testing.T) {
	m := Model{
		appState: state.NewAppState(nil, 0, []*models.Column{}, nil, nil),
		uiState:  state.NewUIState(),
	}

	col := m.getCurrentColumn()
	if col != nil {
		t.Errorf("getCurrentColumn() with no columns = %v, want nil", col)
	}
}

// TestRemoveCurrentTask_LastTask ensures removal of final task adjusts selection correctly.
// Edge case: Deleting the last task in a column.
// Security value: Selection index adjusts correctly (doesn't go negative or out of bounds).
func TestRemoveCurrentTask_LastTask(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
			{ID: 3, Title: "Task 3"},
		},
	}

	m := Model{
		appState: state.NewAppState(nil, 0, columns, tasks, nil),
		uiState:  state.NewUIState(),
	}
	m.uiState.SetSelectedColumn(0)
	m.uiState.SetSelectedTask(2) // Select last task (index 2)

	// Remove the last task
	m.removeCurrentTask()

	// Should adjust selection to previous task (index 1)
	if m.uiState.SelectedTask() != 1 {
		t.Errorf("SelectedTask after removing last task = %d, want 1", m.uiState.SelectedTask())
	}

	// Verify task was removed
	if len(m.appState.Tasks()[1]) != 2 {
		t.Errorf("Tasks length after removal = %d, want 2", len(m.appState.Tasks()[1]))
	}
}

// TestRemoveCurrentTask_EmptyColumn ensures removal attempt on empty column is safe.
// Edge case: Calling removeCurrentTask when column has no tasks.
// Security value: No-op without panic (no slice operations on empty slice).
func TestRemoveCurrentTask_EmptyColumn(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Empty"}}
	tasks := map[int][]*models.TaskSummary{
		1: {}, // No tasks
	}

	m := Model{
		appState: state.NewAppState(nil, 0, columns, tasks, nil),
		uiState:  state.NewUIState(),
	}
	m.uiState.SetSelectedColumn(0)
	m.uiState.SetSelectedTask(0)

	// Should be no-op
	m.removeCurrentTask()

	// Verify nothing changed
	if len(m.appState.Tasks()[1]) != 0 {
		t.Errorf("Tasks length after remove on empty = %d, want 0", len(m.appState.Tasks()[1]))
	}
}

// TestRemoveCurrentColumn_LastColumn ensures removal of final column adjusts selection and viewport.
// Edge case: Deleting the last remaining column.
// Security value: Selection and viewport adjust correctly (no negative indices).
func TestRemoveCurrentColumn_LastColumn(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
		{ID: 3, Name: "Col3"},
	}

	m := Model{
		appState: state.NewAppState(nil, 0, columns, nil, nil),
		uiState:  state.NewUIState(),
	}
	m.uiState.SetSelectedColumn(2) // Select last column
	m.uiState.SetSelectedTask(5)   // Some task

	// Remove the last column
	m.removeCurrentColumn()

	// Selection should adjust to column 1 (previous column)
	if m.uiState.SelectedColumn() != 1 {
		t.Errorf("SelectedColumn after removing last = %d, want 1", m.uiState.SelectedColumn())
	}

	// Task selection should reset to 0
	if m.uiState.SelectedTask() != 0 {
		t.Errorf("SelectedTask after removeCurrentColumn = %d, want 0", m.uiState.SelectedTask())
	}

	// Verify column was removed
	if len(m.appState.Columns()) != 2 {
		t.Errorf("Columns length after removal = %d, want 2", len(m.appState.Columns()))
	}
}
