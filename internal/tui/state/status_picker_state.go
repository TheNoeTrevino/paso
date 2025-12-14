package state

import "github.com/thenoetrevino/paso/internal/models"

// StatusPickerState manages the status picker modal state.
// This modal allows users to change the status/column of a task in list view.
type StatusPickerState struct {
	// taskID is the ID of the task being edited
	taskID int

	// columns is the list of available columns/statuses to choose from
	columns []*models.Column

	// cursor is the current cursor position in the status picker
	cursor int
}

// NewStatusPickerState creates a new StatusPickerState with default values.
func NewStatusPickerState() *StatusPickerState {
	return &StatusPickerState{
		taskID:  0,
		columns: []*models.Column{},
		cursor:  0,
	}
}

// TaskID returns the ID of the task being edited.
func (s *StatusPickerState) TaskID() int {
	return s.taskID
}

// SetTaskID updates the task ID.
func (s *StatusPickerState) SetTaskID(id int) {
	s.taskID = id
}

// Columns returns the list of available columns.
func (s *StatusPickerState) Columns() []*models.Column {
	return s.columns
}

// SetColumns updates the list of available columns.
func (s *StatusPickerState) SetColumns(cols []*models.Column) {
	s.columns = cols
}

// Cursor returns the current cursor position.
func (s *StatusPickerState) Cursor() int {
	return s.cursor
}

// SetCursor updates the cursor position.
func (s *StatusPickerState) SetCursor(idx int) {
	s.cursor = idx
}

// MoveUp moves the cursor up one position if possible.
func (s *StatusPickerState) MoveUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

// MoveDown moves the cursor down one position if possible.
func (s *StatusPickerState) MoveDown() {
	if len(s.columns) > 0 && s.cursor < len(s.columns)-1 {
		s.cursor++
	}
}

// SelectedColumn returns the currently selected column.
// Returns nil if no columns are available or cursor is out of bounds.
func (s *StatusPickerState) SelectedColumn() *models.Column {
	if len(s.columns) == 0 || s.cursor < 0 || s.cursor >= len(s.columns) {
		return nil
	}
	return s.columns[s.cursor]
}

// Reset resets all state to default values.
func (s *StatusPickerState) Reset() {
	s.taskID = 0
	s.columns = []*models.Column{}
	s.cursor = 0
}
