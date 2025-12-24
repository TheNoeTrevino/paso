package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

func (m Model) handleToggleView() (tea.Model, tea.Cmd) {
	m.ListViewState.ToggleView()

	if m.ListViewState.IsListView() {
		m.syncKanbanToListSelection()
	} else {
		m.syncListToKanbanSelection()
	}
	return m, nil
}

func (m Model) handleChangeStatus() (tea.Model, tea.Cmd) {
	if !m.ListViewState.IsListView() {
		return m, nil
	}

	task := m.getSelectedListTask()
	if task == nil {
		m.NotificationState.Add(state.LevelError, "No task selected")
		return m, nil
	}

	m.StatusPickerState.SetTaskID(task.ID)
	m.StatusPickerState.SetColumns(m.AppState.Columns())

	for i, col := range m.AppState.Columns() {
		if col.ID == task.ColumnID {
			m.StatusPickerState.SetCursor(i)
			break
		}
	}

	m.UiState.SetMode(state.StatusPickerMode)
	return m, nil
}

func (m Model) handleSortList() (tea.Model, tea.Cmd) {
	if !m.ListViewState.IsListView() {
		return m, nil
	}
	m.ListViewState.CycleSort()
	return m, nil
}

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

func (m Model) confirmStatusChange() (tea.Model, tea.Cmd) {
	selectedCol := m.StatusPickerState.SelectedColumn()
	taskID := m.StatusPickerState.TaskID()

	if selectedCol == nil {
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

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

	if currentColumnID == selectedCol.ID {
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

	ctx, cancel := m.DbContext()
	defer cancel()

	err := m.App.TaskService.MoveTaskToColumn(ctx, taskID, selectedCol.ID)
	if err != nil {
		m.HandleDBError(err, "Moving task to new status")
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return m, nil
	}

	if taskToMove != nil {
		currentTasks := m.AppState.Tasks()[currentColumnID]
		for i, t := range currentTasks {
			if t.ID == taskID {
				m.AppState.Tasks()[currentColumnID] = append(currentTasks[:i], currentTasks[i+1:]...)
				break
			}
		}

		newPosition := len(m.AppState.Tasks()[selectedCol.ID])
		taskToMove.ColumnID = selectedCol.ID
		taskToMove.Position = newPosition
		m.AppState.Tasks()[selectedCol.ID] = append(m.AppState.Tasks()[selectedCol.ID], taskToMove)
	}

	m.StatusPickerState.Reset()
	m.UiState.SetMode(state.NormalMode)
	return m, nil
}
