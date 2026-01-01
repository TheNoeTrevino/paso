package state

// TypePickerState manages the type picker modal state.
// This modal allows users to change the type of a task (task vs feature).
type TypePickerState struct {
	// selectedTypeID is the currently selected type ID
	selectedTypeID int

	// cursor is the current cursor position in the type picker
	cursor int

	// ReturnMode is the mode to return to after closing the picker
	ReturnMode Mode
}

// NewTypePickerState creates a new TypePickerState with default values.
func NewTypePickerState() *TypePickerState {
	return &TypePickerState{
		selectedTypeID: 1, // Default to task (id=1)
		cursor:         0,
		ReturnMode:     TicketFormMode,
	}
}

// SelectedTypeID returns the currently selected type ID.
func (s *TypePickerState) SelectedTypeID() int {
	return s.selectedTypeID
}

// SetSelectedTypeID updates the selected type ID.
func (s *TypePickerState) SetSelectedTypeID(id int) {
	s.selectedTypeID = id
}

// Cursor returns the current cursor position.
func (s *TypePickerState) Cursor() int {
	return s.cursor
}

// SetCursor updates the cursor position.
func (s *TypePickerState) SetCursor(idx int) {
	s.cursor = idx
}

// MoveUp moves the cursor up one position if possible.
func (s *TypePickerState) MoveUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

// MoveDown moves the cursor down one position if possible.
// There are 2 type options (indices 0-1).
func (s *TypePickerState) MoveDown() {
	const maxTypes = 2
	if s.cursor < maxTypes-1 {
		s.cursor++
	}
}

// Reset resets all state to default values.
func (s *TypePickerState) Reset() {
	s.selectedTypeID = 1 // Default to task
	s.cursor = 0
	s.ReturnMode = TicketFormMode
}
