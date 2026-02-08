package components

import "github.com/thenoetrevino/paso/internal/tui/layout"

const (
	TaskCardHeight       = 5 // TaskCardHeight is the fixed height of the task card
	columnBorderOverhead = 3 // top border + bottom padding + bottom border
	headerLines          = 1 // column name and count
	topIndicatorLines    = 1 // empty line or "▲ more above"

	// taskCardLeadingSpace is the single-character " " prefix prepended
	// in renderTaskSummaryTitle, renderTaskCardLabels, and renderTaskSummaryMetadata.
	taskCardLeadingSpace = 1
	// taskCardMeasurementSlack is extra margin to prevent lipgloss from
	// word-wrapping when its internal width measurement disagrees with
	// runewidth by a character or two (observed with styled/colored text).
	taskCardMeasurementSlack = 2
	// taskCardSafetyBuffer is subtracted from the card content width to
	// get the max width available for segment text.
	taskCardSafetyBuffer = taskCardLeadingSpace + taskCardMeasurementSlack

	// Picker footer/help text strings
	PickerFooterSelectConfirm = "Enter: select  Esc: cancel"       // Used by: Color, Priority, Type, Relation Type pickers
	PickerFooterToggleCreate  = "Enter: toggle/create  Esc: close" // Used by: Label picker
	PickerFooterToggle        = "Enter: toggle  Esc: close"        // Used by: Task picker
	PickerFooterConfirm       = "Enter: confirm  Esc: cancel"      // Used by: Status picker
)

// Re-export layout constants for backward compatibility
const (
	ScrollIndicatorChrome  = layout.Chrome
	ColumnMinContentWidth  = layout.ColumnMinContentWidth
	ColumnMaxContentWidth  = layout.ColumnMaxContentWidth
	ColumnOverhead         = layout.ColumnOverhead
	ColumnMinTotalWidth    = layout.ColumnMinTotalWidth
	DetailPanelPercent     = layout.DetailPanelPercent
	DetailPanelMinWidth    = layout.DetailPanelMinWidth
	DetailPanelMaxWidth    = layout.DetailPanelMaxWidth
	TaskCardBorderOverhead = layout.TaskCardBorderOverhead
)
