package state

import "github.com/thenoetrevino/paso/internal/models"

// CommentItem represents a single comment for display.
type CommentItem struct {
	// Comment is the comment data from the database
	Comment *models.Comment
}

// CommentState manages the comments section state for a task.
// This includes displaying comments and navigating them.
// Individual comment editing is handled through CommentFormMode with huh forms.
type CommentState struct {
	// Items contains all comments for the current task
	Items []CommentItem

	// Cursor is the current cursor position in the comments list
	Cursor int

	// TaskID is the ID of the task being viewed
	TaskID int

	// ScrollOffset is the vertical scroll offset for the comments list
	ScrollOffset int
}

// NewCommentState creates a new CommentState with default values.
func NewCommentState() *CommentState {
	return &CommentState{
		Items:        []CommentItem{},
		Cursor:       0,
		TaskID:       0,
		ScrollOffset: 0,
	}
}

// Clear resets all state to default values.
func (s *CommentState) Clear() {
	s.Items = []CommentItem{}
	s.Cursor = 0
	s.TaskID = 0
	s.ScrollOffset = 0
}

// MoveCursorUp moves the cursor up one position if possible.
// Returns true if the cursor moved, false if already at top.
func (s *CommentState) MoveCursorUp() bool {
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
func (s *CommentState) MoveCursorDown(maxIdx int) bool {
	if s.Cursor < maxIdx {
		s.Cursor++
		return true
	}
	return false
}

// GetSelectedComment returns the currently selected comment, or nil if none
func (s *CommentState) GetSelectedComment() *models.Comment {
	if s.Cursor >= 0 && s.Cursor < len(s.Items) {
		return s.Items[s.Cursor].Comment
	}
	return nil
}

// SetComments replaces the comment list with new data
func (s *CommentState) SetComments(comments []*models.Comment) {
	s.Items = make([]CommentItem, len(comments))
	for i, c := range comments {
		s.Items[i] = CommentItem{Comment: c}
	}

	// Reset cursor if out of bounds
	if s.Cursor >= len(s.Items) {
		if len(s.Items) > 0 {
			s.Cursor = len(s.Items) - 1
		} else {
			s.Cursor = 0
		}
	}

	// Reset scroll offset if out of bounds
	if s.ScrollOffset >= len(s.Items) && len(s.Items) > 0 {
		s.ScrollOffset = max(0, len(s.Items)-1)
	}
}

// IsEmpty returns true if there are no comments
func (s *CommentState) IsEmpty() bool {
	return len(s.Items) == 0
}

// DeleteSelected removes the currently selected comment from the list
// and adjusts the cursor position appropriately
func (s *CommentState) DeleteSelected() {
	if s.Cursor < 0 || s.Cursor >= len(s.Items) {
		return
	}

	// Remove the item at cursor
	s.Items = append(s.Items[:s.Cursor], s.Items[s.Cursor+1:]...)

	// Adjust cursor
	if len(s.Items) == 0 {
		s.Cursor = 0
	} else if s.Cursor >= len(s.Items) {
		s.Cursor = len(s.Items) - 1
	}
}

// EnsureCommentVisible adjusts scroll offset to keep the cursor visible.
// Mirrors the EnsureTaskVisible pattern from UIState.
//
// Parameters:
//   - maxVisible: maximum number of comments that can be displayed at once
func (s *CommentState) EnsureCommentVisible(maxVisible int) {
	// If cursor is above visible area, scroll up
	if s.Cursor < s.ScrollOffset {
		s.ScrollOffset = s.Cursor
	}

	// If cursor is below visible area, scroll down
	if s.Cursor >= s.ScrollOffset+maxVisible {
		s.ScrollOffset = s.Cursor - maxVisible + 1
	}

	// Ensure scroll offset never goes negative
	if s.ScrollOffset < 0 {
		s.ScrollOffset = 0
	}
}
