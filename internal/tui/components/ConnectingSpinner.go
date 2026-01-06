package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// spinnerFrames contains the moon phase icons for the spinner animation
// The animation cycles forward through all frames continuously
var spinnerFrames = []string{
	"󰋙", "󰫃", "󰫄", "󰫅", "󰫆", "󰫇", "󰫈", "󰫇", "󰫆", "󰫅", "󰫄", "󰫃",
}

// GetSpinnerFrameCount returns the total number of spinner animation frames
// This allows external state management to validate frame indices
func GetSpinnerFrameCount() int {
	return len(spinnerFrames)
}

// RenderConnectingSpinner renders the connecting spinner with the current animation frame
// Parameters:
//   - dbName: the database name being connected to
//   - frame: the current animation frame index (0 to len(spinnerFrames)-1)
//
// Returns a styled string showing "connecting to {dbName} {spinner}"
func RenderConnectingSpinner(dbName string, frame int) string {
	// Wrap frame to valid range for safety
	frame = frame % len(spinnerFrames)

	spinnerIcon := spinnerFrames[frame]
	text := fmt.Sprintf("connecting to %s %s", dbName, spinnerIcon)

	// Use bold + highlight color for visibility
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	return style.Render(text)
}
