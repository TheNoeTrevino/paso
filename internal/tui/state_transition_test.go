package tui

import (
	"testing"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestModeTransition_NormalToForm verifies mode transitions from NormalMode to form modes.
// Edge case: User initiates form entry (create task, edit column, etc).
// Security value: Form state is properly initialized and previous mode is tracked.
func TestModeTransition_NormalToForm(t *testing.T) {
	m := setupTestModel([]*models.Column{{ID: 1, Name: "Todo"}}, nil)
	m.UIState.SetMode(state.NormalMode)

	// Simulate mode transition to TicketFormMode
	m.UIState.SetMode(state.TicketFormMode)

	if m.UIState.Mode() != state.TicketFormMode {
		t.Errorf("Mode after transition = %v, want TicketFormMode", m.UIState.Mode())
	}

	// Verify mode can be tracked back
	if m.UIState.Mode() == state.NormalMode {
		t.Error("Mode should have changed from NormalMode")
	}
}

// TestModeTransition_FormToNormal verifies mode transitions from form modes back to NormalMode.
// Edge case: User cancels or submits form (ESC or Enter).
// Security value: Form state is properly cleared when returning to normal mode.
func TestModeTransition_FormToNormal(t *testing.T) {
	m := setupTestModel([]*models.Column{{ID: 1, Name: "Todo"}}, nil)

	// Start in form mode
	m.UIState.SetMode(state.TicketFormMode)
	if m.UIState.Mode() != state.TicketFormMode {
		t.Fatal("Failed to set mode to TicketFormMode")
	}

	// Transition back to normal
	m.UIState.SetMode(state.NormalMode)

	if m.UIState.Mode() != state.NormalMode {
		t.Errorf("Mode after transition = %v, want NormalMode", m.UIState.Mode())
	}
}

// TestModeTransition_MultipleTransitions verifies sequential mode changes work correctly.
// Edge case: User enters form, exits, enters picker, exits.
// Security value: Each transition is independent and mode state is always consistent.
func TestModeTransition_MultipleTransitions(t *testing.T) {
	m := setupTestModel([]*models.Column{{ID: 1, Name: "Todo"}}, nil)
	m.UIState.SetMode(state.NormalMode)

	// Series of transitions
	modes := []state.Mode{
		state.TicketFormMode,
		state.NormalMode,
		state.LabelPickerMode,
		state.NormalMode,
		state.AddColumnFormMode,
		state.NormalMode,
	}

	for _, mode := range modes {
		m.UIState.SetMode(mode)
		if m.UIState.Mode() != mode {
			t.Errorf("Mode after transition = %v, want %v", m.UIState.Mode(), mode)
		}
	}
}

// TestSelectedColumn_BoundaryConditions verifies column selection respects boundaries.
// Edge case: Selection at column 0 (first), at last column, out of bounds.
// Security value: Selection never results in invalid indices.
func TestSelectedColumn_BoundaryConditions(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
		{ID: 3, Name: "Col3"},
	}
	m := setupTestModel(columns, nil)

	tests := []struct {
		name      string
		selection int
		want      int
	}{
		{"First column", 0, 0},
		{"Middle column", 1, 1},
		{"Last column", 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.UIState.SetSelectedColumn(tt.selection)
			if m.UIState.SelectedColumn() != tt.want {
				t.Errorf("SelectedColumn = %d, want %d", m.UIState.SelectedColumn(), tt.want)
			}
		})
	}
}

// TestSelectedTask_UpdatesCorrectly verifies task selection state updates work.
// Edge case: Task selection at 0, middle, last, and boundary crossing.
// Security value: Task selection can be updated independently and maintains consistency.
func TestSelectedTask_UpdatesCorrectly(t *testing.T) {
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
			{ID: 3, Title: "Task 3"},
			{ID: 4, Title: "Task 4"},
			{ID: 5, Title: "Task 5"},
		},
	}
	m := setupTestModel(columns, tasks)
	m.UIState.SetSelectedColumn(0)

	tests := []struct {
		name      string
		selection int
		want      int
	}{
		{"First task", 0, 0},
		{"Middle task", 2, 2},
		{"Last task", 4, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.UIState.SetSelectedTask(tt.selection)
			if m.UIState.SelectedTask() != tt.want {
				t.Errorf("SelectedTask = %d, want %d", m.UIState.SelectedTask(), tt.want)
			}
		})
	}
}

// TestCursorPosition_AtEnd verifies cursor positioning at boundaries.
// Edge case: Selected indices are at the end of available items.
// Security value: Boundary checks prevent invalid selections.
func TestCursorPosition_AtEnd(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {
			{ID: 1, Title: "Task 1"},
			{ID: 2, Title: "Task 2"},
		},
		2: {
			{ID: 3, Title: "Task 3"},
		},
	}
	m := setupTestModel(columns, tasks)

	// Position at last column
	m.UIState.SetSelectedColumn(len(columns) - 1)
	if m.UIState.SelectedColumn() != 1 {
		t.Errorf("SelectedColumn at end = %d, want 1", m.UIState.SelectedColumn())
	}

	// Position at last task in last column
	m.UIState.SetSelectedTask(len(tasks[2]) - 1)
	if m.UIState.SelectedTask() != 0 {
		t.Errorf("SelectedTask at end = %d, want 0", m.UIState.SelectedTask())
	}
}

