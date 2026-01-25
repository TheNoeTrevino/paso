package components

const (
	TaskCardHeight       = 5 // TaskCardHeight is the fixed height of the task card
	columnBorderOverhead = 3 // top border + bottom padding + bottom border
	headerLines          = 1 // column name and count
	topIndicatorLines    = 1 // empty line or "▲ more above"

	// Responsive layout constants
	ScrollIndicatorChrome = 4  // "◀ " + " ▶" - fixed chrome for scroll indicators
	ColumnMinContentWidth = 35 // Minimum usable column content width
	ColumnMaxContentWidth = 60 // Maximum column content width (prevents oversized columns)
	ColumnOverhead        = 4  // 2 border + 2 padding (left+right)

	// Detail panel dimensions
	DetailPanelPercent  = 35  // Detail panel takes ~35% of available content width
	DetailPanelMinWidth = 50  // Minimum panel width for usability
	DetailPanelMaxWidth = 120 // Maximum panel width

	// Task card layout
	TaskCardBorderOverhead = 4 // Task card border (2 left + 2 right)
	TaskCardMinPadding     = 3 // Minimum padding for task content

	// Picker footer/help text strings
	PickerFooterSelectConfirm = "Enter: select  Esc: cancel"       // Used by: Color, Priority, Type, Relation Type pickers
	PickerFooterToggleCreate  = "Enter: toggle/create  Esc: close" // Used by: Label picker
	PickerFooterToggle        = "Enter: toggle  Esc: close"        // Used by: Task picker
	PickerFooterConfirm       = "Enter: confirm  Esc: cancel"      // Used by: Status picker
)

// TaskTitleMaxLength calculates the max title length based on available card width
func TaskTitleMaxLength(cardWidth int) int {
	// Leave room for padding and potential ellipsis
	maxLen := cardWidth - TaskCardMinPadding
	if maxLen < 10 {
		return 10 // Absolute minimum
	}
	return maxLen
}

// TaskLabelsMaxLength calculates the max labels length based on available card width
func TaskLabelsMaxLength(cardWidth int) int {
	// Labels line has similar constraints to title
	maxLen := cardWidth - TaskCardMinPadding
	if maxLen < 10 {
		return 10
	}
	return maxLen
}
