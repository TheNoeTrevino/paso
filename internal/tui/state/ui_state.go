package state

import (
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/layout"
)

// Mode represents the current interaction mode of the TUI.
// Each mode determines which keyboard shortcuts are active and what UI is displayed.
type Mode int

const (
	NormalMode                 Mode = iota // Default navigation mode
	DeleteConfirmMode                      // Confirming task deletion
	DiscardConfirmMode                     // Confirming discard of form/input changes
	AddColumnFormMode                      // Creating a new column (huh form)
	EditColumnFormMode                     // Renaming an existing column (huh form)
	DeleteColumnConfirmMode                // Confirming column deletion
	HelpMode                               // Displaying help screen
	TicketFormMode                         // Full task form with huh
	TicketFormLoadingMode                  // Loading task data for edit form
	ProjectFormMode                        // Creating a new project with huh
	EditProjectFormMode                    // Editing an existing project (huh form)
	ProjectFormLoadingMode                 // Loading git data for project form
	LabelManagementMode                    // Managing labels (create/edit/delete)
	LabelAssignMode                        // Quick label assignment to task
	LabelPickerMode                        // GitHub-style label picker popup
	ParentPickerMode                       // Parent issue picker popup
	ChildPickerMode                        // Child issue picker popup
	PriorityPickerMode                     // Priority picker popup
	TypePickerMode                         // Type picker popup
	RelationTypePickerMode                 // Relation type picker popup
	CommentEditMode                        // Comment editing mode (list navigation)
	CommentFormMode                        // Comment form (huh form for creating/editing individual comment)
	CommentsViewMode                       // Dedicated comments list view for a task
	SearchMode                             // Vim-style search mode (/)
	StatusPickerMode                       // Status picker popup for list view
	TaskFormHelpMode                       // Help screen for task form shortcuts
	DatabaseSelectMode                     // Selecting from existing database connections
	DatabaseCreateMode                     // Creating a new database connection (huh form)
	DatabaseConnectConfirmMode             // Confirming connection to newly created database
	DatabaseDeleteConfirmMode              // Confirming deletion of a database connection
	ProjectBranchConfirmMode               // Confirming git branch association for project
	DeleteProjectConfirmMode               // Confirming project deletion
)

// UsesLayers returns true if this mode uses layer-based rendering.
// Modes that use layers render a base view with modal overlays on top.
// Modes that don't use layers render full-screen content.
func (m Mode) UsesLayers() bool {
	switch m {
	case TicketFormMode,
		TicketFormLoadingMode,
		ProjectFormMode,
		EditProjectFormMode,
		ProjectFormLoadingMode,
		AddColumnFormMode,
		EditColumnFormMode,
		CommentFormMode,
		CommentsViewMode,
		HelpMode,
		TaskFormHelpMode,
		LabelPickerMode,
		ParentPickerMode,
		ChildPickerMode,
		PriorityPickerMode,
		TypePickerMode,
		RelationTypePickerMode,
		StatusPickerMode,
		DiscardConfirmMode,
		NormalMode,
		SearchMode,
		DatabaseSelectMode,
		DatabaseCreateMode,
		DatabaseConnectConfirmMode,
		DatabaseDeleteConfirmMode,
		ProjectBranchConfirmMode,
		DeleteProjectConfirmMode:
		return true
	default:
		return false
	}
}

// DiscardContext tracks information for discard confirmation dialogs.
// It stores the mode to return to if the user cancels, and a context-specific message.
type DiscardContext struct {
	SourceMode Mode   // The mode to return to if user cancels discard (N/ESC)
	Message    string // Context-specific message (e.g., "Discard task?", "Discard project?")
}

// ProjectBranchContext tracks information for project git branch association confirmation
type ProjectBranchContext struct {
	ProjectName        string
	ProjectDescription string
	GitBranch          string
	ExistingProject    *models.Project
}

// UIState manages the user interface state.
// This includes navigation (column/task selection), viewport scrolling,
// terminal dimensions, and the current interaction mode.
type UIState struct {
	SelectedColumn int
	SelectedTask   int

	// width is kept private because SetWidth() has side effects (calls calculateViewportSize)
	width int

	Height         int
	Mode           Mode
	ViewportOffset int

	// viewportSize is the number of columns that fit on the screen (calculated)
	viewportSize int

	// taskScrollOffsets tracks the vertical scroll offset for each column
	// Key: columnID, Value: scroll offset (index of first visible task)
	taskScrollOffsets map[int]int

	DiscardContext       *DiscardContext
	ProjectBranchContext *ProjectBranchContext
}

