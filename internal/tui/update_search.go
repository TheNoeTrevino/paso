package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// handleEnterSearch enters search mode and clears any previous search state.
// Inlined from search.go (deleted to reduce duplication)
func (m Model) handleEnterSearch() (tea.Model, tea.Cmd) {
	m.UI.Search.Clear()
	m.UI.Search.Deactivate()
	m.UIState.Mode = state.SearchMode
	return m, nil
}

// handleSearchMode handles keyboard input in search mode.
// Inlined from search.go (deleted to reduce duplication)
func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleSearchConfirm()
	case "esc":
		return m.handleSearchCancel()
	case "backspace", "ctrl+h":
		if m.UI.Search.Backspace() {
			return m.executeSearch()
		}
		return m, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			if m.UI.Search.AppendChar(rune(key[0])) {
				return m.executeSearch()
			}
		}
		return m, nil
	}
}

// handleSearchConfirm activates the filter and returns to normal mode.
// Inlined from search.go (deleted to reduce duplication)
func (m Model) handleSearchConfirm() (tea.Model, tea.Cmd) {
	m.UI.Search.Activate()
	m.UIState.Mode = state.NormalMode
	return m, nil
}

// handleSearchCancel clears the search and returns to normal mode.
// Inlined from search.go (deleted to reduce duplication)
func (m Model) handleSearchCancel() (tea.Model, tea.Cmd) {
	m.UI.Search.Clear()
	m.UI.Search.Deactivate()
	m.UIState.Mode = state.NormalMode
	return m.executeSearch()
}

// executeSearch runs the search query and updates the task list.
// Delegates to fetchTasksForCurrentProject which handles both search and field filters.
func (m Model) executeSearch() (tea.Model, tea.Cmd) {
	project := m.getCurrentProject()
	if project == nil {
		return m, nil
	}

	ctx, cancel := m.DBContext()
	defer cancel()

	tasksByColumn, err := m.fetchTasksForCurrentProject(ctx, project.ID)
	if err != nil {
		slog.Error("failed to filter tasks", "error", err)
		return m, nil
	}

	m.AppState.SetTasks(tasksByColumn)
	m.UIState.SelectedTask = 0

	return m, nil
}
