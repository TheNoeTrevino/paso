package modelops

import (
	"log/slog"

	"github.com/thenoetrevino/paso/internal/tui"
)

// ReloadProjects reloads the list of all projects from the database.
// Updates the AppState with fresh project data.
func ReloadProjects(m *tui.Model) {
	ctx, cancel := m.DbContext()
	defer cancel()
	projects, err := m.App.ProjectService.GetAllProjects(ctx)
	if err != nil {
		slog.Error("Error reloading projects", "error", err)
		return
	}
	m.AppState.SetProjects(projects)
}

// ReloadCurrentProject reloads columns, tasks, and labels for the currently selected project.
// Preserves cursor position while updating data.
// Calls HandleDBError for each operation that fails.
func ReloadCurrentProject(m *tui.Model) {
	currentProject := m.AppState.GetCurrentProject()
	if currentProject == nil {
		return
	}

	ctx, cancel := m.DbContext()
	defer cancel()

	// Reload columns
	columns, err := m.App.ColumnService.GetColumnsByProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("Error reloading columns", "error", err)
		m.HandleDBError(err, "reload columns")
		return
	}

	// Reload tasks
	tasks, err := m.App.TaskService.GetTaskSummariesByProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("Error reloading tasks", "error", err)
		m.HandleDBError(err, "reload tasks")
		return
	}

	// Reload labels
	labels, err := m.App.LabelService.GetLabelsByProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("Error reloading labels", "error", err)
		m.HandleDBError(err, "reload labels")
		return
	}

	// Update state with new data (preserves cursor position)
	m.AppState.SetColumns(columns)
	m.AppState.SetTasks(tasks)
	m.AppState.SetLabels(labels)
}
