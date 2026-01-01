// Package layers provides utility functions for creating and managing UI layers
package layers

import "charm.land/lipgloss/v2"

// CreateCenteredLayer creates a layer positioned at the center of the screen.
// Typically called with ui.ScreenWidth() and ui.ScreenHeight() as dimensions.
//
// Parameters:
//   - content: the rendered content to center
//   - screenWidth: the width of the screen
//   - screenHeight: the height of the screen
//
// Returns:
//   - A layer positioned at the center of the screen, or nil if content is empty
func CreateCenteredLayer(content string, screenWidth int, screenHeight int) *lipgloss.Layer {
	if content == "" {
		return nil
	}

	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)

	x := (screenWidth - contentWidth) / 2
	y := (screenHeight - contentHeight) / 2

	x = max(x, 0)
	y = max(y, 0)

	return lipgloss.NewLayer(content).X(x).Y(y)
}

// CalculatePickerDimensions determines optimal picker size based on content
// Returns (width, height) suitable for dynamic picker sizing
func CalculatePickerDimensions(
	itemCount int,
	hasFilter bool,
	screenWidth int,
	screenHeight int,
	minWidth int,
	maxWidth int,
) (int, int) {
	width := min(max(screenWidth/PickerDefaultWidthDivisor, minWidth), maxWidth)

	chromeHeight := PickerChromeHeightNoFilter
	if hasFilter {
		chromeHeight = PickerChromeHeightWithFilter
	}

	visibleItems := min(itemCount, PickerMaxVisibleItems)

	height := chromeHeight + visibleItems

	maxHeight := screenHeight * PickerMaxHeightNumerator / PickerMaxHeightDivisor

	height = max(height, PickerMinHeight)
	height = min(height, maxHeight)

	return width, height
}
