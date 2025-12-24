package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// LIST VIEW HANDLERS
// ============================================================================

// handleToggleView toggles between kanban and list view.
func (m Model) handleToggleView() (tea.Model, tea.Cmd) {
	m.ListViewState.ToggleView()

	// Sync selection when toggling views
	if m.ListViewState.IsListView() {
		m.syncKanbanToListSelection()
	} else {
		m.syncListToKanbanSelection()
	}
	return m, nil
}

// handleChangeStatus opens the status picker for the selected task.
// Only works in list view mode.
func (m Model) handleChangeStatus() (tea.Model, tea.Cmd) {
	if !m.ListViewState.IsListView() {
		return m, nil // Only works in list view
	}

	task := m.getSelectedListTask()
	if task == nil {
		m.NotificationState.Add(state.LevelError, "No task selected")
		return m, nil
	}

	// Initialize status picker with columns
	m.StatusPickerState.SetTaskID(task.ID)
	m.StatusPickerState.SetColumns(m.AppState.Columns())

	// Set cursor to current column
	for i, col := range m.AppState.Columns() {
		if col.ID == task.ColumnID {
			m.StatusPickerState.SetCursor(i)
			break
		}
	}

	m.UiState.SetMode(state.StatusPickerMode)
	return m, nil
}

// handleSortList cycles through sort options in list view.
func (m Model) handleSortList() (tea.Model, tea.Cmd) {
	if !m.ListViewState.IsListView() {
		return m, nil // Only works in list view
	}
	m.ListViewState.CycleSort()
	return m, nil
}

// handleStatusPickerMode handles key events in status picker mode.
func (m Model) handleStatusPickerMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	case "enter":
		return m.confirmStatusChange()
	case "j", "down":
		m.StatusPickerState.MoveDown()
		return m, nil
	case "k", "up":
		m.StatusPickerState.MoveUp()
		return m, nil
	}
	return m, nil
}

// confirmStatusChange moves the task to the selected column.
func (m Model) confirmStatusChange() (tea.Model, tea.Cmd) {
	selectedCol := m.StatusPickerState.SelectedColumn()
	taskID := m.StatusPickerState.TaskID()

	if selectedCol == nil {
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Find the current task and its column
	var currentColumnID int
	var taskToMove *models.TaskSummary
	for colID, tasks := range m.AppState.Tasks() {
		for _, task := range tasks {
			if task.ID == taskID {
				currentColumnID = colID
				taskToMove = task
				break
			}
		}
		if taskToMove != nil {
			break
		}
	}

	// If already in the target column, just close the picker
	if currentColumnID == selectedCol.ID {
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Move task to new column in database
	ctx, cancel := m.DbContext()
	defer cancel()

	err := m.Repo.MoveTaskToColumn(ctx, taskID, selectedCol.ID)
	if err != nil {
		m.handleDBError(err, "Moving task to new status")
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Update local state: remove from current column
	if taskToMove != nil {
		currentTasks := m.AppState.Tasks()[currentColumnID]
		for i, t := range currentTasks {
			if t.ID == taskID {
				m.AppState.Tasks()[currentColumnID] = append(currentTasks[:i], currentTasks[i+1:]...)
				break
			}
		}

		// Add to new column
		newPosition := len(m.AppState.Tasks()[selectedCol.ID])
		taskToMove.ColumnID = selectedCol.ID
		taskToMove.Position = newPosition
		m.AppState.Tasks()[selectedCol.ID] = append(m.AppState.Tasks()[selectedCol.ID], taskToMove)
	}

	m.StatusPickerState.Reset()
	m.UiState.SetMode(state.NormalMode)
	return m, nil
}
