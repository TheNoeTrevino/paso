// Package layout provides shared layout constants for the TUI.
// This package exists to avoid import cycles between state and components.
package layout

const (
	// Chrome is the fixed width for scroll indicators ("< " + " >")
	Chrome = 4

	// Column dimensions
	ColumnMinContentWidth = 35                                     // Minimum usable column content width
	ColumnMaxContentWidth = 60                                     // Maximum column content width
	ColumnOverhead        = 4                                      // Border + padding (2 left + 2 right)
	ColumnMinTotalWidth   = ColumnMinContentWidth + ColumnOverhead // Minimum total column width (39)

	// Detail panel dimensions
	DetailPanelPercent  = 35  // Detail panel takes ~35% of available content width
	DetailPanelMinWidth = 50  // Minimum panel width for usability
	DetailPanelMaxWidth = 120 // Maximum panel width

	// Task card layout
	TaskCardBorderOverhead = 4 // Task card border (2 left + 2 right)

	// Viewport
	MinViewportColumns = 3 // Minimum visible columns
)
