package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// CONFIRMATION HANDLERS (Inlined from deleted confirmation.go)
// ============================================================================

// handleDeleteConfirm handles task deletion confirmation.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteTask()
	case "n", "N", "esc":
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteTask performs the actual task deletion.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDeleteTask() (tea.Model, tea.Cmd) {
	task := m.getCurrentTask()
	if task != nil {
		ctx, cancel := m.DbContext()
		defer cancel()
		err := m.App.Repo().DeleteTask(ctx, task.ID)
		if err != nil {
			slog.Error("Error deleting task", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to delete task")
		} else {
			m.removeCurrentTask()
		}
	}
	m.UiState.SetMode(state.NormalMode)
	return m, nil
}

// handleDiscardConfirm handles discard confirmation for forms and inputs.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDiscardConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ctx := m.UiState.DiscardContext()
	if ctx == nil {
		// Safety: if context is missing, return to normal mode
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed discard - clear form/input and return to normal mode
		return m.confirmDiscard()

	case "n", "N", "esc":
		// User cancelled - return to source mode without clearing
		m.UiState.SetMode(ctx.SourceMode)
		m.UiState.ClearDiscardContext()
		return m, nil
	}

	return m, nil
}

// confirmDiscard performs the actual discard operation based on context.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDiscard() (tea.Model, tea.Cmd) {
	ctx := m.UiState.DiscardContext()
	if ctx == nil {
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Clear the appropriate form/input based on source mode
	switch ctx.SourceMode {
	case state.TicketFormMode:
		m.FormState.ClearTicketForm()

	case state.ProjectFormMode:
		m.FormState.ClearProjectForm()

	case state.AddColumnMode, state.EditColumnMode:
		m.InputState.Clear()
	}

	// Always return to normal mode after discard
	m.UiState.SetMode(state.NormalMode)
	m.UiState.ClearDiscardContext()

	return m, tea.ClearScreen
}

// handleDeleteColumnConfirm handles column deletion confirmation.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) handleDeleteColumnConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m.confirmDeleteColumn()
	case "n", "N", "esc":
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}
	return m, nil
}

// confirmDeleteColumn performs the actual column deletion.
// Inlined from confirmation.go (deleted to reduce duplication)
func (m Model) confirmDeleteColumn() (tea.Model, tea.Cmd) {
	column := m.getCurrentColumn()
	if column != nil {
		ctx, cancel := m.DbContext()
		defer cancel()
		err := m.App.Repo().DeleteColumn(ctx, column.ID)
		if err != nil {
			slog.Error("Error deleting column", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to delete column")
		} else {
			delete(m.AppState.Tasks(), column.ID)
			m.removeCurrentColumn()
		}
	}
	m.UiState.SetMode(state.NormalMode)
	return m, nil
}
