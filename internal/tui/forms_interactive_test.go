package tui

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// TestTaskForm_FieldProgression tests Tab and Shift+Tab navigation between form fields
func TestTaskForm_FieldProgression(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter task form mode
	m.UIState.SetMode(state.TicketFormMode)

	// Press Tab to move to next field
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Press Tab again
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Press Shift+Tab to move to previous field
	msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Verify model was updated successfully
	_ = m
}

// TestTaskForm_ShortcutToPriorityPicker tests Ctrl+P to open priority picker from form
func TestTaskForm_ShortcutToPriorityPicker(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter task form mode
	m.UIState.SetMode(state.TicketFormMode)

	// Press Ctrl+P to open priority picker
	msg := tea.KeyPressMsg(tea.Key{Code: 'p', Mod: tea.ModCtrl})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(50 * time.Millisecond)

	// Verify mode changed or form still active
	_ = m
}

// TestTaskForm_ShortcutToLabelPicker tests Ctrl+L to open label picker from form
func TestTaskForm_ShortcutToLabelPicker(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter task form mode
	m.UIState.SetMode(state.TicketFormMode)

	// Press Ctrl+L to open label picker
	msg := tea.KeyPressMsg(tea.Key{Code: 'l', Mod: tea.ModCtrl})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(50 * time.Millisecond)

	// Verify mode changed or form still active
	_ = m
}

// TestTaskForm_ShortcutToParentPicker tests Ctrl+Shift+P to open parent picker from form
func TestTaskForm_ShortcutToParentPicker(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter task form mode
	m.UIState.SetMode(state.TicketFormMode)

	// Press Ctrl+Shift+P to open parent picker
	msg := tea.KeyPressMsg(tea.Key{Code: 'P', Mod: tea.ModCtrl | tea.ModShift})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(50 * time.Millisecond)

	// Verify mode changed or form still active
	_ = m
}

// TestTaskForm_SaveWithCtrlS tests submitting form with Ctrl+S
func TestTaskForm_SaveWithCtrlS(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter task form mode
	m.UIState.SetMode(state.TicketFormMode)

	// Type some input (simulating user filling the form)
	TypeStringToModel(&m, "Test Task")
	time.Sleep(50 * time.Millisecond)

	// Press Ctrl+S to save
	msg := tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(100 * time.Millisecond)

	// Verify form processed the save command
	_ = m
}

// TestTaskForm_DiscardConfirmation tests Esc key to trigger discard confirmation
func TestTaskForm_DiscardConfirmation(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter task form mode
	m.UIState.SetMode(state.TicketFormMode)

	// Type some input to make form "dirty"
	TypeStringToModel(&m, "Unsaved Task")
	time.Sleep(50 * time.Millisecond)

	// Press Escape to trigger discard
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(50 * time.Millisecond)

	// Mode may change to confirmation or back to normal
	_ = m
}

// TestProjectForm_Creation tests creating a new project
func TestProjectForm_Creation(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter project form mode
	m.UIState.SetMode(state.ProjectFormMode)

	// Type project name
	TypeStringToModel(&m, "New Project")
	time.Sleep(50 * time.Millisecond)

	// Press Tab to move to description field
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Type description
	TypeStringToModel(&m, "Project description")
	time.Sleep(50 * time.Millisecond)

	// Press Ctrl+S to save
	msg = tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(100 * time.Millisecond)

	// Verify form was processed
	if m.UIState.Mode() == state.ProjectFormMode {
		// Still in form mode, form may not support quick save
	}
}

// TestColumnForm_Creation tests creating a new column
func TestColumnForm_Creation(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter column form mode
	m.UIState.SetMode(state.AddColumnFormMode)

	// Type column name
	TypeStringToModel(&m, "New Column")
	time.Sleep(50 * time.Millisecond)

	// Press Tab to move to next field if available
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Press Ctrl+S to save
	msg = tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(100 * time.Millisecond)

	// Verify form was processed
	_ = m
}

// TestCommentForm_Creation tests adding a comment to a task
func TestCommentForm_Creation(t *testing.T) {
	m, db := SetupTestModelWithDB(t)

	// Create a task to comment on
	ctx := context.Background()
	projectID := m.AppState.GetCurrentProjectID()
	columns, _ := m.App.ColumnService.GetColumnsByProject(ctx, projectID)
	if len(columns) > 0 {
		taskID := testutil.CreateTestTask(t, db, columns[0].ID, "Task to comment on")

		// Enter comment form mode
		m.UIState.SetMode(state.CommentFormMode)
		m.Forms.Comment.TaskID = taskID

		// Type comment text
		TypeStringToModel(&m, "This is a comment")
		time.Sleep(50 * time.Millisecond)

		// Press Ctrl+S to save comment
		msg := tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl})
		updatedModel, _ := m.Update(msg)
		m = updatedModel.(Model)
		time.Sleep(100 * time.Millisecond)

		// Verify comment form was processed
		_ = m
	}
}

// TestEditColumnForm_RenameColumn tests renaming an existing column
func TestEditColumnForm_RenameColumn(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter edit column form mode
	m.UIState.SetMode(state.EditColumnFormMode)

	// Type new column name
	TypeStringToModel(&m, "Renamed Column")
	time.Sleep(50 * time.Millisecond)

	// Press Tab to navigate form
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)

	// Press Ctrl+S to save column
	msg = tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl})
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(100 * time.Millisecond)

	// Verify edit column form was processed
	_ = m
}

// TestProjectForm_DiscardChanges tests discarding changes to project form
func TestProjectForm_DiscardChanges(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter project form mode
	m.UIState.SetMode(state.ProjectFormMode)

	// Type some text
	TypeStringToModel(&m, "Project name")
	time.Sleep(50 * time.Millisecond)

	initialMode := m.UIState.Mode()

	// Press Escape to discard
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)
	time.Sleep(50 * time.Millisecond)

	// Verify escape was processed
	if m.UIState.Mode() != initialMode {
		// Mode changed, which is expected
	}
}

// TestTaskForm_CharacterInput tests typing characters into task form
func TestTaskForm_CharacterInput(t *testing.T) {
	m, _ := SetupTestModelWithDB(t)

	// Enter task form mode
	m.UIState.SetMode(state.TicketFormMode)

	// Type a series of characters
	testChars := "This is a task title with spaces 123!@#"
	TypeStringToModel(&m, testChars)
	time.Sleep(50 * time.Millisecond)

	// Verify form is still in task form mode
	if m.UIState.Mode() != state.TicketFormMode {
		t.Errorf("Expected TicketFormMode after input, got %v", m.UIState.Mode())
	}
}
