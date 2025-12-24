package notifications

import (
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// Render renders a notification banner based on severity level
func Render(severity Severity, message string) string {
	style := severity.style()

	// Calculate max width needed
	headerText := style.icon + " " + style.title
	maxWidth := max(lipgloss.Width(headerText), lipgloss.Width(message))

	// Create header with icon and title
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(style.foreground)).
		Bold(true).
		Width(maxWidth)

	if severity == Info {
		headerStyle = headerStyle.Background(lipgloss.Color(style.background))
	}

	header := headerStyle.Render(headerText)

	// Create message content
	messageContent := lipgloss.NewStyle().
		Foreground(lipgloss.Color(style.foreground)).
		Width(maxWidth).
		Render(message)

	// Combine header and message vertically
	content := lipgloss.JoinVertical(lipgloss.Left, header, messageContent)

	// Apply border and background
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(style.borderForeground)).
		Background(lipgloss.Color(style.background)).
		Padding(0, 1).
		Render(content)
}

// RenderFromState renders a notification banner from a state.Notification
func RenderFromState(n state.Notification) string {
	switch n.Level {
	case state.LevelInfo:
		return Render(Info, n.Message)
	case state.LevelWarning:
		return Render(Warning, n.Message)
	case state.LevelError:
		return Render(Error, n.Message)
	default:
		return Render(Info, n.Message)
	}
}

// RenderInline renders a compact inline notification (for tab bar)
func RenderInline(severity Severity, message string) string {
	style := severity.style()

	// Icon + message on single line
	content := style.icon + " " + message

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(style.foreground)).
		Background(lipgloss.Color(style.background)).
		Padding(0, 1).
		Render(content)
}

// RenderInlineFromState renders a compact inline notification from state
func RenderInlineFromState(n state.Notification) string {
	switch n.Level {
	case state.LevelInfo:
		return RenderInline(Info, n.Message)
	case state.LevelWarning:
		return RenderInline(Warning, n.Message)
	case state.LevelError:
		return RenderInline(Error, n.Message)
	default:
		return RenderInline(Info, n.Message)
	}
}
