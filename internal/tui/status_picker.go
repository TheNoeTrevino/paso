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
	m.listViewState.ToggleView()

	// Sync selection when toggling views
	if m.listViewState.IsListView() {
		m.syncKanbanToListSelection()
	} else {
		m.syncListToKanbanSelection()
	}
	return m, nil
}

// handleChangeStatus opens the status picker for the selected task.
// Only works in list view mode.
func (m Model) handleChangeStatus() (tea.Model, tea.Cmd) {
	if !m.listViewState.IsListView() {
		return m, nil // Only works in list view
	}

	task := m.getSelectedListTask()
	if task == nil {
		m.notificationState.Add(state.LevelError, "No task selected")
		return m, nil
	}

	// Initialize status picker with columns
	m.statusPickerState.SetTaskID(task.ID)
	m.statusPickerState.SetColumns(m.appState.Columns())

	// Set cursor to current column
	for i, col := range m.appState.Columns() {
		if col.ID == task.ColumnID {
			m.statusPickerState.SetCursor(i)
			break
		}
	}

	m.uiState.SetMode(state.StatusPickerMode)
	return m, nil
}

// handleSortList cycles through sort options in list view.
func (m Model) handleSortList() (tea.Model, tea.Cmd) {
	if !m.listViewState.IsListView() {
		return m, nil // Only works in list view
	}
	m.listViewState.CycleSort()
	return m, nil
}

// handleStatusPickerMode handles key events in status picker mode.
func (m Model) handleStatusPickerMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.statusPickerState.Reset()
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	case "enter":
		return m.confirmStatusChange()
	case "j", "down":
		m.statusPickerState.MoveDown()
		return m, nil
	case "k", "up":
		m.statusPickerState.MoveUp()
		return m, nil
	}
	return m, nil
}

// confirmStatusChange moves the task to the selected column.
func (m Model) confirmStatusChange() (tea.Model, tea.Cmd) {
	selectedCol := m.statusPickerState.SelectedColumn()
	taskID := m.statusPickerState.TaskID()

	if selectedCol == nil {
		m.statusPickerState.Reset()
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Find the current task and its column
	var currentColumnID int
	var taskToMove *models.TaskSummary
	for colID, tasks := range m.appState.Tasks() {
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
		m.statusPickerState.Reset()
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Move task to new column in database
	ctx, cancel := m.dbContext()
	defer cancel()

	err := m.repo.MoveTaskToColumn(ctx, taskID, selectedCol.ID)
	if err != nil {
		m.handleDBError(err, "Moving task to new status")
		m.statusPickerState.Reset()
		m.uiState.SetMode(state.NormalMode)
		return m, nil
	}

	// Update local state: remove from current column
	if taskToMove != nil {
		currentTasks := m.appState.Tasks()[currentColumnID]
		for i, t := range currentTasks {
			if t.ID == taskID {
				m.appState.Tasks()[currentColumnID] = append(currentTasks[:i], currentTasks[i+1:]...)
				break
			}
		}

		// Add to new column
		newPosition := len(m.appState.Tasks()[selectedCol.ID])
		taskToMove.ColumnID = selectedCol.ID
		taskToMove.Position = newPosition
		m.appState.Tasks()[selectedCol.ID] = append(m.appState.Tasks()[selectedCol.ID], taskToMove)
	}

	m.statusPickerState.Reset()
	m.uiState.SetMode(state.NormalMode)
	return m, nil
}
