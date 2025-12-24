package handlers

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// CONFIRMATION HANDLERS
// ============================================================================

// HandleDeleteConfirm handles task deletion confirmation.
func HandleDeleteConfirm(m *tui.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		return confirmDeleteTask(m)
	case "n", "N", "esc":
		m.UiState.SetMode(state.NormalMode)
		return nil
	}
	return nil
}

// confirmDeleteTask performs the actual task deletion.
func confirmDeleteTask(m *tui.Model) tea.Cmd {
	task := modelops.GetCurrentTask(m)
	if task != nil {
		ctx, cancel := m.DbContext()
		defer cancel()
		err := m.Repo.DeleteTask(ctx, task.ID)
		if err != nil {
			slog.Error("Error deleting task", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to delete task")
		} else {
			modelops.RemoveCurrentTask(m)
		}
	}
	m.UiState.SetMode(state.NormalMode)
	return nil
}

// HandleDiscardConfirm handles discard confirmation for forms and inputs.
// This provides a generic Y/N/ESC handler that works for all discard scenarios.
func HandleDiscardConfirm(m *tui.Model, msg tea.KeyMsg) tea.Cmd {
	ctx := m.UiState.DiscardContext()
	if ctx == nil {
		// Safety: if context is missing, return to normal mode
		m.UiState.SetMode(state.NormalMode)
		return nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed discard - clear form/input and return to normal mode
		return confirmDiscard(m)

	case "n", "N", "esc":
		// User cancelled - return to source mode without clearing
		m.UiState.SetMode(ctx.SourceMode)
		m.UiState.ClearDiscardContext()
		return nil
	}

	return nil
}

// confirmDiscard performs the actual discard operation based on context.
func confirmDiscard(m *tui.Model) tea.Cmd {
	ctx := m.UiState.DiscardContext()
	if ctx == nil {
		m.UiState.SetMode(state.NormalMode)
		return nil
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

	return tea.ClearScreen
}

// HandleDeleteColumnConfirm handles column deletion confirmation.
func HandleDeleteColumnConfirm(m *tui.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		return confirmDeleteColumn(m)
	case "n", "N", "esc":
		m.UiState.SetMode(state.NormalMode)
		return nil
	}
	return nil
}

// confirmDeleteColumn performs the actual column deletion.
func confirmDeleteColumn(m *tui.Model) tea.Cmd {
	column := modelops.GetCurrentColumn(m)
	if column != nil {
		ctx, cancel := m.DbContext()
		defer cancel()
		err := m.Repo.DeleteColumn(ctx, column.ID)
		if err != nil {
			slog.Error("Error deleting column", "error", err)
			m.NotificationState.Add(state.LevelError, "Failed to delete column")
		} else {
			delete(m.AppState.Tasks(), column.ID)
			modelops.RemoveCurrentColumn(m)
		}
	}
	m.UiState.SetMode(state.NormalMode)
	return nil
}
