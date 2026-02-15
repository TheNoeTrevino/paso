package state

// FilterChip represents the index of each chip in the filter bar.
type FilterChip int

const (
	FilterChipLabel    FilterChip = iota // Label filter chip
	FilterChipPriority                   // Priority filter chip
	FilterChipType                       // Type filter chip
	FilterChipAssignee                   // Assignee filter chip
	FilterChipClearAll                   // Clear All pseudo-chip
	FilterChipCount                      // Total number of chips (used for bounds)
)

// FilterState manages the filter bar selections and focus state.
// Each filter field is nullable — nil means "no filter applied".
type FilterState struct {
	PriorityID  *int       // Selected priority ID, nil = no filter
	TypeID      *int       // Selected type ID, nil = no filter
	AssigneeID  *int       // Selected assignee ID, nil = no filter (use -1 for "Unassigned" later)
	LabelIDs    []int      // Selected label IDs (OR logic), empty = no filter
	FocusedChip FilterChip // Currently focused chip in the filter bar
	IsActive    bool       // Whether the filter bar is focused for keyboard navigation
}

// NewFilterState creates a new FilterState with default values.
func NewFilterState() *FilterState {
	return &FilterState{
		FocusedChip: FilterChipLabel,
	}
}

// HasAnyFilter returns true if any filter is currently applied.
func (f *FilterState) HasAnyFilter() bool {
	return f.PriorityID != nil || f.TypeID != nil || f.AssigneeID != nil || len(f.LabelIDs) > 0
}

// ClearAll resets all filter selections to their default (unfiltered) state.
func (f *FilterState) ClearAll() {
	f.PriorityID = nil
	f.TypeID = nil
	f.AssigneeID = nil
	f.LabelIDs = nil
}

// SetPriority sets the priority filter. Pass nil to clear.
func (f *FilterState) SetPriority(id *int) {
	f.PriorityID = id
}

// ClearPriority clears the priority filter.
func (f *FilterState) ClearPriority() {
	f.PriorityID = nil
}

// SetType sets the type filter. Pass nil to clear.
func (f *FilterState) SetType(id *int) {
	f.TypeID = id
}

// ClearType clears the type filter.
func (f *FilterState) ClearType() {
	f.TypeID = nil
}

// SetAssignee sets the assignee filter. Pass nil to clear.
func (f *FilterState) SetAssignee(id *int) {
	f.AssigneeID = id
}

// ClearAssignee clears the assignee filter.
func (f *FilterState) ClearAssignee() {
	f.AssigneeID = nil
}

// SetLabels sets the label IDs filter. Pass nil or empty to clear.
func (f *FilterState) SetLabels(ids []int) {
	f.LabelIDs = ids
}

// ClearLabels clears the label filter.
func (f *FilterState) ClearLabels() {
	f.LabelIDs = nil
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
	case FilterChipLabel:
		if len(f.LabelIDs) > 0 {
			f.ClearLabels()
			return true
		}
	case FilterChipPriority:
		if f.PriorityID != nil {
			f.ClearPriority()
			return true
		}
	case FilterChipType:
		if f.TypeID != nil {
			f.ClearType()
			return true
		}
	case FilterChipAssignee:
		if f.AssigneeID != nil {
			f.ClearAssignee()
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
