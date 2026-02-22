package standup

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

// FormatListQuiet returns a slice of log IDs as strings
func FormatListQuiet(logs []models.StandupLog) []string {
	ids := make([]string, len(logs))
	for i, l := range logs {
		ids[i] = fmt.Sprintf("%d", l.ID)
	}
	return ids
}

// FormatListJSON generates the JSON output structure
func FormatListJSON(logs []models.StandupLog) map[string]any {
	items := make([]map[string]any, len(logs))
	for i, l := range logs {
		items[i] = map[string]any{
			"id":         l.ID,
			"project_id": l.ProjectID,
			"content":    l.Content,
			"created_at": l.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return map[string]any{
		"success": true,
		"logs":    items,
	}
}

// FormatListHuman renders standup logs as a human-readable timeline
func FormatListHuman(logs []models.StandupLog, colorScheme colors.ColorScheme) string {
	var out strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorScheme.Normal))
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorScheme.Subtle))
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorScheme.Normal))

	fmt.Fprintf(&out, "Found %d standup log(s):\n\n", len(logs))

	for _, l := range logs {
		timestamp := l.CreatedAt.Format("Jan 02, 2006 3:04 PM")
		idStr := fmt.Sprintf("#%d", l.ID)

		out.WriteString(headerStyle.Render(idStr))
		out.WriteString(" ")
		out.WriteString(dimStyle.Render(timestamp))
		out.WriteString("\n")

		// Indent content lines
		for line := range strings.SplitSeq(l.Content, "\n") {
			out.WriteString("  ")
			out.WriteString(contentStyle.Render(line))
			out.WriteString("\n")
		}
		out.WriteString("\n")
	}

	return out.String()
}
