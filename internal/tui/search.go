package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// SEARCH MODE HANDLERS
// ============================================================================

// handleEnterSearch enters search mode and clears any previous search state.
func (m Model) handleEnterSearch() (tea.Model, tea.Cmd) {
	m.searchState.Clear()
	m.searchState.Deactivate()
	m.uiState.SetMode(state.SearchMode)
	return m, nil
}

// handleSearchMode handles keyboard input in search mode.
func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleSearchConfirm()
	case "esc":
		return m.handleSearchCancel()
	case "backspace", "ctrl+h":
		if m.searchState.Backspace() {
			return m.executeSearch()
		}
		return m, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			if m.searchState.AppendChar(rune(key[0])) {
				return m.executeSearch()
			}
		}
		return m, nil
	}
}

// handleSearchConfirm activates the filter and returns to normal mode.
// The search query persists and continues to filter the kanban view.
func (m Model) handleSearchConfirm() (tea.Model, tea.Cmd) {
	m.searchState.Activate()
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}

// handleSearchCancel clears the search and returns to normal mode.
// All tasks are shown again.
func (m Model) handleSearchCancel() (tea.Model, tea.Cmd) {
	m.searchState.Clear()
	m.searchState.Deactivate()
	m.uiState.SetMode(state.NormalMode)
	return m.executeSearch() // Reload all tasks
}

// executeSearch runs the search query and updates the task list.
// If the query is empty, all tasks are loaded. Otherwise, only matching tasks are loaded.
func (m Model) executeSearch() (tea.Model, tea.Cmd) {
	project := m.getCurrentProject()
	if project == nil {
		return m, nil
	}

	ctx, cancel := m.dbContext()
	defer cancel()
	var tasksByColumn map[int][]*models.TaskSummary
	var err error

	if m.searchState.Query == "" {
		tasksByColumn, err = m.repo.GetTaskSummariesByProject(ctx, project.ID)
	} else {
		tasksByColumn, err = m.repo.GetTaskSummariesByProjectFiltered(ctx, project.ID, m.searchState.Query)
	}

	if err != nil {
		slog.Error("Error filtering tasks", "error", err)
		return m, nil
	}

	m.appState.SetTasks(tasksByColumn)
	// Reset task selection to 0 to avoid out-of-bounds
	m.uiState.SetSelectedTask(0)

	return m, nil
}