// NewUIState creates a new UIState with default values.
func NewUIState() *UIState {
	return &UIState{
		SelectedColumn:    0,
		SelectedTask:      0,
		width:             0,
		Height:            0,
		Mode:              NormalMode,
		ViewportOffset:    0,
		viewportSize:      layout.MinViewportColumns, // Default to minimum, recalculated when width is set
		taskScrollOffsets: make(map[int]int),
	}
}

// Width returns the current terminal width.
func (s *UIState) Width() int {
	return s.width
}

// SetWidth updates the terminal width and recalculates viewport size.
func (s *UIState) SetWidth(width int) {
	s.width = width
	s.calculateViewportSize()
}

// ContentHeight returns the available height for the main content area.
// This is terminal height minus tab bar and status bar, ensuring a minimum of 5.
func (s *UIState) ContentHeight() int {
	const tabBarHeight = 3    // tabs + gap line
	const statusBarHeight = 2 // status bar + gap line
	return max(s.Height-tabBarHeight-statusBarHeight, 5)
}

// ViewportSize returns the number of columns that fit on screen.
func (s *UIState) ViewportSize() int {
	return s.viewportSize
}

// calculateViewportSize calculates how many columns can fit in the terminal width.
//
// This uses the responsive layout system where columns expand to fill available space.
// The viewport size is determined by how many columns can fit at minimum width,
// ensuring we always show at least 3 columns when possible.
func (s *UIState) calculateViewportSize() {
	if s.width == 0 {
		s.viewportSize = layout.MinViewportColumns
		return
	}

	availableWidth := s.width - layout.Chrome

	if s.ShouldShowDetailPanel() {
		availableWidth -= s.DetailPanelWidth()
	}

	s.viewportSize = max(layout.MinViewportColumns, availableWidth/layout.ColumnMinTotalWidth)
}

// AdjustViewportAfterColumnRemoval adjusts the viewport offset after a column is removed.
// This ensures the viewport stays within valid bounds and the selection remains visible.
//
// Parameters:
//   - selectedColumn: the current selected column index
//   - columnsLen: the total number of columns after removal
func (s *UIState) AdjustViewportAfterColumnRemoval(selectedColumn, columnsLen int) {
	if columnsLen == 0 {
		s.ViewportOffset = 0
		return
	}
	if selectedColumn < s.ViewportOffset {
		s.ViewportOffset = selectedColumn
	}
	if s.ViewportOffset+s.viewportSize > columnsLen {
		s.ViewportOffset = max(0, columnsLen-s.viewportSize)
	}
}

// ScrollViewportLeft scrolls the viewport one column to the left.
// Returns true if scrolling occurred, false if already at leftmost position.
func (s *UIState) ScrollViewportLeft() bool {
	if s.ViewportOffset > 0 {
		s.ViewportOffset--
		return true
	}
	return false
}

// ScrollViewportRight scrolls the viewport one column to the right.
// Returns true if scrolling occurred, false if already at rightmost position.
//
// Parameters:
//   - columnsLen: the total number of columns
func (s *UIState) ScrollViewportRight(columnsLen int) bool {
	if s.ViewportOffset+s.viewportSize < columnsLen {
		s.ViewportOffset++
		return true
	}
	return false
}

// EnsureSelectionVisible adjusts the viewport to ensure the selected column is visible.
// This should be called after navigation or when the selection changes.
func (s *UIState) EnsureSelectionVisible(selectedColumn int) {
	// If selection is off-screen to the left, scroll left
	if selectedColumn < s.ViewportOffset {
		s.ViewportOffset = selectedColumn
	}

	// If selection is off-screen to the right, scroll right
	if selectedColumn >= s.ViewportOffset+s.viewportSize {
		s.ViewportOffset = selectedColumn - s.viewportSize + 1
	}
}

// ResetSelection resets both column and task selection to zero.
// This is typically called when switching projects or clearing state.
func (s *UIState) ResetSelection() {
	s.SelectedColumn = 0
	s.SelectedTask = 0
	s.ViewportOffset = 0
}

// TaskScrollOffset returns the vertical scroll offset for a given column.
// Returns 0 if the column has no scroll offset set.
func (s *UIState) TaskScrollOffset(columnID int) int {
	if offset, ok := s.taskScrollOffsets[columnID]; ok {
		return offset
	}
	return 0
}

