package state

import "github.com/thenoetrevino/paso/internal/models"

// AssigneePickerState manages the assignee picker modal state.
// This modal allows users to assign a task to an assignee.
type AssigneePickerState struct {
	assignees    []*models.Assignee
	selectedID   int
	cursor       int
	ReturnMode   Mode
	showClearOpt bool // Whether to show "clear assignee" as first option
}

// NewAssigneePickerState creates a new AssigneePickerState with default values.
func NewAssigneePickerState() *AssigneePickerState {
	return &AssigneePickerState{
		ReturnMode:   TaskFormMode,
		showClearOpt: true,
	}
}

// Assignees returns the list of available assignees.
func (s *AssigneePickerState) Assignees() []*models.Assignee {
	return s.assignees
}

// SetAssignees sets the available assignees and resets the cursor.
func (s *AssigneePickerState) SetAssignees(assignees []*models.Assignee) {
	s.assignees = assignees
	s.cursor = 0
}

// SelectedID returns the currently selected assignee ID (0 means no assignee).
func (s *AssigneePickerState) SelectedID() int {
	return s.selectedID
}

// SetSelectedID sets the currently selected assignee ID.
func (s *AssigneePickerState) SetSelectedID(id int) {
	s.selectedID = id
}

// Cursor returns the current cursor position.
func (s *AssigneePickerState) Cursor() int {
	return s.cursor
}

// SetCursor updates the cursor position.
func (s *AssigneePickerState) SetCursor(idx int) {
	s.cursor = idx
}

// itemCount returns the total number of selectable items.
func (s *AssigneePickerState) itemCount() int {
	count := len(s.assignees)
	if s.showClearOpt {
		count++
	}
	return count
}

// MoveUp moves the cursor up one position.
func (s *AssigneePickerState) MoveUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

// MoveDown moves the cursor down one position.
func (s *AssigneePickerState) MoveDown() {
	max := s.itemCount() - 1
	if s.cursor < max {
		s.cursor++
	}
}

// IsClearSelected returns true if the cursor is on the "clear assignee" option.
func (s *AssigneePickerState) IsClearSelected() bool {
	return s.showClearOpt && s.cursor == 0
}

// SelectedAssignee returns the assignee at the current cursor position,
// or nil if "clear" is selected or no assignees are loaded.
func (s *AssigneePickerState) SelectedAssignee() *models.Assignee {
	if s.IsClearSelected() {
		return nil
	}
	idx := s.cursor
	if s.showClearOpt {
		idx--
	}
	if idx < 0 || idx >= len(s.assignees) {
		return nil
	}
	return s.assignees[idx]
}

// Reset resets all state to default values.
func (s *AssigneePickerState) Reset() {
	s.assignees = nil
	s.selectedID = 0
	s.cursor = 0
	s.ReturnMode = TaskFormMode
}
