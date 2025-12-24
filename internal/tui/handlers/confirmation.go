package handlers

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// CONFIRMATION HANDLERS
// ============================================================================

// HandleDeleteConfirm handles task deletion confirmation.
func (w *Wrapper) HandleDeleteConfirm(msg tea.KeyMsg) (*Wrapper, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return w.confirmDeleteTask()
	case "n", "N", "esc":
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	}
	return w, nil
}

// confirmDeleteTask performs the actual task deletion.
func (w *Wrapper) confirmDeleteTask() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	task := ops.GetCurrentTask()
	if task != nil {
		ctx, cancel := w.DbContext()
		defer cancel()
		err := w.Repo.DeleteTask(ctx, task.ID)
		if err != nil {
			slog.Error("Error deleting task", "error", err)
			w.NotificationState.Add(state.LevelError, "Failed to delete task")
		} else {
			ops.RemoveCurrentTask()
		}
	}
	w.UiState.SetMode(state.NormalMode)
	return w, nil
}

// HandleDiscardConfirm handles discard confirmation for forms and inputs.
// This provides a generic Y/N/ESC handler that works for all discard scenarios.
func (w *Wrapper) HandleDiscardConfirm(msg tea.KeyMsg) (*Wrapper, tea.Cmd) {
	ctx := w.UiState.DiscardContext()
	if ctx == nil {
		// Safety: if context is missing, return to normal mode
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed discard - clear form/input and return to normal mode
		return w.confirmDiscard()

	case "n", "N", "esc":
		// User cancelled - return to source mode without clearing
		w.UiState.SetMode(ctx.SourceMode)
		w.UiState.ClearDiscardContext()
		return w, nil
	}

	return w, nil
}

// confirmDiscard performs the actual discard operation based on context.
func (w *Wrapper) confirmDiscard() (*Wrapper, tea.Cmd) {
	ctx := w.UiState.DiscardContext()
	if ctx == nil {
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	}

	// Clear the appropriate form/input based on source mode
	switch ctx.SourceMode {
	case state.TicketFormMode:
		w.FormState.ClearTicketForm()

	case state.ProjectFormMode:
		w.FormState.ClearProjectForm()

	case state.AddColumnMode, state.EditColumnMode:
		w.InputState.Clear()
	}

	// Always return to normal mode after discard
	w.UiState.SetMode(state.NormalMode)
	w.UiState.ClearDiscardContext()

	return w, tea.ClearScreen
}

// HandleDeleteColumnConfirm handles column deletion confirmation.
func (w *Wrapper) HandleDeleteColumnConfirm(msg tea.KeyMsg) (*Wrapper, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return w.confirmDeleteColumn()
	case "n", "N", "esc":
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	}
	return w, nil
}

// confirmDeleteColumn performs the actual column deletion.
func (w *Wrapper) confirmDeleteColumn() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	column := ops.GetCurrentColumn()
	if column != nil {
		ctx, cancel := w.DbContext()
		defer cancel()
		err := w.Repo.DeleteColumn(ctx, column.ID)
		if err != nil {
			slog.Error("Error deleting column", "error", err)
			w.NotificationState.Add(state.LevelError, "Failed to delete column")
		} else {
			delete(w.AppState.Tasks(), column.ID)
			ops.RemoveCurrentColumn()
		}
	}
	w.UiState.SetMode(state.NormalMode)
	return w, nil
}
