package modelops

import (
	"log/slog"

	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

func RemoveCurrentTask(m *tui.Model) {
	currentCol := GetCurrentColumn(m)
	if currentCol == nil {
		return
	}

	tasks := GetTasksForColumn(m, currentCol.ID)

	if len(tasks) == 0 || m.UiState.SelectedTask() >= len(tasks) {
		return
	}

	// Remove the task at selectedTask index
	m.AppState.Tasks()[currentCol.ID] = append(tasks[:m.UiState.SelectedTask()], tasks[m.UiState.SelectedTask()+1:]...)

	// Adjust selectedTask if we removed the last task
	if m.UiState.SelectedTask() >= len(m.AppState.Tasks()[currentCol.ID]) && m.UiState.SelectedTask() > 0 {
		m.UiState.SetSelectedTask(m.UiState.SelectedTask() - 1)
	}
}

func RemoveCurrentColumn(m *tui.Model) {
	columns := m.AppState.Columns()
	selectedCol := m.UiState.SelectedColumn()

	if len(columns) == 0 || selectedCol >= len(columns) {
		return
	}

	// Remove the column at selectedColumn index
	m.AppState.SetColumns(append(columns[:selectedCol], columns[selectedCol+1:]...))

	// Adjust selectedColumn if we removed the last column
	if selectedCol >= len(m.AppState.Columns()) && selectedCol > 0 {
		m.UiState.SetSelectedColumn(selectedCol - 1)
	}

	// Reset task selection
	m.UiState.SetSelectedTask(0)

	// Adjust viewportOffset using UIState helper
	m.UiState.AdjustViewportAfterColumnRemoval(m.UiState.SelectedColumn(), len(m.AppState.Columns()))
}

func MoveTaskRight(m *tui.Model) {
	// Get the current task
	task := GetCurrentTask(m)
	if task == nil {
		return
	}

	// Check if there's a next column using the linked list
	currentCol := GetCurrentColumn(m)
	if currentCol == nil {
		return
	}
	if currentCol.NextID == nil {
		// Already at last column - show notification
		m.NotificationState.Add(state.LevelInfo, "There are no more columns to move to.")
		return
	}

	// Use the new database function to move task
	ctx, cancel := m.UiContext()
	defer cancel()
	err := m.Repo.MoveTaskToNextColumn(ctx, task.ID)
	if err != nil {
		slog.Error("Error moving task to next column", "error", err)
		if err != models.ErrAlreadyLastColumn {
			m.NotificationState.Add(state.LevelError, "Failed to move task to next column")
		}
		return
	}

	// Update local state: remove from current column
	tasks := m.AppState.Tasks()[currentCol.ID]
	m.AppState.Tasks()[currentCol.ID] = append(tasks[:m.UiState.SelectedTask()], tasks[m.UiState.SelectedTask()+1:]...)

	// Find the next column and add task there
	nextColID := *currentCol.NextID
	newPosition := len(m.AppState.Tasks()[nextColID])
	task.ColumnID = nextColID
	task.Position = newPosition
	m.AppState.Tasks()[nextColID] = append(m.AppState.Tasks()[nextColID], task)

	// Move selection to follow the task
	m.UiState.SetSelectedColumn(m.UiState.SelectedColumn() + 1)
	m.UiState.SetSelectedTask(newPosition)

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.UiState.SelectedColumn() >= m.UiState.ViewportOffset()+m.UiState.ViewportSize() {
		m.UiState.SetViewportOffset(m.UiState.ViewportOffset() + 1)
	}
}

func MoveTaskLeft(m *tui.Model) {
	// Get the current task
	task := GetCurrentTask(m)
	if task == nil {
		return
	}

	// Check if there's a previous column using the linked list
	currentCol := GetCurrentColumn(m)
	if currentCol == nil {
		return
	}
	if currentCol.PrevID == nil {
		// Already at first column - show notification
		m.NotificationState.Add(state.LevelInfo, "There are no more columns to move to.")
		return
	}

	// Use the new database function to move task
	ctx, cancel := m.UiContext()
	defer cancel()
	err := m.Repo.MoveTaskToPrevColumn(ctx, task.ID)
	if err != nil {
		slog.Error("Error moving task to previous column", "error", err)
		if err != models.ErrAlreadyFirstColumn {
			m.NotificationState.Add(state.LevelError, "Failed to move task to previous column")
		}
		return
	}

	// Update local state: remove from current column
	tasks := m.AppState.Tasks()[currentCol.ID]
	m.AppState.Tasks()[currentCol.ID] = append(tasks[:m.UiState.SelectedTask()], tasks[m.UiState.SelectedTask()+1:]...)

	// Find the previous column and add task there
	prevColID := *currentCol.PrevID
	newPosition := len(m.AppState.Tasks()[prevColID])
	task.ColumnID = prevColID
	task.Position = newPosition
	m.AppState.Tasks()[prevColID] = append(m.AppState.Tasks()[prevColID], task)

	// Move selection to follow the task
	m.UiState.SetSelectedColumn(m.UiState.SelectedColumn() - 1)
	m.UiState.SetSelectedTask(newPosition)

	// Ensure the moved task is visible (auto-scroll viewport if needed)
	if m.UiState.SelectedColumn() < m.UiState.ViewportOffset() {
		m.UiState.SetViewportOffset(m.UiState.ViewportOffset() - 1)
	}
}

func MoveTaskUp(m *tui.Model) {
	task := GetCurrentTask(m)
	if task == nil {
		return
	}

	// Check if already at top (edge case handled here for quick feedback)
	if m.UiState.SelectedTask() == 0 {
		m.NotificationState.Add(state.LevelInfo, "Task is already at the top")
		return
	}

	// Call database swap
	ctx, cancel := m.UiContext()
	defer cancel()
	err := m.Repo.SwapTaskUp(ctx, task.ID)
	if err != nil {
		slog.Error("Error moving task up", "error", err)
		if err != models.ErrAlreadyFirstTask {
			m.NotificationState.Add(state.LevelError, "Failed to move task up")
		}
		return
	}

	// Update local state: swap tasks in slice
	currentCol := GetCurrentColumn(m)
	if currentCol == nil {
		return
	}

	tasks := GetTasksForColumn(m, currentCol.ID)
	if len(tasks) < 2 {
		return
	}

	selectedIdx := m.UiState.SelectedTask()
	if selectedIdx == 0 || selectedIdx >= len(tasks) {
		return
	}

	// Swap positions in slice
	tasks[selectedIdx], tasks[selectedIdx-1] = tasks[selectedIdx-1], tasks[selectedIdx]

	// Update position values on the task objects
	tasks[selectedIdx].Position = selectedIdx
	tasks[selectedIdx-1].Position = selectedIdx - 1

	// Move selection to follow the task
	m.UiState.SetSelectedTask(selectedIdx - 1)
}

func MoveTaskDown(m *tui.Model) {
	task := GetCurrentTask(m)
	if task == nil {
		return
	}

	// Get current tasks for edge case check
	currentCol := GetCurrentColumn(m)
	if currentCol == nil {
		return
	}

	tasks := GetTasksForColumn(m, currentCol.ID)
	selectedIdx := m.UiState.SelectedTask()

	// Check if already at bottom
	if selectedIdx >= len(tasks)-1 {
		m.NotificationState.Add(state.LevelInfo, "Task is already at the bottom")
		return
	}

	// Call database swap
	ctx, cancel := m.UiContext()
	defer cancel()
	err := m.Repo.SwapTaskDown(ctx, task.ID)
	if err != nil {
		slog.Error("Error moving task down", "error", err)
		if err != models.ErrAlreadyLastTask {
			m.NotificationState.Add(state.LevelError, "Failed to move task down")
		}
		return
	}

	// Update local state: swap tasks in slice
	tasks[selectedIdx], tasks[selectedIdx+1] = tasks[selectedIdx+1], tasks[selectedIdx]

	// Update position values on the task objects
	tasks[selectedIdx].Position = selectedIdx
	tasks[selectedIdx+1].Position = selectedIdx + 1

	// Move selection to follow the task
	m.UiState.SetSelectedTask(selectedIdx + 1)
}

// getCurrentProject returns the currently selected project

func SwitchToProject(m *tui.Model, projectIndex int) {
	if projectIndex < 0 || projectIndex >= len(m.AppState.Projects()) {
		return
	}

	// Update state
	m.AppState.SetSelectedProject(projectIndex)

	project := m.AppState.Projects()[projectIndex]

	// Create context for database operations
	ctx, cancel := m.DbContext()
	defer cancel()

	// Reload columns for this project
	columns, err := m.Repo.GetColumnsByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading columns for project", "project_id", project.ID, "error", err)
		columns = []*models.Column{}
	}
	m.AppState.SetColumns(columns)

	// Reload task summaries for the entire project
	tasks, err := m.Repo.GetTaskSummariesByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading tasks for project", "project_id", project.ID, "error", err)
		tasks = make(map[int][]*models.TaskSummary)
	}
	m.AppState.SetTasks(tasks)

	// Reload labels for this project
	labels, err := m.Repo.GetLabelsByProject(ctx, project.ID)
	if err != nil {
		slog.Error("Error loading labels for project", "project_id", project.ID, "error", err)
		labels = []*models.Label{}
	}
	m.AppState.SetLabels(labels)

	// Reset selection state
	m.UiState.ResetSelection()
}
