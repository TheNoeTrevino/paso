package tui

import (
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

func (m Model) handleFilterBarMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.UI.Filter.SearchInputActive {
		return m.handleSearchInput(msg)
	}

	switch msg.String() {
	case "esc", "q":
		m.UI.Filter.IsActive = false
		m.UIState.Mode = state.NormalMode
		return m, nil

	case "enter":
		if m.UI.Filter.FocusedChip == state.FilterChipSearch {
			m.UI.Filter.SearchInputActive = true
			return m, nil
		}
		if m.UI.Filter.FocusedChip == state.FilterChipClearAll {
			m.UI.Filter.ClearAll()
			return m.refreshFilteredTasks()
		}
		if m.UI.Filter.FocusedChip == state.FilterChipArchived {
			m.UI.Filter.ShowArchived = !m.UI.Filter.ShowArchived
			return m.refreshFilteredTasks()
		}
		return m.openFilterPicker()

	case "h", "left":
		m.UI.Filter.MoveFocusLeft()
		return m, nil

	case "l", "right":
		m.UI.Filter.MoveFocusRight()
		return m, nil

	case "x", "delete", "backspace":
		if m.UI.Filter.ClearFocusedChip() {
			return m.refreshFilteredTasks()
		}
		return m, nil
	}

	return m, nil
}

// handleSearchInput handles keyboard input when the search chip is in text input mode.
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.UI.Filter.SearchInputActive = false
		return m.refreshFilteredTasks()
	case "esc":
		m.UI.Filter.SearchInputActive = false
		return m, nil
	case "backspace", "ctrl+h":
		if m.UI.Filter.SearchBackspace() {
			return m.refreshFilteredTasks()
		}
		return m, nil
	case "space":
		if m.UI.Filter.AppendSearchChar(' ') {
			return m.refreshFilteredTasks()
		}
		return m, nil
	default:
		key := msg.String()
		if len(key) == 1 {
			if m.UI.Filter.AppendSearchChar(rune(key[0])) {
				return m.refreshFilteredTasks()
			}
		}
		return m, nil
	}
}

// refreshFilteredTasks runs the current filters and updates the task list.
// Delegates to fetchTasksForCurrentProject which handles both search and field filters.
func (m Model) refreshFilteredTasks() (tea.Model, tea.Cmd) {
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
	m.UIState.ResetTaskScrollOffsets()

	return m, nil
}

// openFilterPicker opens the appropriate picker modal for the currently focused filter chip.
func (m Model) openFilterPicker() (tea.Model, tea.Cmd) {
	switch m.UI.Filter.FocusedChip {
	case state.FilterChipPriority:
		if m.initPriorityPickerForFilter() {
			m.UIState.Mode = state.PriorityPickerMode
		}
		return m, nil

	case state.FilterChipType:
		if m.initTypePickerForFilter() {
			m.UIState.Mode = state.TypePickerMode
		}
		return m, nil

	case state.FilterChipAssignee:
		if m.initAssigneePickerForFilter() {
			m.UIState.Mode = state.AssigneePickerMode
		}
		return m, nil

	case state.FilterChipLabel:
		if m.initLabelPickerForFilter() {
			m.UIState.Mode = state.LabelPickerMode
		}
		return m, nil
	}

	return m, nil
}

// initPriorityPickerForFilter initializes the priority picker for filter bar mode.
func (m *Model) initPriorityPickerForFilter() bool {
	cursorPos := 0

	if m.UI.Filter.PriorityID != nil {
		cursorPos = max(*m.UI.Filter.PriorityID-1, 0)
	}

	m.Pickers.Priority.SetSelectedPriorityID(0)
	m.Pickers.Priority.SetCursor(cursorPos)
	m.Pickers.Priority.ReturnMode = state.FilterBarMode

	return true
}

// initTypePickerForFilter initializes the type picker for filter bar mode.
func (m *Model) initTypePickerForFilter() bool {
	cursorPos := 0

	if m.UI.Filter.TypeID != nil {
		cursorPos = max(*m.UI.Filter.TypeID-1, 0)
	}

	m.Pickers.Type.SetSelectedTypeID(0)
	m.Pickers.Type.SetCursor(cursorPos)
	m.Pickers.Type.ReturnMode = state.FilterBarMode

	return true
}

// initAssigneePickerForFilter initializes the assignee picker for filter bar mode.
func (m *Model) initAssigneePickerForFilter() bool {
	ctx, cancel := m.DBContext()
	defer cancel()

	assignees, err := m.App.AssigneeService.List(ctx)
	if err != nil {
		m.UI.Notification.Add(state.LevelError, "Failed to load assignees")
		return false
	}

	m.Pickers.Assignee.SetAssignees(assignees)

	currentAssigneeID := 0
	if m.UI.Filter.AssigneeID != nil {
		currentAssigneeID = *m.UI.Filter.AssigneeID
	}

	m.Pickers.Assignee.SetSelectedID(currentAssigneeID)

	// Position cursor: 0 = "Unassigned" sentinel, then assignees offset by 1
	cursorPos := 0
	if currentAssigneeID > 0 {
		for i, a := range assignees {
			if a.ID == currentAssigneeID {
				cursorPos = i + 1
				break
			}
		}
	}

	m.Pickers.Assignee.SetCursor(cursorPos)
	m.Pickers.Assignee.ReturnMode = state.FilterBarMode

	return true
}

// initLabelPickerForFilter initializes the label picker for filter bar mode.
func (m *Model) initLabelPickerForFilter() bool {
	project := m.getCurrentProject()
	if project == nil {
		return false
	}

	labelIDMap := make(map[int]bool)
	for _, labelID := range m.UI.Filter.LabelIDs {
		labelIDMap[labelID] = true
	}

	var items []state.LabelPickerItem
	for _, label := range m.AppState.Labels() {
		items = append(items, state.LabelPickerItem{
			Label:    label,
			Selected: labelIDMap[label.ID],
		})
	}

	m.Pickers.Label.Items = items
	m.Pickers.Label.TaskID = 0
	m.Pickers.Label.Cursor = 0
	m.Pickers.Label.Filter = ""
	m.Pickers.Label.ReturnMode = state.FilterBarMode

	return true
}
