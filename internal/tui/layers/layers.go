package layers

import "charm.land/lipgloss/v2"

// CreateCenteredLayer creates a layer positioned at the center of the screen.
// This replaces the lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content) pattern.
//
// Parameters:
//   - content: the rendered content to center
//   - screenWidth: the width of the screen
//   - screenHeight: the height of the screen
//
// Returns:
//   - A layer positioned at the center of the screen, or nil if content is empty
func CreateCenteredLayer(content string, screenWidth, screenHeight int) *lipgloss.Layer {
	if content == "" {
		return nil
	}

	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)

	// Calculate center position
	x := (screenWidth - contentWidth) / 2
	y := (screenHeight - contentHeight) / 2

	// Ensure we don't go off screen
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

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
	// Width: Use percentage of screen with min/max bounds
	width := screenWidth / 2 // 50% default
	if width < minWidth {
		width = minWidth
	}
	if width > maxWidth {
		width = maxWidth
	}

	// Height: Based on content with chrome overhead
	// Chrome: title (1) + filter (2 if present) + footer (2) + padding (2) + border (2) = 7-9 lines
	chromeHeight := 7
	if hasFilter {
		chromeHeight = 9
	}

	// Calculate height for items (max 15 visible items)
	visibleItems := itemCount
	if visibleItems > 15 {
		visibleItems = 15
	}

	height := chromeHeight + visibleItems

	// Apply min/max bounds
	minHeight := 10
	maxHeight := screenHeight * 3 / 4
	if height < minHeight {
		height = minHeight
	}
	if height > maxHeight {
		height = maxHeight
	}

	return width, height
}