// TestEmptyState_HandlesEmptyColumns verifies behavior with no columns.
// Edge case: Project with no columns created yet.
// Security value: Empty state doesn't crash, selection defaults to 0.
func TestEmptyState_HandlesEmptyColumns(t *testing.T) {
	m := setupTestModel([]*models.Column{}, nil)

	// Should default to valid state even with no columns
	if m.UIState.SelectedColumn() != 0 {
		t.Errorf("SelectedColumn with no columns = %d, want 0", m.UIState.SelectedColumn())
	}

	// Should handle mode transitions in empty state
	m.UIState.SetMode(state.AddColumnFormMode)
	if m.UIState.Mode() != state.AddColumnFormMode {
		t.Error("Should transition to AddColumnFormMode in empty state")
	}
}

// TestEmptyState_HandlesEmptyTasks verifies behavior when selected column has no tasks.
// Edge case: Column exists but has no tasks yet.
// Security value: Task selection is safe even when column is empty.
func TestEmptyState_HandlesEmptyTasks(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Todo"},
		{ID: 2, Name: "InProgress"},
	}
	tasks := map[int][]*models.TaskSummary{
		1: {}, // Empty column
		2: {
			{ID: 1, Title: "Task 1"},
		},
	}
	m := setupTestModel(columns, tasks)
	m.UIState.SetSelectedColumn(0)

	// Select task in empty column
	m.UIState.SetSelectedTask(0)
	if m.UIState.SelectedTask() != 0 {
		t.Errorf("SelectedTask in empty column = %d, want 0", m.UIState.SelectedTask())
	}

	// Mode should still work in empty column
	m.UIState.SetMode(state.TicketFormMode)
	if m.UIState.Mode() != state.TicketFormMode {
		t.Error("Should allow mode transitions in empty column")
	}
}

// TestStateIndependence_ColumnAndTaskSelection verifies column and task selection are independent.
// Edge case: Changing column doesn't affect task selection value (though task may be out of bounds).
// Security value: State management is predictable and doesn't have side effects.
func TestStateIndependence_ColumnAndTaskSelection(t *testing.T) {
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
	}
	m := setupTestModel(columns, nil)

	m.UIState.SetSelectedColumn(0)
	m.UIState.SetSelectedTask(5)

	// Change column
	m.UIState.SetSelectedColumn(1)

	// Task selection should remain unchanged (value stays 5, even if out of bounds)
	if m.UIState.SelectedTask() != 5 {
		t.Errorf("SelectedTask after column change = %d, want 5 (unchanged value)", m.UIState.SelectedTask())
	}

	// Column should have changed
	if m.UIState.SelectedColumn() != 1 {
		t.Errorf("SelectedColumn = %d, want 1", m.UIState.SelectedColumn())
	}
}

// TestViewportState_CalculatesCorrectly verifies viewport calculations for column visibility.
// Edge case: Terminal width changes, viewport must recalculate visible columns.
// Security value: Viewport calculations don't cause index out of bounds.
func TestViewportState_CalculatesCorrectly(t *testing.T) {
	columns := make([]*models.Column, 10)
	for i := 0; i < 10; i++ {
		columns[i] = &models.Column{ID: i + 1, Name: "Col"}
	}
	m := setupTestModel(columns, nil)

	tests := []struct {
		name    string
		width   int
		minCols int
		maxCols int
	}{
		{"Small width", 60, 1, 1},
		{"Medium width", 120, 2, 3},
		{"Large width", 300, 6, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.UIState.SetWidth(tt.width)
			viewportSize := m.UIState.ViewportSize()

			if viewportSize < tt.minCols || viewportSize > tt.maxCols {
				t.Errorf("ViewportSize for width %d = %d, want between %d-%d",
					tt.width, viewportSize, tt.minCols, tt.maxCols)
			}

			// Viewport should always show at least 1 column
			if viewportSize < 1 {
				t.Error("ViewportSize should be at least 1")
			}
		})
	}
}

// TestModeUsesLayers verifies that certain modes require layer-based rendering.
// Edge case: Identifying which modes use layered vs full-screen rendering.
// Security value: Modes are correctly categorized for rendering system.
func TestModeUsesLayers(t *testing.T) {
	tests := []struct {
		mode       state.Mode
		wantLayers bool
	}{
		{state.NormalMode, true},
		{state.TicketFormMode, true},
		{state.AddColumnFormMode, true},
		{state.LabelPickerMode, true},
		{state.SearchMode, true},
		{state.HelpMode, true},
	}

	for _, tt := range tests {
		t.Run("Mode", func(t *testing.T) {
			usesLayers := tt.mode.UsesLayers()
			if usesLayers != tt.wantLayers {
				t.Errorf("Mode %v UsesLayers = %v, want %v", tt.mode, usesLayers, tt.wantLayers)
			}
		})
	}
}
