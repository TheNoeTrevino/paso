package handlers

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// SEARCH MODE HANDLERS
// ============================================================================

// HandleEnterSearch enters search mode and clears any previous search state.
func HandleEnterSearch(m *tui.Model) tea.Cmd {
	m.SearchState.Clear()
	m.SearchState.Deactivate()
	m.UiState.SetMode(state.SearchMode)
	return nil
}

// HandleSearchMode handles keyboard input in search mode.
func HandleSearchMode(m *tui.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		return handleSearchConfirm(m)
	case "esc":
		return handleSearchCancel(m)
	case "backspace", "ctrl+h":
		if m.SearchState.Backspace() {
			return executeSearch(m)
		}
		return nil
	default:
		key := msg.String()
		if len(key) == 1 {
			if m.SearchState.AppendChar(rune(key[0])) {
				return executeSearch(m)
			}
		}
		return nil
	}
}

// handleSearchConfirm activates the filter and returns to normal mode.
// The search query persists and continues to filter the kanban view.
func handleSearchConfirm(m *tui.Model) tea.Cmd {
	m.SearchState.Activate()
	m.UiState.SetMode(state.NormalMode)
	return nil
}

// handleSearchCancel clears the search and returns to normal mode.
// All tasks are shown again.
func handleSearchCancel(m *tui.Model) tea.Cmd {
	m.SearchState.Clear()
	m.SearchState.Deactivate()
	m.UiState.SetMode(state.NormalMode)
	return executeSearch(m) // Reload all tasks
}

// executeSearch runs the search query and updates the task list.
// If the query is empty, all tasks are loaded. Otherwise, only matching tasks are loaded.
func executeSearch(m *tui.Model) tea.Cmd {
	project := modelops.GetCurrentProject(m)
	if project == nil {
		return nil
	}

	ctx, cancel := m.DbContext()
	defer cancel()
	var tasksByColumn map[int][]*models.TaskSummary
	var err error

	if m.SearchState.Query == "" {
		tasksByColumn, err = m.Repo.GetTaskSummariesByProject(ctx, project.ID)
	} else {
		tasksByColumn, err = m.Repo.GetTaskSummariesByProjectFiltered(ctx, project.ID, m.SearchState.Query)
	}

	if err != nil {
		slog.Error("Error filtering tasks", "error", err)
		return nil
	}

	m.AppState.SetTasks(tasksByColumn)
	// Reset task selection to 0 to avoid out-of-bounds
	m.UiState.SetSelectedTask(0)

	return nil
}
