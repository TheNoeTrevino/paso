package styles

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/config/colors"
)

// Detail represents a key-value pair for detailed output
type Detail struct {
	Key   string
	Value string
}

// RenderSuccess renders a minimal success message with green checkmark
// Example: "✓ Task created successfully"
func RenderSuccess(message string, colorScheme colors.ColorScheme) string {
	checkmark := renderCheckmark(colorScheme)
	return fmt.Sprintf("%s %s\n", checkmark, message)
}

// RenderSuccessWithDetails renders success message with key-value details
// Example:
//
//	✓ Task created successfully
//	  ID: 239
//	  Title: Fix bug
//	  Project: tabular output
func RenderSuccessWithDetails(message string, details []Detail, colorScheme colors.ColorScheme) string {
	var out strings.Builder

	// Render header with checkmark
	checkmark := renderCheckmark(colorScheme)
	fmt.Fprintf(&out, "%s %s\n", checkmark, message)

	// Render details with indentation
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorScheme.Normal))

	for _, detail := range details {
		line := fmt.Sprintf("  %s: %s", detail.Key, detail.Value)
		out.WriteString(normalStyle.Render(line))
		out.WriteString("\n")
	}

	return out.String()
}

// renderCheckmark creates a green checkmark using the Create color
func renderCheckmark(colorScheme colors.ColorScheme) string {
	checkmarkStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorScheme.Create))

	return checkmarkStyle.Render("✓")
}

// RenderError renders an error message with red X mark
// Example: "✗ Error: invalid ID 'abc': must be a number"
func RenderError(message string, colorScheme colors.ColorScheme) string {
	xMark := renderErrorMark(colorScheme)
	return fmt.Sprintf("%s Error: %s", xMark, message)
}

// renderErrorMark creates a red X mark using the Delete color
func renderErrorMark(colorScheme colors.ColorScheme) string {
	errorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorScheme.Delete))

	return errorStyle.Render("✗")
}

// TruncateString truncates a string to maxLen characters and adds "..." if truncated
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
