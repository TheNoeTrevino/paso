package state

import "github.com/thenoetrevino/paso/internal/models"

// ProjectPickerState manages the project picker modal state.
// This modal allows users to move a task to a different project.
type ProjectPickerState struct {
	projects   []*models.Project
	cursor     int
	ReturnMode Mode
}

// NewProjectPickerState creates a new ProjectPickerState with default values.
func NewProjectPickerState() *ProjectPickerState {
	return &ProjectPickerState{}
}

// Projects returns the list of available projects.
func (s *ProjectPickerState) Projects() []*models.Project {
	return s.projects
}

// SetProjects sets the available projects and resets the cursor.
func (s *ProjectPickerState) SetProjects(projects []*models.Project) {
	s.projects = projects
	s.cursor = 0
}

// Cursor returns the current cursor position.
func (s *ProjectPickerState) Cursor() int {
	return s.cursor
}

// SetCursor updates the cursor position.
func (s *ProjectPickerState) SetCursor(cursor int) {
	s.cursor = cursor
}

// MoveUp moves the cursor up one position.
func (s *ProjectPickerState) MoveUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

// MoveDown moves the cursor down one position.
func (s *ProjectPickerState) MoveDown() {
	max := len(s.projects) - 1
	if s.cursor < max {
		s.cursor++
	}
}

// SelectedProject returns the project at the current cursor position,
// or nil if no projects are loaded.
func (s *ProjectPickerState) SelectedProject() *models.Project {
	if len(s.projects) == 0 {
		return nil
	}
	if s.cursor < 0 || s.cursor >= len(s.projects) {
		return nil
	}
	return s.projects[s.cursor]
}

// Reset resets all state to default values.
func (s *ProjectPickerState) Reset() {
	s.projects = nil
	s.cursor = 0
}
