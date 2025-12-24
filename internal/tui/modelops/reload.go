package modelops

import (
	"log/slog"
)

// ReloadProjects reloads the list of all projects from the database.
// Updates the AppState with fresh project data.
func (w *Wrapper) ReloadProjects() {
	ctx, cancel := w.DbContext()
	defer cancel()
	projects, err := w.Repo.GetAllProjects(ctx)
	if err != nil {
		slog.Error("Error reloading projects", "error", err)
		return
	}
	w.AppState.SetProjects(projects)
}

// ReloadCurrentProject reloads columns, tasks, and labels for the currently selected project.
// Preserves cursor position while updating data.
// Calls HandleDBError for each operation that fails.
func (w *Wrapper) ReloadCurrentProject() {
	currentProject := w.AppState.GetCurrentProject()
	if currentProject == nil {
		return
	}

	ctx, cancel := w.DbContext()
	defer cancel()

	// Reload columns
	columns, err := w.Repo.GetColumnsByProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("Error reloading columns", "error", err)
		w.HandleDBError(err, "reload columns")
		return
	}

	// Reload tasks
	tasks, err := w.Repo.GetTaskSummariesByProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("Error reloading tasks", "error", err)
		w.HandleDBError(err, "reload tasks")
		return
	}

	// Reload labels
	labels, err := w.Repo.GetLabelsByProject(ctx, currentProject.ID)
	if err != nil {
		slog.Error("Error reloading labels", "error", err)
		w.HandleDBError(err, "reload labels")
		return
	}

	// Update state with new data (preserves cursor position)
	w.AppState.SetColumns(columns)
	w.AppState.SetTasks(tasks)
	w.AppState.SetLabels(labels)
}
