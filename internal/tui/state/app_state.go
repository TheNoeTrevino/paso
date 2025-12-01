package state

import (
	"github.com/thenoetrevino/paso/internal/models"
)

// AppState manages the application's domain data.
// This includes projects, columns, tasks, and labels loaded from the database.
// It maintains the current project selection and provides methods to access
// the active project's data.
type AppState struct {
	// columns contains all columns for the current project, ordered by linked list traversal
	columns []*models.Column

	// tasks maps column IDs to their task summaries (includes labels)
	tasks map[int][]*models.TaskSummary

	// labels contains all labels available in the current project
	labels []*models.Label

	// projects contains all available projects in the database
	projects []*models.Project

	// selectedProject is the index of the currently active project in the projects slice
	selectedProject int

	// totalTaskCount caches the total number of tasks across all columns
	totalTaskCount int
}

// NewAppState creates a new AppState with the provided data.
// This constructor ensures all map fields are properly initialized.
func NewAppState(
	projects []*models.Project,
	selectedProject int,
	columns []*models.Column,
	tasks map[int][]*models.TaskSummary,
	labels []*models.Label,
) *AppState {
	// Ensure tasks map is never nil
	if tasks == nil {
		tasks = make(map[int][]*models.TaskSummary)
	}

	return &AppState{
		projects:        projects,
		selectedProject: selectedProject,
		columns:         columns,
		tasks:           tasks,
		labels:          labels,
		totalTaskCount:  calculateTotalTaskCount(tasks),
	}
}

// calculateTotalTaskCount computes the total number of tasks across all columns
func calculateTotalTaskCount(tasks map[int][]*models.TaskSummary) int {
	total := 0
	for _, columnTasks := range tasks {
		total += len(columnTasks)
	}
	return total
}

// GetCurrentProject returns the currently selected project.
// Returns nil if there are no projects or the selection index is invalid.
func (s *AppState) GetCurrentProject() *models.Project {
	if len(s.projects) == 0 {
		return nil
	}
	if s.selectedProject >= len(s.projects) || s.selectedProject < 0 {
		return nil
	}
	return s.projects[s.selectedProject]
}

// GetCurrentProjectID returns the ID of the currently selected project.
// Returns 0 if there is no valid project selected.
func (s *AppState) GetCurrentProjectID() int {
	project := s.GetCurrentProject()
	if project == nil {
		return 0
	}
	return project.ID
}

// Columns returns a copy of the columns slice to prevent external modification.
func (s *AppState) Columns() []*models.Column {
	// Return the slice directly - caller should not modify
	// In a future refactoring, we could return a copy for full immutability
	return s.columns
}

// SetColumns replaces the entire columns slice.
// This should be called after reloading columns from the database.
func (s *AppState) SetColumns(columns []*models.Column) {
	s.columns = columns
}

// Tasks returns the tasks map.
// Note: This returns the internal map - modifications will affect state.
func (s *AppState) Tasks() map[int][]*models.TaskSummary {
	return s.tasks
}

// SetTasks replaces the entire tasks map.
// This should be called after reloading all tasks from the database.
// It also recalculates the cached total task count.
func (s *AppState) SetTasks(tasks map[int][]*models.TaskSummary) {
	if tasks == nil {
		tasks = make(map[int][]*models.TaskSummary)
	}
	s.tasks = tasks
	s.totalTaskCount = calculateTotalTaskCount(tasks)
}

// TotalTaskCount returns the cached total number of tasks across all columns.
func (s *AppState) TotalTaskCount() int {
	return s.totalTaskCount
}

// Labels returns the labels slice.
func (s *AppState) Labels() []*models.Label {
	return s.labels
}

// SetLabels replaces the entire labels slice.
// This should be called after reloading labels from the database.
func (s *AppState) SetLabels(labels []*models.Label) {
	s.labels = labels
}

// Projects returns the projects slice.
func (s *AppState) Projects() []*models.Project {
	return s.projects
}

// SetProjects replaces the entire projects slice.
// This should be called after reloading projects from the database.
func (s *AppState) SetProjects(projects []*models.Project) {
	s.projects = projects
}

// SelectedProject returns the index of the currently selected project.
func (s *AppState) SelectedProject() int {
	return s.selectedProject
}

// SetSelectedProject updates the selected project index.
// No validation is performed - caller is responsible for ensuring valid index.
func (s *AppState) SetSelectedProject(index int) {
	s.selectedProject = index
}
