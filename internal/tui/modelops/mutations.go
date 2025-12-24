package modelops

import (
	"log/slog"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

func (w *Wrapper) RemoveCurrentTask() {
	currentCol := w.GetCurrentColumn()
	if currentCol == nil {
		return
	}

	tasks := w.GetTasksForColumn(currentCol.ID)

	if len(tasks) == 0 || w.UiState.SelectedTask() >= len(tasks) {
		return
	}

	// Remove the task at selectedTask index
	w.AppState.Tasks()[currentCol.ID] = append(tasks[:w.UiState.SelectedTask()], tasks[w.UiState.SelectedTask()+1:]...)

	// Adjust selectedTask if we removed the last task
	if w.UiState.SelectedTask() >= len(w.AppState.Tasks()[currentCol.ID]) && w.UiState.SelectedTask() > 0 {
		w.UiState.SetSelectedTask(w.UiState.SelectedTask() - 1)
	}
}

func (w *Wrapper) RemoveCurrentColumn() {
	columns := w.AppState.Columns()
	selectedCol := w.UiState.SelectedColumn()

	if len(columns) == 0 || selectedCol >= len(columns) {
		return
	}

	// Remove the column at selectedColumn index
	w.AppState.SetColumns(append(columns[:selectedCol], columns[selectedCol+1:]...))

	// Adjust selectedColumn if we removed the last column
	if selectedCol >= len(w.AppState.Columns()) && selectedCol > 0 {
		w.UiState.SetSelectedColumn(selectedCol - 1)
	}

	// Reset task selection
	w.UiState.SetSelectedTask(0)

	// Adjust viewportOffset using UIState helper
	w.UiState.AdjustViewportAfterColumnRemoval(w.UiState.SelectedColumn(), len(w.AppState.Columns()))
}

func (w *Wrapper) MoveTaskRight() {
	// Get the current task
	task := w.GetCurrentTask()
	if task == nil {
		return
	}

	// Check if there's a next column using the linked list
	currentCol := w.GetCurrentColumn()
	if currentCol == nil {
		return
	}
	if currentCol.NextID == nil {
		// Already at last column - show notification
		w.NotificationState.Add(state.LevelInfo, "There are no more columns to move to.")
		return
	}

	// Use the new database function to move task
	ctx, cancel := w.UiContext()
	defer cancel()
	err := w.Repo.MoveTaskToNextColumn(ctx, task.ID)
	if err != nil {
		slog.Error("Error moving task to next column", "error", err)
		if err != models.ErrAlreadyLastColumn {
			w.NotificationState.Add(state.LevelError, "Failed to move task to next column")
		}
		return
	}

	// Update local state: remove from current column
	tasks := w.AppState.Tasks()[currentCol.ID]
	w.AppState.Tasks()[currentCol.ID] = append(tasks[:w.UiState.SelectedTask()], tasks[w.UiState.SelectedTask()+1:]...)

	// Find the next column and add task there
	nextColID := *currentCol.NextID
	newPosition := len(w.AppState.Tasks()[nextColID])
	task.ColumnID = nextColID
	task.Position = newPosition
	w.AppState.Tasks()[nextColID] = append(w.AppState.Tasks()[nextColID], task)

	// Move selection to follow the task
	w.UiState.SetSelectedColumn(w.UiState.SelectedColumn() + 1)
	w.UiState.SetSelectedTask(newPosition)

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if w.UiState.SelectedColumn() >= w.UiState.ViewportOffset()+w.UiState.ViewportSize() {
		w.UiState.SetViewportOffset(w.UiState.ViewportOffset() + 1)
	}
}

func (w *Wrapper) MoveTaskLeft() {
	// Get the current task
	task := w.GetCurrentTask()
	if task == nil {
		return
	}

	// Check if there's a previous column using the linked list
	currentCol := w.GetCurrentColumn()
	if currentCol == nil {
		return
	}
	if currentCol.PrevID == nil {
		// Already at first column - show notification
		w.NotificationState.Add(state.LevelInfo, "There are no more columns to move to.")
		return
	}

	// Use the new database function to move task
	ctx, cancel := w.UiContext()
	defer cancel()
	err := w.Repo.MoveTaskToPrevColumn(ctx, task.ID)
	if err != nil {
		slog.Error("Error moving task to previous column", "error", err)
		if err != models.ErrAlreadyFirstColumn {
			w.NotificationState.Add(state.LevelError, "Failed to move task to previous column")
		}
		return
	}

	// Update local state: remove from current column
	tasks := w.AppState.Tasks()[currentCol.ID]
	w.AppState.Tasks()[currentCol.ID] = append(tasks[:w.UiState.SelectedTask()], tasks[w.UiState.SelectedTask()+1:]...)

	// Find the previous column and add task there
	prevColID := *currentCol.PrevID
	newPosition := len(w.AppState.Tasks()[prevColID])
	task.ColumnID = prevColID
	task.Position = newPosition
	w.AppState.Tasks()[prevColID] = append(w.AppState.Tasks()[prevColID], task)

	// Move selection to follow the task
	w.UiState.SetSelectedColumn(w.UiState.SelectedColumn() - 1)
	w.UiState.SetSelectedTask(newPosition)

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if w.UiState.SelectedColumn() < w.UiState.ViewportOffset() {
		w.UiState.SetViewportOffset(w.UiState.ViewportOffset() - 1)
	}
}

func (w *Wrapper) MoveTaskUp() {
	task := w.GetCurrentTask()
	if task == nil {
		return
	}

	// Check if already at top (edge case handled here for quick feedback)
	if w.UiState.SelectedTask() == 0 {
		w.NotificationState.Add(state.LevelInfo, "Task is already at the top")
		return
	}

	// Call database swap
	ctx, cancel := w.UiContext()
	defer cancel()
	err := w.Repo.SwapTaskUp(ctx, task.ID)
	if err != nil {
		slog.Error("Error moving task up", "error", err)
		if err != models.ErrAlreadyFirstTask {
			w.NotificationState.Add(state.LevelError, "Failed to move task up")
		}
		return
	}

	// Update local state: swap tasks in slice
	currentCol := w.GetCurrentColumn()
	if currentCol == nil {
		return
	}

	tasks := w.GetTasksForColumn(currentCol.ID)
	if len(tasks) < 2 {
		return
	}

	selectedIdx := w.UiState.SelectedTask()
	if selectedIdx == 0 || selectedIdx >= len(tasks) {
		return
	}

	// Swap positions in slice
	tasks[selectedIdx], tasks[selectedIdx-1] = tasks[selectedIdx-1], tasks[selectedIdx]

	// Update position values on the task objects
	tasks[selectedIdx].Position = selectedIdx
	tasks[selectedIdx-1].Position = selectedIdx - 1

	// Move selection to follow the task
	w.UiState.SetSelectedTask(selectedIdx - 1)
}

func (w *Wrapper) MoveTaskDown() {
	task := w.GetCurrentTask()
	if task == nil {
		return
	}

	// Get current tasks for edge case check
	currentCol := w.GetCurrentColumn()
	if currentCol == nil {
		return
	}

	tasks := w.GetTasksForColumn(currentCol.ID)
	selectedIdx := w.UiState.SelectedTask()

	// Check if already at bottom
	if selectedIdx >= len(tasks)-1 {
		w.NotificationState.Add(state.LevelInfo, "Task is already at the bottom")
		return
	}

	// Call database swap
	ctx, cancel := w.UiContext()
	defer cancel()
	err := w.Repo.SwapTaskDown(ctx, task.ID)
	if err != nil {
		slog.Error("Error moving task down", "error", err)
		if err != models.ErrAlreadyLastTask {
			w.NotificationState.Add(state.LevelError, "Failed to move task down")
		}
		return
	}

	// Update local state: swap tasks in slice
	tasks[selectedIdx], tasks[selectedIdx+1] = tasks[selectedIdx+1], tasks[selectedIdx]

	// Update position values on the task objects
	tasks[selectedIdx].Position = selectedIdx
	tasks[selectedIdx+1].Position = selectedIdx + 1

	// Move selection to follow the task
	w.UiState.SetSelectedTask(selectedIdx + 1)
}

// getCurrentProject returns the currently selected project

func (w *Wrapper) SwitchToProject(projectIndex int) {
	if projectIndex < 0 || projectIndex >= len(w.AppState.Projects()) {
		return
	}

	// Update state
	w.AppState.SetSelectedProject(projectIndex)

	project := w.AppState.Projects()[projectIndex]

	// Create context for database operations
	ctx, cancel := w.DbContext()
	defer cancel()

	// Reload columns for this project
	columns, err := w.Repo.GetColumnsByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading columns for project", "project_id", project.ID, "error", err)
		columns = []*models.Column{}
	}
	w.AppState.SetColumns(columns)

	// Reload task summaries for the entire project
	tasks, err := w.Repo.GetTaskSummariesByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading tasks for project", "project_id", project.ID, "error", err)
		tasks = make(map[int][]*models.TaskSummary)
	}
	w.AppState.SetTasks(tasks)

	// Reload labels for this project
	labels, err := w.Repo.GetLabelsByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading labels for project", "project_id", project.ID, "error", err)
		labels = []*models.Label{}
	}
	w.AppState.SetLabels(labels)

	// Reset selection state
	w.UiState.ResetSelection()
}

