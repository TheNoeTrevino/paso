package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================

// handleEnterSearch enters search mode and clears any previous search state.
// Inlined from search.go (deleted to reduce duplication)
func (m Model) handleEnterSearch() (tea.Model, tea.Cmd) {
	m.UI.Search.Clear()
	m.UI.Search.Deactivate()
	m.UIState.SetMode(state.SearchMode)
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
	m.UIState.SetMode(state.NormalMode)
	return m, nil
}

// handleSearchCancel clears the search and returns to normal mode.
// Inlined from search.go (deleted to reduce duplication)
func (m Model) handleSearchCancel() (tea.Model, tea.Cmd) {
	m.UI.Search.Clear()
	m.UI.Search.Deactivate()
	m.UIState.SetMode(state.NormalMode)
	return m.executeSearch()
}

// executeSearch runs the search query and updates the task list.
// Inlined from search.go (deleted to reduce duplication)
func (m Model) executeSearch() (tea.Model, tea.Cmd) {
	project := m.getCurrentProject()
	if project == nil {
		return m, nil
	}

	ctx, cancel := m.DBContext()
	defer cancel()
	var tasksByColumn map[int][]*models.TaskSummary
	var err error

	if m.UI.Search.Query == "" {
		tasksByColumn, err = m.App.TaskService.GetTaskSummariesByProject(ctx, project.ID)
	} else {
		tasksByColumn, err = m.App.TaskService.GetTaskSummariesByProjectFiltered(ctx, project.ID, m.UI.Search.Query)
	}

	if err != nil {
		slog.Error("failed to filtering tasks", "error", err)
		return m, nil
	}

	m.AppState.SetTasks(tasksByColumn)
	// Reset task selection to 0 to avoid out-of-bounds
	m.UIState.SetSelectedTask(0)

	return m, nil
}
