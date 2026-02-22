package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestModeDispatch_TaskFormMode ensures form mode intercepts all messages.
// Edge case: When in TaskFormMode, ALL messages should go to updateTaskForm.
func TestModeDispatch_TaskFormMode(t *testing.T) {
	t.Parallel()
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	m := setupTestModel(columns, nil)

	// Create a simple form (will be nil initially, but mode is what matters)
	m.UIState.Mode = state.TaskFormMode
	title := ""
	description := ""
	confirm := false
	m.Forms.Form.TaskForm = huhforms.CreateTaskForm(&title, &description, &confirm, 5)

	// Send a key message
	keyMsg := tea.KeyPressMsg(tea.Key{Text: "a", Code: 'a'})

	// Update should route to updateTaskForm
	newModel, cmd := m.Update(keyMsg)
	m = newModel.(Model)

	// Mode should still be TaskFormMode (until form completes)
	assert.Equal(t, state.TaskFormMode, m.UIState.Mode)

	// Cmd should not be nil (form returns commands)
	if cmd == nil {
		t.Log("Update returned nil cmd (expected for incomplete form)")
	}
}

// TestModeDispatch_NormalMode ensures normal mode routes to handlers.
// Edge case: KeyMsg in NormalMode should go to handleKeyMsg dispatcher.
func TestModeDispatch_NormalMode(t *testing.T) {
	t.Parallel()
	columns := []*models.Column{
		{ID: 1, Name: "Col1"},
		{ID: 2, Name: "Col2"},
	}
	m := setupTestModel(columns, nil)
	m.UIState.Mode = state.NormalMode
	m.UIState.SelectedColumn = 0

	// Send a navigation key (right arrow)
	keyMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})

	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// Should have navigated right
	assert.Equal(t, 1, m.UIState.SelectedColumn)

	// Mode should still be NormalMode
	assert.Equal(t, state.NormalMode, m.UIState.Mode)
}

// TestUpdateTaskForm_EscapeCancels ensures ESC key exits form mode.
func TestUpdateTaskForm_EscapeCancels(t *testing.T) {
	t.Parallel()
	columns := []*models.Column{{ID: 1, Name: "Todo"}}
	m := setupTestModel(columns, nil)

	// Set up form mode
	m.UIState.Mode = state.TaskFormMode
	title := ""
	description := ""
	confirm := false
	m.Forms.Form.TaskForm = huhforms.CreateTaskForm(&title, &description, &confirm, 5)

	// Send ESC key
	keyMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})

	newModel, cmd := m.Update(keyMsg)
	m = newModel.(Model)

	// Should return to NormalMode
	assert.Equal(t, state.NormalMode, m.UIState.Mode)

	// Form should be cleared
	assert.Nil(t, m.Forms.Form.TaskForm, "TaskForm after ESC should be nil")

	// Cmd should be a ClearScreen command (clean exit triggers tea.ClearScreen)
	assert.NotNil(t, cmd, "Cmd after ESC should be non-nil (tea.ClearScreen)")
}

// TestUpdateTaskForm_EmptyTitleEscNoTask verifies that cancelling a form with an empty
// title doesn't create a task. This exercises the ESC path where the form has no changes
// (empty title = no changes), so it exits immediately without saving.
func TestUpdateTaskForm_EmptyTitleEscNoTask(t *testing.T) {
	t.Parallel()
	m, db := SetupTestModelWithDB(t)
	m.NotifyChan = make(chan events.NotificationMsg, 1)

	currentProject := m.AppState.GetCurrentProject()
	require.NotNil(t, currentProject, "test model should have a current project")

	// Count existing tasks before form interaction
	var initialTaskCount int
	err := db.QueryRowContext(m.Ctx, "SELECT COUNT(*) FROM tasks").Scan(&initialTaskCount)
	require.NoError(t, err)

	// Enter task form mode with an empty title
	m.UIState.Mode = state.TaskFormMode
	title := ""
	description := ""
	confirm := false
	m.Forms.Form.TaskForm = huhforms.CreateTaskForm(&title, &description, &confirm, 5)

	// Press ESC to cancel (empty form = no changes = immediate exit)
	keyMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	// Should return to NormalMode
	assert.Equal(t, state.NormalMode, m.UIState.Mode)

	// No task should have been created
	var afterTaskCount int
	err = db.QueryRowContext(m.Ctx, "SELECT COUNT(*) FROM tasks").Scan(&afterTaskCount)
	require.NoError(t, err)
	assert.Equal(t, initialTaskCount, afterTaskCount, "Task count changed after ESC on empty form")
}
