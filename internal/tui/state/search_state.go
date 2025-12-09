package state

// SearchState manages the vim-style search functionality state.
// This includes the search query text and whether the filter is currently active.
type SearchState struct {
	// Query is the current search text entered by the user
	Query string

	// IsActive indicates whether the search filter is applied
	// When true, the kanban view shows only matching tasks
	IsActive bool
}

// NewSearchState creates a new SearchState with default values.
func NewSearchState() *SearchState {
	return &SearchState{
		Query:    "",
		IsActive: false,
	}
}

// AppendChar appends a character to the search query.
// Returns true if the character was added, false if query is at max length.
func (s *SearchState) AppendChar(c rune) bool {
	const maxQueryLength = 100

	if len(s.Query) >= maxQueryLength {
		return false
	}

	s.Query += string(c)
	return true
}

// Backspace removes the last character from the search query.
// Returns true if a character was removed, false if query was already empty.
func (s *SearchState) Backspace() bool {
	if len(s.Query) == 0 {
		return false
	}

	s.Query = s.Query[:len(s.Query)-1]
	return true
}

// Clear resets the search query to empty string.
func (s *SearchState) Clear() {
	s.Query = ""
}

// Activate sets the filter as active.
// This is called when the user presses Enter in search mode.
func (s *SearchState) Activate() {
	s.IsActive = true
}

// Deactivate clears the filter.
// This is called when the user presses ESC in search mode.
func (s *SearchState) Deactivate() {
	s.IsActive = false
}
