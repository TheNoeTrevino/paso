package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// previewContentMaxLines is the maximum number of content lines shown per
// activity item in the compact preview (excluding the header line).
const previewContentMaxLines = 2

// activityPreviewLines is the total number of lines a single activity item
// occupies in the compact preview: 1 header + up to previewContentMaxLines
// content lines + 1 blank separator line.
const activityPreviewLines = 4

// RenderActivityBadge renders a compact colored badge identifying the
// activity type (Comment, Event, or Unknown). The badge uses a colored
// background with white text and single-space horizontal padding.
func RenderActivityBadge(activityType models.ActivityType) string {
	var bgColor, fgColor, text string

	switch activityType {
	case models.ActivityTypeEvent:
		bgColor = theme.Subtle
		fgColor = "#ffffff"
		text = "Event"
	case models.ActivityTypeComment:
		bgColor = theme.Create
		fgColor = "#ffffff"
		text = "Comment"
	default:
		bgColor = theme.Subtle
		fgColor = "#ffffff"
		text = "Unknown"
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color(bgColor)).
		Foreground(lipgloss.Color(fgColor)).
		Padding(0, 1).
		Render(text)
}

// RenderActivityPreviewItem renders a single activity item in the compact
// preview format used by the detail panel sidebar and task form:
//
//	[Comment] 󰀄 alice  Mar 15 09:30
//	This is the comment content, word wrapped
//	and truncated to 2 lines...
//
// Content is word-wrapped to max(width-4, 20) columns and truncated to
// previewContentMaxLines, with "..." appended to the final line when
// truncation occurs. An empty author is rendered as "Unknown".
func RenderActivityPreviewItem(item models.ActivityItem, width int) string {
	contentWidth := max(width-4, 20)

	lines := strings.Split(wordwrap.String(item.Content, contentWidth), "\n")
	if len(lines) > previewContentMaxLines {
		lines = lines[:previewContentMaxLines]
		lines[previewContentMaxLines-1] = lines[previewContentMaxLines-1] + "..."
	}

	badge := RenderActivityBadge(item.Type)

	timestamp := item.CreatedAt.Format("Jan 2 15:04")
	authorIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render(" 󰀄 ")
	author := item.Author
	if author == "" {
		author = "Unknown"
	}
	dateIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("  ")

	activityHeader := fmt.Sprintf("%s%s%s%s%s", badge, authorIcon, author, dateIcon, timestamp)
	headerLine := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render(activityHeader)

	contentLines := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal)).Render(strings.Join(lines, "\n"))

	return headerLine + "\n" + contentLines
}

// RenderActivityPreviews renders a height-aware list of activity items in
// the compact preview format. Items are assumed to be pre-sorted (typically
// newest first) and non-empty (callers are expected to pre-guard). At most
// availableHeight/activityPreviewLines items are rendered (always at least
// one), joined by a blank line.
func RenderActivityPreviews(items []models.ActivityItem, width, availableHeight int) string {
	maxItems := max(availableHeight/activityPreviewLines, 1)
	displayCount := min(len(items), maxItems)

	rendered := make([]string, 0, displayCount)
	for i := range displayCount {
		rendered = append(rendered, RenderActivityPreviewItem(items[i], width))
	}

	return strings.Join(rendered, "\n\n")
}
