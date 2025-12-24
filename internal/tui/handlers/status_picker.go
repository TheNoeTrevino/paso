package handlers

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ============================================================================
// LIST VIEW HANDLERS
// ============================================================================

// HandleToggleView toggles between kanban and list view.
func (w *Wrapper) HandleToggleView() (*Wrapper, tea.Cmd) {
	ops := modelops.New(w.Model)
	w.ListViewState.ToggleView()

	// Sync selection when toggling views
	if w.ListViewState.IsListView() {
		ops.SyncKanbanToListSelection()
	} else {
		ops.SyncListToKanbanSelection()
	}
	return w, nil
}

// HandleChangeStatus opens the status picker for the selected task.
// Only works in list view mode.
func (w *Wrapper) HandleChangeStatus() (*Wrapper, tea.Cmd) {
	if !w.ListViewState.IsListView() {
		return w, nil // Only works in list view
	}

	ops := modelops.New(w.Model)
	task := ops.GetSelectedListTask()
	if task == nil {
		w.NotificationState.Add(state.LevelError, "No task selected")
		return w, nil
	}

	// Initialize status picker with columns
	w.StatusPickerState.SetTaskID(task.ID)
	w.StatusPickerState.SetColumns(w.AppState.Columns())

	// Set cursor to current column
	for i, col := range w.AppState.Columns() {
		if col.ID == task.ColumnID {
			w.StatusPickerState.SetCursor(i)
			break
		}
	}

	w.UiState.SetMode(state.StatusPickerMode)
	return w, nil
}

// HandleSortList cycles through sort options in list view.
func (w *Wrapper) HandleSortList() (*Wrapper, tea.Cmd) {
	if !w.ListViewState.IsListView() {
		return w, nil // Only works in list view
	}
	w.ListViewState.CycleSort()
	return w, nil
}

// HandleStatusPickerMode handles key events in status picker mode.
func (w *Wrapper) HandleStatusPickerMode(msg tea.KeyMsg) (*Wrapper, tea.Cmd) {
	switch msg.String() {
	case "esc":
		w.StatusPickerState.Reset()
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	case "enter":
		return w.confirmStatusChange()
	case "j", "down":
		w.StatusPickerState.MoveDown()
		return w, nil
	case "k", "up":
		w.StatusPickerState.MoveUp()
		return w, nil
	}
	return w, nil
}

// confirmStatusChange moves the task to the selected column.
func (w *Wrapper) confirmStatusChange() (*Wrapper, tea.Cmd) {
	selectedCol := w.StatusPickerState.SelectedColumn()
	taskID := w.StatusPickerState.TaskID()

	if selectedCol == nil {
		w.StatusPickerState.Reset()
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	}

	// Find the current task and its column
	var currentColumnID int
	var taskToMove *models.TaskSummary
	for colID, tasks := range w.AppState.Tasks() {
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
		w.StatusPickerState.Reset()
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	}

	// Move task to new column in database
	ctx, cancel := w.DbContext()
	defer cancel()

	err := w.Repo.MoveTaskToColumn(ctx, taskID, selectedCol.ID)
	if err != nil {
		w.HandleDBError(err, "Moving task to new status")
		w.StatusPickerState.Reset()
		w.UiState.SetMode(state.NormalMode)
		return w, nil
	}

	// Update local state: remove from current column
	if taskToMove != nil {
		currentTasks := w.AppState.Tasks()[currentColumnID]
		for i, t := range currentTasks {
			if t.ID == taskID {
				w.AppState.Tasks()[currentColumnID] = append(currentTasks[:i], currentTasks[i+1:]...)
				break
			}
		}

		// Add to new column
		newPosition := len(w.AppState.Tasks()[selectedCol.ID])
		taskToMove.ColumnID = selectedCol.ID
		taskToMove.Position = newPosition
		w.AppState.Tasks()[selectedCol.ID] = append(w.AppState.Tasks()[selectedCol.ID], taskToMove)
	}

	w.StatusPickerState.Reset()
	w.UiState.SetMode(state.NormalMode)
	return w, nil
}
