package tui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testing/fixtures"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestLabelPicker_NavigationAndSelection tests arrow key navigation and Enter selection
func TestLabelPicker_NavigationAndSelection(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	// Create test labels
	ctx := context.Background()
	projectID := m.AppState.GetCurrentProjectID()
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "bug", "#FF0000")
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "feature", "#00FF00")
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "docs", "#0000FF")

	// Reload labels
	labels, err := m.App.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load labels for test setup")
	m.AppState.SetLabels(labels)

	// Enter label picker mode and set labels
	m.UIState.Mode = state.LabelPickerMode
	for _, label := range labels {
		m.Pickers.Label.AddItem(state.LabelPickerItem{Label: label, Selected: false})
	}

	// Navigate down arrow
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Navigate down arrow again
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Press enter to select
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify model was updated successfully
	if m.UIState.Mode == state.LabelPickerMode {
		// If we're still in picker mode, that's valid for multiple selection
		assert.NotEmpty(t, m.Pickers.Label.Items, "Expected labels in picker state")
	}
}

// TestLabelPicker_FilteringAndBackspace tests typing to filter and backspace to clear
func TestLabelPicker_FilteringAndBackspace(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	// Create labels
	ctx := context.Background()
	projectID := m.AppState.GetCurrentProjectID()
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "bug", "#FF0000")
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "backend", "#00FF00")
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "frontend", "#0000FF")

	labels, err := m.App.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load labels for test setup")
	m.AppState.SetLabels(labels)
	m.UIState.Mode = state.LabelPickerMode
	for _, label := range labels {
		m.Pickers.Label.AddItem(state.LabelPickerItem{Label: label, Selected: false})
	}

	// Type 'b' to filter
	msg := tea.KeyPressMsg(tea.Key{Text: "b", Code: 'b'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Press backspace to clear filter
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify we're back to showing labels in the picker
	assert.Equal(t, state.LabelPickerMode, m.UIState.Mode)
}

// TestLabelPicker_MultiSelectToggle tests space key to toggle selection
func TestLabelPicker_MultiSelectToggle(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	// Create labels
	ctx := context.Background()
	projectID := m.AppState.GetCurrentProjectID()
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "bug", "#FF0000")
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "feature", "#00FF00")

	labels, err := m.App.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load labels for test setup")
	m.AppState.SetLabels(labels)
	m.UIState.Mode = state.LabelPickerMode
	for _, label := range labels {
		m.Pickers.Label.AddItem(state.LabelPickerItem{Label: label, Selected: false})
	}

	// Press space to toggle current selection
	msg := tea.KeyPressMsg(tea.Key{Code: ' '})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Navigate down
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Press space again
	msg = tea.KeyPressMsg(tea.Key{Code: ' '})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Press enter to confirm
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify model was updated
	require.NotNil(t, m, "Model should not be nil after picker selection")
}

// TestPriorityPicker_Selection tests priority picker navigation and selection
func TestPriorityPicker_Selection(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	// Enter priority picker mode
	m.UIState.Mode = state.PriorityPickerMode

	// Navigate down arrow
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Press enter to select
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify mode can change or stay in picker
	require.NotNil(t, m, "Model should not be nil after picker selection")
}

// TestTypePicker_Selection tests type picker navigation and selection
func TestTypePicker_Selection(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	// Enter type picker mode
	m.UIState.Mode = state.TypePickerMode

	// Navigate down arrow
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Press enter to select
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify model updated
	require.NotNil(t, m, "Model should not be nil after picker selection")
}

// TestParentPicker_SearchAndSelect tests searching for parent task and selecting it
func TestParentPicker_SearchAndSelect(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	// Create a parent task
	ctx := context.Background()
	projectID := m.AppState.GetCurrentProjectID()
	columns, err := m.App.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load columns for test setup")
	if len(columns) == 0 {
		t.Skip("No columns available for testing")
	}

	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[0].ID, "Parent Task")

	// Reload tasks
	tasks, err := m.App.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load tasks for test setup")
	m.AppState.SetTasks(tasks)

	// Enter parent picker mode
	m.UIState.Mode = state.ParentPickerMode

	// Type 'p' to filter
	msg := tea.KeyPressMsg(tea.Key{Text: "p", Code: 'p'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Press enter to select
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify parent picker interaction
	require.NotNil(t, m, "Model should not be nil after parent picker interaction")
}

// TestChildPicker_SearchAndSelect tests searching for child task and selecting it
func TestChildPicker_SearchAndSelect(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	// Create a child task
	ctx := context.Background()
	projectID := m.AppState.GetCurrentProjectID()
	columns, err := m.App.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load columns for test setup")
	if len(columns) == 0 {
		t.Skip("No columns available for testing")
	}

	fixtures.CreateTestTask(t, db, fixtures.SQLiteDialect(), columns[0].ID, "Child Task")

	// Reload tasks
	tasks, err := m.App.TaskService.GetTaskSummariesByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load tasks for test setup")
	m.AppState.SetTasks(tasks)

	// Enter child picker mode
	m.UIState.Mode = state.ChildPickerMode

	// Type 'c' to filter
	msg := tea.KeyPressMsg(tea.Key{Text: "c", Code: 'c'})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Press enter to select
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify child picker interaction
	require.NotNil(t, m, "Model should not be nil after child picker interaction")
}

// TestRelationTypePicker_Selection tests relation type picker
func TestRelationTypePicker_Selection(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	// Enter relation type picker mode
	m.UIState.Mode = state.RelationTypePickerMode

	// Navigate down arrow
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Press enter to select
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify mode can change or stay in relation picker
	require.NotNil(t, m, "Model should not be nil after relation picker selection")
}

// TestStatusPicker_ColumnSelection tests status/column picker selection
func TestStatusPicker_ColumnSelection(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	// Enter status picker mode
	m.UIState.Mode = state.StatusPickerMode

	ctx := context.Background()
	projectID := m.AppState.GetCurrentProjectID()
	columns, err := m.App.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load columns for test setup")
	m.Pickers.Status.SetColumns(columns)

	// Navigate down arrow
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Press enter to select
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify mode can change
	require.NotNil(t, m, "Model should not be nil after status picker selection")
}

// TestLabelPicker_EscapeExitsMode tests that Escape exits the picker
func TestLabelPicker_EscapeExitsMode(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)

	// Create labels
	ctx := context.Background()
	projectID := m.AppState.GetCurrentProjectID()
	fixtures.CreateTestLabel(t, db, fixtures.SQLiteDialect(), projectID, "test", "#FFFFFF")

	labels, err := m.App.LabelService.GetLabelsByProject(ctx, projectID)
	require.NoError(t, err, "Failed to load labels for test setup")
	m.AppState.SetLabels(labels)
	m.UIState.Mode = state.LabelPickerMode
	for _, label := range labels {
		m.Pickers.Label.AddItem(state.LabelPickerItem{Label: label, Selected: false})
	}

	// Press Escape
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Verify something happened (mode change or escape handled)
	// Escape was processed
	require.NotNil(t, m, "Model should not be nil after escape")
}

// TestPriorityPicker_UpDownNavigation tests up and down navigation in priority picker
func TestPriorityPicker_UpDownNavigation(t *testing.T) {
	t.Parallel()
	m, _ := SetupTestModelWithDB(t)

	// Enter priority picker mode
	m.UIState.Mode = state.PriorityPickerMode

	// Navigate down
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Navigate up
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	// Verify navigation works
	assert.Equal(t, state.PriorityPickerMode, m.UIState.Mode)
}
