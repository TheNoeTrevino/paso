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
