package state

import "github.com/thenoetrevino/paso/internal/models"

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
)

// UsesLayers returns true if this mode uses layer-based rendering.
// Modes that use layers render a base view with modal overlays on top.
// Modes that don't use layers render full-screen content.
func (m Mode) UsesLayers() bool {
	switch m {
	case TicketFormMode,
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
		ProjectBranchConfirmMode:
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
		viewportSize:      1, // Default to 1, will be recalculated when width is set
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
// Column layout:
//   - Content width: 40 characters
//   - Padding: 2 characters (1 on each side)
//   - Border: 2 characters (1 on each side)
//   - Spacing: 2 characters (between columns)
//   - Total per column: 46 characters
//
// The calculation reserves 4 characters for margins and scroll indicators,
// and ensures at least 1 column is always visible.
// When the detail panel is visible, its width is subtracted from available space.
func (s *UIState) calculateViewportSize() {
	if s.width == 0 {
		s.viewportSize = 1
		return
	}

	const columnWidth = 46       // 40 content + 2 padding + 2 border + 2 spacing
	const reservedWidth = 4      // margins and scroll indicators
	const detailPanelWidth = 120 // max width of detail panel (prevents sawtooth resizing)

	availableWidth := s.width - reservedWidth

	// Reserve space for detail panel when screen is wide enough
	// We reserve the MAX panel width to prevent the panel from shrinking
	// when the terminal gets larger and gains another column
	if s.ShouldShowDetailPanel() {
		availableWidth -= detailPanelWidth
	}

	// Calculate how many columns fit, with minimum of 1
	s.viewportSize = max(1, availableWidth/columnWidth)
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
// can fit at least 5 columns (approximately 230 characters wide).
func (s *UIState) ShouldShowDetailPanel() bool {
	const (
		columnWidth        = 46 // 40 content + 2 padding + 2 border + 2 spacing
		minColumnsForPanel = 4  // require at least 4 columns worth of space
		reservedWidth      = 4  // margins and scroll indicators
	)
	minWidth := (minColumnsForPanel * columnWidth) + reservedWidth
	return s.width >= minWidth
}

// DetailPanelWidth calculates the width for the detail panel based on terminal width.
// The width is clamped between minPanelWidth and maxPanelWidth.
//
// The clamping logic handles edge cases gracefully:
//   - If availableWidth > maxPanelWidth: returns maxPanelWidth (prevents oversized panel)
//   - If availableWidth < minPanelWidth: returns minPanelWidth (ensures minimum usable width)
//   - If the terminal is very narrow and availableWidth becomes negative, the lower bound
//     check ensures we still return at least minPanelWidth for a usable panel.
func (s *UIState) DetailPanelWidth() int {
	const (
		columnWidth   = 46
		reservedWidth = 4
		minPanelWidth = 50
		maxPanelWidth = 120
	)

	// Calculate board width based on visible columns
	boardWidth := (s.viewportSize * columnWidth) + reservedWidth

	// Remaining space for panel (may be negative if terminal is very narrow)
	availableWidth := s.width - boardWidth

	// Clamp to min/max bounds
	if availableWidth < minPanelWidth {
		return minPanelWidth
	}
	if availableWidth > maxPanelWidth {
		return maxPanelWidth
	}
	return availableWidth
}
