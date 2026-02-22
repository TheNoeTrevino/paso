package standup

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

type DateGroup struct {
	Date string
	Logs []models.StandupLog
}

// ComputeSince calculates the "since" time from days and weeks, going back from the given reference time.
// Returns the start of the day (midnight) that many days+weeks ago.
func ComputeSince(now time.Time, days, weeks int) time.Time {
	totalDays := days + (weeks * 7)
	since := now.AddDate(0, 0, -totalDays)
	return time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, since.Location())
}

// GroupLogsByDate groups standup logs by their date (newest date first).
// Logs within each group preserve their original order (newest first).
func GroupLogsByDate(logs []models.StandupLog) []DateGroup {
	if len(logs) == 0 {
		return nil
	}

	orderMap := map[string]int{}
	groupMap := map[string][]models.StandupLog{}

	for _, l := range logs {
		dateKey := l.CreatedAt.Format("2006-01-02")
		if _, exists := groupMap[dateKey]; !exists {
			orderMap[dateKey] = len(orderMap)
		}
		groupMap[dateKey] = append(groupMap[dateKey], l)
	}

	groups := make([]DateGroup, 0, len(groupMap))
	ordered := make([]string, len(orderMap))
	for key, idx := range orderMap {
		ordered[idx] = key
	}

	for _, dateKey := range ordered {
		t, _ := time.Parse("2006-01-02", dateKey)
		groups = append(groups, DateGroup{
			Date: t.Format("Mon, Jan 02 2006"),
			Logs: groupMap[dateKey],
		})
	}

	return groups
}

// FormatGenerateQuiet returns log IDs as strings
func FormatGenerateQuiet(logs []models.StandupLog) []string {
	ids := make([]string, len(logs))
	for i, l := range logs {
		ids[i] = fmt.Sprintf("%d", l.ID)
	}
	return ids
}

// FormatGenerateJSON builds the JSON output structure for generate
func FormatGenerateJSON(logs []models.StandupLog, since, until time.Time) map[string]any {
	groups := GroupLogsByDate(logs)

	jsonGroups := make([]map[string]any, len(groups))
	for i, g := range groups {
		items := make([]map[string]any, len(g.Logs))
		for j, l := range g.Logs {
			items[j] = map[string]any{
				"id":         l.ID,
				"project_id": l.ProjectID,
				"content":    l.Content,
				"created_at": l.CreatedAt.UTC().Format(time.RFC3339),
			}
		}
		jsonGroups[i] = map[string]any{
			"date": g.Date,
			"logs": items,
		}
	}

	return map[string]any{
		"success": true,
		"since":   since.UTC().Format(time.RFC3339),
		"until":   until.UTC().Format(time.RFC3339),
		"count":   len(logs),
		"groups":  jsonGroups,
	}
}

// FormatGenerateHuman renders standup logs grouped by date as a timeline
func FormatGenerateHuman(logs []models.StandupLog, since, until time.Time, colorScheme colors.ColorScheme) string {
	var out strings.Builder

	groups := GroupLogsByDate(logs)

	dateStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorScheme.Title))
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorScheme.Subtle))
	idStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorScheme.Normal))
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorScheme.Normal))
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorScheme.Subtle))

	rangeStr := fmt.Sprintf("%s to %s", since.Format("Jan 02, 2006"), until.Format("Jan 02, 2006"))
	fmt.Fprintf(&out, "%s\n\n", headerStyle.Render(fmt.Sprintf("Standup report (%s) - %d log(s)", rangeStr, len(logs))))

	for i, g := range groups {
		out.WriteString(dateStyle.Render(g.Date))
		out.WriteString("\n")

		for _, l := range g.Logs {
			timestamp := l.CreatedAt.Format("3:04 PM")
			idStr := fmt.Sprintf("#%d", l.ID)

			out.WriteString("  ")
			out.WriteString(idStyle.Render(idStr))
			out.WriteString(" ")
			out.WriteString(timeStyle.Render(timestamp))
			out.WriteString("\n")

			for _, line := range strings.Split(l.Content, "\n") {
				out.WriteString("    ")
				out.WriteString(contentStyle.Render(line))
				out.WriteString("\n")
			}
		}

		if i < len(groups)-1 {
			out.WriteString("\n")
		}
	}

	return out.String()
}
