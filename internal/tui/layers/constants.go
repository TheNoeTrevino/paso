package layers

const (
	PickerDefaultWidthDivisor = 2

	PickerChromeHeightNoFilter   = 4 // title, footer, padding, border
	PickerChromeHeightWithFilter = 6 // above + filter line + spacing

	PickerMaxVisibleItems = 15

	PickerMinHeight          = 10
	PickerMaxHeightDivisor   = 4 // 3/4 = 75% of screen height
	PickerMaxHeightNumerator = 3

	PickerBorderPaddingWidth  = 8 // left + right
	PickerBorderPaddingHeight = 4 // top + bottom

	PickerDefaultMinWidth  = 40
	PickerDefaultMaxWidth  = 60
	PickerDefaultMinHeight = 10

	PickerTaskMinWidth  = 50
	PickerTaskMaxWidth  = 70
	PickerTaskMinHeight = 12

	PickerPriorityWidth  = 40
	PickerPriorityHeight = 12 // 5 options + chrome

	PickerTypeWidth  = 40
	PickerTypeHeight = 9 // 2 options + chrome

	PickerRelationTypeWidth  = 45 // wider for longer option descriptions
	PickerRelationTypeHeight = 11 // 3 options + chrome. WARN: might need to be dynamic

	PickerStatusWidth           = 40 // height is dynamic based on column count
	PickerStatusChromeHeight    = 6  // title, spacing, footer
	PickerColorDefaultItemCount = 10
)
