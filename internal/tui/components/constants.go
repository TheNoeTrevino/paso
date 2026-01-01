package components

const (
	TaskCardHeight        = 5  // TaskCardHeight is the fixed height of the task card
	taskTitleMaxLength    = 30 // Maximum display length for task title before truncation
	taskTitlePaddedLength = 33 // Total padded length including ellipsis space
	columnBorderOverhead  = 3  // top border + bottom padding + bottom border
	headerLines           = 1  // column name and count
	topIndicatorLines     = 1  // empty line or "▲ more above"

	// Picker footer/help text strings
	PickerFooterSelectConfirm = "Enter: select  Esc: cancel"       // Used by: Color, Priority, Type, Relation Type pickers
	PickerFooterToggleCreate  = "Enter: toggle/create  Esc: close" // Used by: Label picker
	PickerFooterToggle        = "Enter: toggle  Esc: close"        // Used by: Task picker
	PickerFooterConfirm       = "Enter: confirm  Esc: cancel"      // Used by: Status picker
)
