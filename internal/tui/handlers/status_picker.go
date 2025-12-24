package handlers

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// LIST VIEW HANDLERS
// ============================================================================

// HandleToggleView toggles between kanban and list view.
func HandleToggleView(m *tui.Model) tea.Cmd {
	m.ListViewState.ToggleView()

	// Sync selection when toggling views
	if m.ListViewState.IsListView() {
		modelops.SyncKanbanToListSelection(m)
	} else {
		modelops.SyncListToKanbanSelection(m)
	}
	return nil
}

// HandleChangeStatus opens the status picker for the selected task.
// Only works in list view mode.
func HandleChangeStatus(m *tui.Model) tea.Cmd {
	if !m.ListViewState.IsListView() {
		return nil // Only works in list view
	}

	task := modelops.GetSelectedListTask(m)
	if task == nil {
		m.NotificationState.Add(state.LevelError, "No task selected")
		return nil
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
	return nil
}

// HandleSortList cycles through sort options in list view.
func HandleSortList(m *tui.Model) tea.Cmd {
	if !m.ListViewState.IsListView() {
		return nil // Only works in list view
	}
	m.ListViewState.CycleSort()
	return nil
}

// HandleStatusPickerMode handles key events in status picker mode.
func HandleStatusPickerMode(m *tui.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return nil
	case "enter":
		return confirmStatusChange(m)
	case "j", "down":
		m.StatusPickerState.MoveDown()
		return nil
	case "k", "up":
		m.StatusPickerState.MoveUp()
		return nil
	}
	return nil
}

// confirmStatusChange moves the task to the selected column.
func confirmStatusChange(m *tui.Model) tea.Cmd {
	selectedCol := m.StatusPickerState.SelectedColumn()
	taskID := m.StatusPickerState.TaskID()

	if selectedCol == nil {
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return nil
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
		return nil
	}

	// Move task to new column in database
	ctx, cancel := m.DbContext()
	defer cancel()

	err := m.Repo.MoveTaskToColumn(ctx, taskID, selectedCol.ID)
	if err != nil {
		m.HandleDBError(err, "Moving task to new status")
		m.StatusPickerState.Reset()
		m.UiState.SetMode(state.NormalMode)
		return nil
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
	return nil
}
