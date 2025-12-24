package modelops

import (
	"github.com/thenoetrevino/paso/internal/models"
)

// GetCurrentTasks returns the task summaries for the currently selected column
// Returns an empty slice if the column has no tasks
func (w *Wrapper) GetCurrentTasks() []*models.TaskSummary {
	if len(w.AppState.Columns()) == 0 {
		return []*models.TaskSummary{}
	}
	if w.UiState.SelectedColumn() >= len(w.AppState.Columns()) {
		return []*models.TaskSummary{}
	}
	currentCol := w.AppState.Columns()[w.UiState.SelectedColumn()]
	tasks := w.AppState.Tasks()[currentCol.ID]
	if tasks == nil {
		return []*models.TaskSummary{}
	}
	return tasks
}

// GetCurrentTask returns the currently selected task summary
// Returns nil if there are no tasks in the current column or no columns exist
func (w *Wrapper) GetCurrentTask() *models.TaskSummary {
	tasks := w.GetCurrentTasks()
	if len(tasks) == 0 {
		return nil
	}
	if w.UiState.SelectedTask() >= len(tasks) {
		return nil
	}
	return tasks[w.UiState.SelectedTask()]
}

// GetCurrentColumn returns the currently selected column
// Returns nil if there are no columns
func (w *Wrapper) GetCurrentColumn() *models.Column {
	if len(w.AppState.Columns()) == 0 {
		return nil
	}
	selectedIdx := w.UiState.SelectedColumn()
	if selectedIdx < 0 || selectedIdx >= len(w.AppState.Columns()) {
		return nil
	}
	return w.AppState.Columns()[selectedIdx]
}

// GetTasksForColumn returns tasks for a specific column ID with safe map access.
// Returns an empty slice if the column ID doesn't exist in the tasks map.
func (w *Wrapper) GetTasksForColumn(columnID int) []*models.TaskSummary {
	tasks, ok := w.AppState.Tasks()[columnID]
	if !ok || tasks == nil {
		return []*models.TaskSummary{}
	}
	return tasks
}

// GetCurrentProject returns the currently selected project
// Returns nil if there are no projects
func (w *Wrapper) GetCurrentProject() *models.Project {
	projects := w.AppState.Projects()
	selectedIdx := w.AppState.SelectedProject()
	if len(projects) == 0 || selectedIdx < 0 || selectedIdx >= len(projects) {
		return nil
	}
	return projects[selectedIdx]
}