// SetTaskScrollOffset updates the vertical scroll offset for a given column.
func (s *UIState) SetTaskScrollOffset(columnID int, offset int) {
	s.taskScrollOffsets[columnID] = max(0, offset)
}

// ScrollTasksUp moves the scroll offset up (decreases it) for a column.
// Returns true if scrolling occurred, false if already at top.
func (s *UIState) ScrollTasksUp(columnID int) bool {
	offset := s.TaskScrollOffset(columnID)
	if offset > 0 {
		s.taskScrollOffsets[columnID] = offset - 1
		return true
	}
	return false
}

// ScrollTasksDown moves the scroll offset down (increases it) for a column.
// Returns true if scrolling occurred, false if already at bottom.
//
// Parameters:
//   - columnID: the column to scroll
//   - taskCount: total number of tasks in the column
//   - visibleCount: number of tasks that can be displayed at once
func (s *UIState) ScrollTasksDown(columnID int, taskCount int, visibleCount int) bool {
	offset := s.TaskScrollOffset(columnID)
	maxOffset := max(0, taskCount-visibleCount)
	if offset < maxOffset {
		s.taskScrollOffsets[columnID] = offset + 1
		return true
	}
	return false
}

// EnsureTaskVisible adjusts the scroll offset to ensure the selected task is visible.
// This should be called after task navigation within a column.
//
// Parameters:
//   - columnID: the column containing the task
//   - selectedTaskIdx: index of the selected task within the column
//   - visibleCount: number of tasks that can be displayed at once
func (s *UIState) EnsureTaskVisible(columnID int, selectedTaskIdx int, visibleCount int) {
	offset := s.TaskScrollOffset(columnID)

	// If selection is above visible area, scroll up
	if selectedTaskIdx < offset {
		s.taskScrollOffsets[columnID] = selectedTaskIdx
	}

	// If selection is below visible area, scroll down
	if selectedTaskIdx >= offset+visibleCount {
		s.taskScrollOffsets[columnID] = selectedTaskIdx - visibleCount + 1
	}
}

// ShouldShowDetailPanel returns true if the screen is wide enough to display
// the detail panel alongside the kanban columns. Panel is shown when the terminal
// can fit at least 3 columns at minimum width plus the minimum panel width.
func (s *UIState) ShouldShowDetailPanel() bool {
	minWidth := (layout.MinViewportColumns * layout.ColumnMinTotalWidth) + layout.Chrome + layout.DetailPanelMinWidth
	return s.width >= minWidth
}

// DetailPanelWidth calculates the width for the detail panel based on terminal width.
// Uses percentage-based sizing (35% of available content area) clamped to min/max bounds.
func (s *UIState) DetailPanelWidth() int {
	availableContent := s.width - layout.Chrome
	panelWidth := (availableContent * layout.DetailPanelPercent) / 100

	if panelWidth < layout.DetailPanelMinWidth {
		return layout.DetailPanelMinWidth
	}
	if panelWidth > layout.DetailPanelMaxWidth {
		return layout.DetailPanelMaxWidth
	}
	return panelWidth
}

// ColumnsAreaWidth returns the total width available for the columns area.
// This is the terminal width minus chrome and detail panel (if shown).
func (s *UIState) ColumnsAreaWidth() int {
	availableWidth := s.width - layout.Chrome

	if s.ShouldShowDetailPanel() {
		availableWidth -= s.DetailPanelWidth()
	}

	return max(availableWidth, 0)
}

// ColumnContentWidth calculates the content width for each column based on
// available space and the number of visible columns.
// Returns the inner content width (excluding border and padding).
func (s *UIState) ColumnContentWidth(visibleColumns int) int {
	if visibleColumns <= 0 {
		visibleColumns = 1
	}

	columnsArea := s.ColumnsAreaWidth()
	columnTotalWidth := columnsArea / visibleColumns
	contentWidth := columnTotalWidth - layout.ColumnOverhead

	if contentWidth < layout.ColumnMinContentWidth {
		return layout.ColumnMinContentWidth
	}
	if contentWidth > layout.ColumnMaxContentWidth {
		return layout.ColumnMaxContentWidth
	}
	return contentWidth
}

// TaskCardWidth calculates the width for task cards based on column content width.
// Task cards have their own border which reduces available content space.
func (s *UIState) TaskCardWidth(visibleColumns int) int {
	return s.ColumnContentWidth(visibleColumns) - layout.TaskCardBorderOverhead
}
