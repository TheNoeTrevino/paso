package state

import "github.com/mattn/go-runewidth"

// FilterChip represents the index of each chip in the filter bar.
type FilterChip int

const (
	FilterChipSearch   FilterChip = iota // Search filter chip (inline text input)
	FilterChipLabel                      // Label filter chip
	FilterChipPriority                   // Priority filter chip
	FilterChipType                       // Type filter chip
	FilterChipAssignee                   // Assignee filter chip
	FilterChipArchived                   // Archived filter chip (toggle)
	FilterChipClearAll                   // Clear All pseudo-chip
	FilterChipCount                      // Total number of chips (used for bounds)
)

// FilterState manages the filter bar selections, focus state, and search query.
// Each filter field is nullable — nil means "no filter applied".
type FilterState struct {
	PriorityID        *int       // Selected priority ID, nil = no filter
	TypeID            *int       // Selected type ID, nil = no filter
	AssigneeID        *int       // Selected assignee ID, nil = no filter (-1 = "Unassigned")
	AssigneeName      string     // Resolved display name for the assignee filter
	LabelIDs          []int      // Selected label IDs (OR logic), empty = no filter
	ShowArchived      bool       // Show archived tasks (false = hide archived, true = show all)
	FocusedChip       FilterChip // Currently focused chip in the filter bar
	IsActive          bool       // Whether the filter bar is focused for keyboard navigation
	SearchQuery       string     // Current search text for title filtering
	SearchInputActive bool       // Whether the search chip is in text input mode
}

const maxSearchQueryLength = 100

// NewFilterState creates a new FilterState with default values.
func NewFilterState() *FilterState {
	return &FilterState{
		FocusedChip: FilterChipSearch,
	}
}

// HasAnyFilter returns true if any filter is currently applied.
func (f *FilterState) HasAnyFilter() bool {
	return f.PriorityID != nil ||
		f.TypeID != nil ||
		f.AssigneeID != nil ||
		len(f.LabelIDs) > 0 ||
		f.ShowArchived ||
		f.SearchQuery != ""
}

// ClearAll resets all filter selections to their default (unfiltered) state,
// including the search query.
func (f *FilterState) ClearAll() {
	f.ClearFieldFilters()
	f.SearchQuery = ""
	f.SearchInputActive = false
}

// ClearFieldFilters resets only the field-level filter selections (priority, type,
// assignee, labels, archived) without affecting the search query. Used when switching
// projects since field filters are project-scoped but search should persist.
func (f *FilterState) ClearFieldFilters() {
	f.PriorityID = nil
	f.TypeID = nil
	f.AssigneeID = nil
	f.AssigneeName = ""
	f.LabelIDs = nil
	f.ShowArchived = false
}

// MoveFocusLeft moves the focused chip one position to the left.
// Wraps around to the last chip if at the beginning.
func (f *FilterState) MoveFocusLeft() {
	if f.FocusedChip <= 0 {
		f.FocusedChip = FilterChipCount - 1
	} else {
		f.FocusedChip--
	}
}

// MoveFocusRight moves the focused chip one position to the right.
// Wraps around to the first chip if at the end.
func (f *FilterState) MoveFocusRight() {
	if f.FocusedChip >= FilterChipCount-1 {
		f.FocusedChip = 0
	} else {
		f.FocusedChip++
	}
}

// ClearFocusedChip clears the filter for the currently focused chip.
// Returns true if a filter was cleared, false if the chip had no active filter.
func (f *FilterState) ClearFocusedChip() bool {
	switch f.FocusedChip {
	case FilterChipSearch:
		if f.SearchQuery != "" {
			f.SearchQuery = ""
			f.SearchInputActive = false
			return true
		}
	case FilterChipLabel:
		if len(f.LabelIDs) > 0 {
			f.LabelIDs = nil
			return true
		}
	case FilterChipPriority:
		if f.PriorityID != nil {
			f.PriorityID = nil
			return true
		}
	case FilterChipType:
		if f.TypeID != nil {
			f.TypeID = nil
			return true
		}
	case FilterChipAssignee:
		if f.AssigneeID != nil {
			f.AssigneeID = nil
			f.AssigneeName = ""
			return true
		}
	case FilterChipArchived:
		if f.ShowArchived {
			f.ShowArchived = false
			return true
		}
	case FilterChipClearAll:
		if f.HasAnyFilter() {
			f.ClearAll()
			return true
		}
	}
	return false
}

// AppendSearchChar appends a character to the search query.
// Returns true if the character was added, false if query is at max length.
func (f *FilterState) AppendSearchChar(c rune) bool {
	if runewidth.StringWidth(f.SearchQuery) >= maxSearchQueryLength {
		return false
	}
	f.SearchQuery += string(c)
	return true
}

// SearchBackspace removes the last character from the search query.
// Returns true if a character was removed, false if query was already empty.
func (f *FilterState) SearchBackspace() bool {
	if len(f.SearchQuery) == 0 {
		return false
	}
	runes := []rune(f.SearchQuery)
	f.SearchQuery = string(runes[:len(runes)-1])
	return true
}

// ClearSearch resets the search query and deactivates search input mode.
func (f *FilterState) ClearSearch() {
	f.SearchQuery = ""
	f.SearchInputActive = false
}
