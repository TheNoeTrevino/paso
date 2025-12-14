package state

// ViewMode represents the current view mode of the kanban board.
// Users can toggle between Kanban (column-based) and List (table-based) views.
type ViewMode int

const (
	KanbanView ViewMode = iota // Default column-based kanban view
	ListView                   // Table-based list view
)

// SortField represents the field to sort by in list view.
type SortField int

const (
	SortNone     SortField = iota // No sorting applied
	SortByTitle                   // Sort by task title
	SortByStatus                  // Sort by column/status
)

// SortOrder represents the sort direction.
type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

// ListViewState manages the list view state.
// This includes view mode toggle, row selection, scrolling, and sorting configuration.
type ListViewState struct {
	// viewMode is the current view mode (kanban or list)
	viewMode ViewMode

	// selectedRow is the index of the currently selected row in list view
	selectedRow int

	// scrollOffset is the vertical scroll offset for list view
	scrollOffset int

	sortField SortField
	sortOrder SortOrder
}

// NewListViewState creates a new ListViewState with default values.
func NewListViewState() *ListViewState {
	return &ListViewState{
		viewMode:     KanbanView,
		selectedRow:  0,
		scrollOffset: 0,
		sortField:    SortNone,
		sortOrder:    SortAsc,
	}
}

// ViewMode returns the current view mode.
func (s *ListViewState) ViewMode() ViewMode {
	return s.viewMode
}

// SetViewMode updates the current view mode.
func (s *ListViewState) SetViewMode(mode ViewMode) {
	s.viewMode = mode
}

// ToggleView toggles between kanban and list views.
func (s *ListViewState) ToggleView() {
	if s.viewMode == KanbanView {
		s.viewMode = ListView
	} else {
		s.viewMode = KanbanView
	}
}

// IsListView returns true if currently in list view mode.
func (s *ListViewState) IsListView() bool {
	return s.viewMode == ListView
}

// SelectedRow returns the index of the currently selected row.
func (s *ListViewState) SelectedRow() int {
	return s.selectedRow
}

// SetSelectedRow updates the selected row index.
func (s *ListViewState) SetSelectedRow(row int) {
	s.selectedRow = row
}

// ScrollOffset returns the current scroll offset.
func (s *ListViewState) ScrollOffset() int {
	return s.scrollOffset
}

// SetScrollOffset updates the scroll offset.
func (s *ListViewState) SetScrollOffset(offset int) {
	s.scrollOffset = offset
}

// MoveUp moves the selection up one row if possible.
func (s *ListViewState) MoveUp() {
	if s.selectedRow > 0 {
		s.selectedRow--
	}
}

// MoveDown moves the selection down one row if possible.
//
// Parameters:
//   - maxRows: the maximum number of rows available
func (s *ListViewState) MoveDown(maxRows int) {
	if maxRows > 0 && s.selectedRow < maxRows-1 {
		s.selectedRow++
	}
}

// ResetSelection resets the row selection and scroll offset to zero.
// This is typically called when switching projects or changing views.
func (s *ListViewState) ResetSelection() {
	s.selectedRow = 0
	s.scrollOffset = 0
}

// SortField returns the current sort field.
func (s *ListViewState) SortField() SortField {
	return s.sortField
}

// SortOrder returns the current sort order.
func (s *ListViewState) SortOrder() SortOrder {
	return s.sortOrder
}

// CycleSort cycles through the sort configurations.
// Order: None -> Title Asc -> Title Desc -> Status Asc -> Status Desc -> None
func (s *ListViewState) CycleSort() {
	switch {
	case s.sortField == SortNone:
		// None -> Title Asc
		s.sortField = SortByTitle
		s.sortOrder = SortAsc
	case s.sortField == SortByTitle && s.sortOrder == SortAsc:
		// Title Asc -> Title Desc
		s.sortOrder = SortDesc
	case s.sortField == SortByTitle && s.sortOrder == SortDesc:
		// Title Desc -> Status Asc
		s.sortField = SortByStatus
		s.sortOrder = SortAsc
	case s.sortField == SortByStatus && s.sortOrder == SortAsc:
		// Status Asc -> Status Desc
		s.sortOrder = SortDesc
	case s.sortField == SortByStatus && s.sortOrder == SortDesc:
		// Status Desc -> None
		s.sortField = SortNone
		s.sortOrder = SortAsc
	default:
		// Fallback to None
		s.sortField = SortNone
		s.sortOrder = SortAsc
	}
}
