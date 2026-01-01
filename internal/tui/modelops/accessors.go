package modelops

import (
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui"
)

// GetCurrentTasks returns the task summaries for the currently selected column
// Returns an empty slice if the column has no tasks
func GetCurrentTasks(m *tui.Model) []*models.TaskSummary {
	if len(m.AppState.Columns()) == 0 {
		return []*models.TaskSummary{}
	}
	if m.UIState.SelectedColumn() >= len(m.AppState.Columns()) {
		return []*models.TaskSummary{}
	}
	currentCol := m.AppState.Columns()[m.UIState.SelectedColumn()]
	tasks := m.AppState.Tasks()[currentCol.ID]
	if tasks == nil {
		return []*models.TaskSummary{}
	}
	return tasks
}

// GetCurrentTask returns the currently selected task summary
// Returns nil if there are no tasks in the current column or no columns exist
func GetCurrentTask(m *tui.Model) *models.TaskSummary {
	tasks := GetCurrentTasks(m)
	if len(tasks) == 0 {
		return nil
	}
	if m.UIState.SelectedTask() >= len(tasks) {
		return nil
	}
	return tasks[m.UIState.SelectedTask()]
}

// GetCurrentColumn returns the currently selected column
// Returns nil if there are no columns
func GetCurrentColumn(m *tui.Model) *models.Column {
	if len(m.AppState.Columns()) == 0 {
		return nil
	}
	selectedIdx := m.UIState.SelectedColumn()
	if selectedIdx < 0 || selectedIdx >= len(m.AppState.Columns()) {
		return nil
	}
	return m.AppState.Columns()[selectedIdx]
}

// GetTasksForColumn returns tasks for a specific column ID with safe map access.
// Returns an empty slice if the column ID doesn't exist in the tasks map.
func GetTasksForColumn(m *tui.Model, columnID int) []*models.TaskSummary {
	tasks, ok := m.AppState.Tasks()[columnID]
	if !ok || tasks == nil {
		return []*models.TaskSummary{}
	}
	return tasks
}

// GetCurrentProject returns the currently selected project
// Returns nil if there are no projects
func GetCurrentProject(m *tui.Model) *models.Project {
	projects := m.AppState.Projects()
	selectedIdx := m.AppState.SelectedProject()
	if len(projects) == 0 || selectedIdx < 0 || selectedIdx >= len(projects) {
		return nil
	}
	return projects[selectedIdx]
}
