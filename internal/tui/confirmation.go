package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// CONFIRMATION HANDLERS
// ============================================================================

// handleDeleteConfirm handles task deletion confirmation.
func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteTask()
	case "n", "N", "esc":
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteTask performs the actual task deletion.
func (m Model) confirmDeleteTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task != nil {
		ctx, cancel := m.dbContext()
		defer cancel()
		err := m.repo.DeleteTask(ctx, task.ID)
		if err != nil {
			slog.Error("Error deleting task", "error", err)
			m.notificationState.Add(state.LevelError, "Failed to delete task")
		} else {
			m.removeCurrentTask()
		}
	}
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}

// handleDiscardConfirm handles discard confirmation for forms and inputs.
// This provides a generic Y/N/ESC handler that works for all discard scenarios.
func (m Model) handleDiscardConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ctx := m.uiState.DiscardContext()
	if ctx == nil {
		// Safety: if context is missing, return to normal mode
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed discard - clear form/input and return to normal mode
		return m.confirmDiscard()

	case "n", "N", "esc":
		// User cancelled - return to source mode without clearing
		m.uiState.SetMode(ctx.SourceMode)
		m.uiState.ClearDiscardContext()
		return m, nil
	}

	return m, nil
}

// confirmDiscard performs the actual discard operation based on context.
func (m Model) confirmDiscard() (tea.Model, tea.Cmd) {
	ctx := m.uiState.DiscardContext()
	if ctx == nil {
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Clear the appropriate form/input based on source mode
	switch ctx.SourceMode {
	case state.TicketFormMode:
		m.formState.ClearTicketForm()

	case state.ProjectFormMode:
		m.formState.ClearProjectForm()

	case state.AddColumnMode, state.EditColumnMode:
		m.inputState.Clear()
	}

	// Always return to normal mode after discard
	m.uiState.SetMode(state.NormalMode)
	m.uiState.ClearDiscardContext()

	return m, tea.ClearScreen
}

// handleDeleteColumnConfirm handles column deletion confirmation.
func (m Model) handleDeleteColumnConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteColumn()
	case "n", "N", "esc":
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteColumn performs the actual column deletion.
func (m Model) confirmDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		ctx, cancel := m.dbContext()
		defer cancel()
		err := m.repo.DeleteColumn(ctx, column.ID)
		if err != nil {
			slog.Error("Error deleting column", "error", err)
			m.notificationState.Add(state.LevelError, "Failed to delete column")
		} else {
			delete(m.appState.Tasks(), column.ID)
			m.removeCurrentColumn()
		}
	}
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}
