package state

import "github.com/thenoetrevino/paso/internal/models"

// NoteItem represents a single note/comment for display.
type NoteItem struct {
	// Comment is the comment data from the database
	Comment *models.Comment
}

// NoteState manages the notes section state for a task.
// This includes displaying notes and navigating them.
// Individual note editing is handled through NoteFormMode with huh forms.
type NoteState struct {
	// Items contains all notes for the current task
	Items []NoteItem

	// Cursor is the current cursor position in the notes list
	Cursor int

	// TaskID is the ID of the task being viewed
	TaskID int

	// ScrollOffset is the vertical scroll offset for the notes list
	ScrollOffset int
}

// NewNoteState creates a new NoteState with default values.
func NewNoteState() *NoteState {
	return &NoteState{
		Items:        []NoteItem{},
		Cursor:       0,
		TaskID:       0,
		ScrollOffset: 0,
	}
}

// Clear resets all state to default values.
func (s *NoteState) Clear() {
	s.Items = []NoteItem{}
	s.Cursor = 0
	s.TaskID = 0
	s.ScrollOffset = 0
}

// MoveCursorUp moves the cursor up one position if possible.
// Returns true if the cursor moved, false if already at top.
func (s *NoteState) MoveCursorUp() bool {
	if s.Cursor > 0 {
		s.Cursor--
		return true
	}
	return false
}

// MoveCursorDown moves the cursor down one position if possible.
// Returns true if the cursor moved, false if already at bottom.
//
// Parameters:
//   - maxIdx: the maximum valid cursor position (typically len(items) - 1)
func (s *NoteState) MoveCursorDown(maxIdx int) bool {
	if s.Cursor < maxIdx {
		s.Cursor++
		return true
	}
	return false
}
