package handlers

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// SEARCH MODE HANDLERS
// ============================================================================

// HandleEnterSearch enters search mode and clears any previous search state.
func (w *Wrapper) HandleEnterSearch() (*Wrapper, tea.Cmd) {
	w.SearchState.Clear()
	w.SearchState.Deactivate()
	w.UiState.SetMode(state.SearchMode)
	return w, nil
}

// HandleSearchMode handles keyboard input in search mode.
func (w *Wrapper) HandleSearchMode(msg tea.KeyMsg) (*Wrapper, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return w.handleSearchConfirm()
	case "esc":
		return w.handleSearchCancel()
	case "backspace", "ctrl+h":
		if w.SearchState.Backspace() {
			return w.executeSearch()
		}
		return w, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			if w.SearchState.AppendChar(rune(key[0])) {
				return w.executeSearch()
			}
		}
		return w, nil
	}
}

// handleSearchConfirm activates the filter and returns to normal mode.
// The search query persists and continues to filter the kanban view.
func (w *Wrapper) handleSearchConfirm() (*Wrapper, tea.Cmd) {
	w.SearchState.Activate()
	w.UiState.SetMode(state.NormalMode)
	return w, nil
}

// handleSearchCancel clears the search and returns to normal mode.
// All tasks are shown again.
func (w *Wrapper) handleSearchCancel() (*Wrapper, tea.Cmd) {
	w.SearchState.Clear()
	w.SearchState.Deactivate()
	w.UiState.SetMode(state.NormalMode)
	return w.executeSearch() // Reload all tasks
}

// executeSearch runs the search query and updates the task list.
// If the query is empty, all tasks are loaded. Otherwise, only matching tasks are loaded.
func (w *Wrapper) executeSearch() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	project := ops.GetCurrentProject()
	if project == nil {
		return w, nil
	}

	ctx, cancel := w.DbContext()
	defer cancel()
	var tasksByColumn map[int][]*models.TaskSummary
	var err error

	if w.SearchState.Query == "" {
		tasksByColumn, err = w.Repo.GetTaskSummariesByProject(ctx, project.ID)
	} else {
		tasksByColumn, err = w.Repo.GetTaskSummariesByProjectFiltered(ctx, project.ID, w.SearchState.Query)
	}

	if err != nil {
		slog.Error("Error filtering tasks", "error", err)
		return w, nil
	}

	w.AppState.SetTasks(tasksByColumn)
	// Reset task selection to 0 to avoid out-of-bounds
	w.UiState.SetSelectedTask(0)

	return w, nil
}
