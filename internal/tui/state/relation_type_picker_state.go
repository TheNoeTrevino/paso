package state

// RelationTypePickerState manages the relation type picker modal state.
// This modal allows users to change the relation type when adding parent/child tasks.
type RelationTypePickerState struct {
	// selectedRelationTypeID is the currently selected relation type ID
	selectedRelationTypeID int

	// cursor is the current cursor position in the relation type picker
	cursor int

	// returnMode is the mode to return to after closing the picker
	returnMode Mode

	// Context for the picker - which task picker item is being edited
	currentTaskPickerIndex int
}

// NewRelationTypePickerState creates a new RelationTypePickerState with default values.
func NewRelationTypePickerState() *RelationTypePickerState {
	return &RelationTypePickerState{
		selectedRelationTypeID: 1, // Default to Parent/Child (id=1)
		cursor:                 0,
		returnMode:             ParentPickerMode,
		currentTaskPickerIndex: -1,
	}
}

// SelectedRelationTypeID returns the currently selected relation type ID.
func (s *RelationTypePickerState) SelectedRelationTypeID() int {
	return s.selectedRelationTypeID
}

// SetSelectedRelationTypeID updates the selected relation type ID.
func (s *RelationTypePickerState) SetSelectedRelationTypeID(id int) {
	s.selectedRelationTypeID = id
}

// Cursor returns the current cursor position.
func (s *RelationTypePickerState) Cursor() int {
	return s.cursor
}

// SetCursor updates the cursor position.
func (s *RelationTypePickerState) SetCursor(idx int) {
	s.cursor = idx
}

// ReturnMode returns the mode to return to after closing.
func (s *RelationTypePickerState) ReturnMode() Mode {
	return s.returnMode
}

// SetReturnMode updates the return mode.
func (s *RelationTypePickerState) SetReturnMode(mode Mode) {
	s.returnMode = mode
}

// CurrentTaskPickerIndex returns the index of the task picker item being edited.
func (s *RelationTypePickerState) CurrentTaskPickerIndex() int {
	return s.currentTaskPickerIndex
}

// SetCurrentTaskPickerIndex updates the task picker item index being edited.
func (s *RelationTypePickerState) SetCurrentTaskPickerIndex(idx int) {
	s.currentTaskPickerIndex = idx
}

// MoveUp moves the cursor up one position if possible.
func (s *RelationTypePickerState) MoveUp() {
	if s.cursor > 0 {
		s.cursor--
	}
}

// MoveDown moves the cursor down one position if possible.
// There are 3 relation types (indices 0-2).
func (s *RelationTypePickerState) MoveDown() {
	const maxRelationTypes = 3
	if s.cursor < maxRelationTypes-1 {
		s.cursor++
	}
}

// Reset resets all state to default values.
func (s *RelationTypePickerState) Reset() {
	s.selectedRelationTypeID = 1 // Default to Parent/Child
	s.cursor = 0
	s.returnMode = ParentPickerMode
	s.currentTaskPickerIndex = -1
}
