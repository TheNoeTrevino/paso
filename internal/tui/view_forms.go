package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderFormTitleDescriptionZone renders the top-left zone with title and description fields
func (m Model) renderFormTitleDescriptionZone(width, height int) string {
	if m.Forms.Form.TaskForm == nil {
		return ""
	}

	formView := m.Forms.Form.TaskForm.View()

	style := lipgloss.NewStyle().
		Width(width).
		Height(height)

	return style.Render(formView)
}

// renderFormMetadataZone renders the right column with metadata
func (m Model) renderFormMetadataZone(width, height int) string {
	var parts []string

	labelHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	// Get current timestamps - for create mode, show placeholders
	var createdStr, updatedStr string
	if m.Forms.Form.EditingTaskID == 0 {
		createdStr = subtleStyle.Render("(not created yet)")
		updatedStr = subtleStyle.Render("(not created yet)")
	} else {
		// In edit mode, show actual timestamps from FormState
		createdStr = m.Forms.Form.FormCreatedAt.Format("Jan 2, 2006 3:04 PM")
		updatedStr = m.Forms.Form.FormUpdatedAt.Format("Jan 2, 2006 3:04 PM")
	}

	// Edited indicator (unsaved changes)
	parts = append(parts, labelHeaderStyle.Render("Status"))
	if m.Forms.Form.HasTaskFormChanges() {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight))
		parts = append(parts, warningStyle.Render("● Unsaved Changes"))
	} else {
		parts = append(parts, subtleStyle.Render("○ No Changes"))
	}
	parts = append(parts, "")

	// Type section
	parts = append(parts, labelHeaderStyle.Render("Type"))
	if m.Forms.Form.FormTypeDescription != "" {
		parts = append(parts, m.Forms.Form.FormTypeDescription)
	} else {
		parts = append(parts, subtleStyle.Render("task"))
	}
	parts = append(parts, "")

	// Priority section
	parts = append(parts, labelHeaderStyle.Render("Priority"))
	if m.Forms.Form.FormPriorityDescription != "" && m.Forms.Form.FormPriorityColor != "" {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Forms.Form.FormPriorityColor))
		parts = append(parts, priorityStyle.Render(m.Forms.Form.FormPriorityDescription))
	} else {
		// Default to medium priority if not set
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EAB308"))
		parts = append(parts, priorityStyle.Render("medium"))
	}
	parts = append(parts, "")

	// Created timestamp
	parts = append(parts, labelHeaderStyle.Render("Created"))
	parts = append(parts, createdStr)
	parts = append(parts, "")

	// Updated timestamp
	parts = append(parts, labelHeaderStyle.Render("Updated"))
	parts = append(parts, updatedStr)
	parts = append(parts, "")

	// Labels section
	parts = append(parts, labelHeaderStyle.Render("Labels"))
	if len(m.Forms.Form.FormLabelIDs) == 0 {
		parts = append(parts, subtleStyle.Render("No labels"))
	} else {
		// Get label objects from IDs
		labelMap := make(map[int]*models.Label)
		for _, label := range m.AppState.Labels() {
			labelMap[label.ID] = label
		}

		for _, labelID := range m.Forms.Form.FormLabelIDs {
			if label, ok := labelMap[labelID]; ok {
				parts = append(parts, components.RenderLabelChip(label, ""))
			}
		}
	}
	parts = append(parts, "")

	// Parent Tasks section (moved from associations zone)
	parts = append(parts, labelHeaderStyle.Render("Parent Tasks"))
	if len(m.Forms.Form.FormParentRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No parents"))
	} else {
		for _, parent := range m.Forms.Form.FormParentRefs {
			taskLine := fmt.Sprintf("#%d - %s", parent.TicketNumber, parent.Title)
			parts = append(parts, taskLine)
		}
	}
	parts = append(parts, "")

	// Child Tasks section (moved from associations zone)
	parts = append(parts, labelHeaderStyle.Render("Child Tasks"))
	if len(m.Forms.Form.FormChildRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No children"))
	} else {
		for _, child := range m.Forms.Form.FormChildRefs {
			taskLine := fmt.Sprintf("#%d - %s", child.TicketNumber, child.Title)
			parts = append(parts, taskLine)
		}
	}

	content := strings.Join(parts, "\n")

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{
			Left: "│",
		}).
		BorderForeground(lipgloss.Color(theme.Subtle))

	return style.Render(content)
}

// renderFormCommentsPreview renders a read-only preview of recent comments
// Users press Ctrl+N to open the full comments view
func (m *Model) renderFormCommentsPreview(width, height int) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	commentCount := len(m.Forms.Form.FormComments)

	// Header: "Comments · {count} total · ctrl+n to open"
	var headerText string
	if commentCount == 0 {
		headerText = "Comments · ctrl+n to add"
	} else {
		headerText = fmt.Sprintf("Recent Comments · %d total · ctrl+n to open all comments", commentCount)
	}
	header := headerStyle.Render(headerText)

	// Calculate available height for preview (excluding header, border, padding)
	// Account for: header (1), blank line (1), top border (1), padding (2) = 5 lines
	availableHeight := max(height-5, 1)

	var previewContent string

	if commentCount == 0 {
		previewContent = subtleStyle.Render("No comments yet · ctrl+n to add")
	} else {
		// Show most recent comments based on available height
		// Each comment takes ~2-3 lines (header + 1-2 lines content)
		// Being very generous to show as many recent comments as possible
		maxComments := max((availableHeight+1)/2, 1)

		var previewLines []string
		displayCount := min(commentCount, maxComments)

		// Show most recent comments first
		for i := commentCount - 1; i >= max(commentCount-displayCount, 0); i-- {
			comment := m.Forms.Form.FormComments[i]

			// Truncate comment content to fit preview
			contentWidth := max(width-4, 20)
			content := comment.Message
			lines := strings.Split(wordwrap.String(content, contentWidth), "\n")

			// Take first 2-3 lines only
			maxLines := 2
			if len(lines) > maxLines {
				lines = lines[:maxLines]
				lines[maxLines-1] = lines[maxLines-1] + "..."
			}

			// Render comment preview
			timestamp := comment.CreatedAt.Format("Jan 2 15:04")
			authorIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("󰀄 ")
			dateIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("  ")

			commentHeader := fmt.Sprintf("%s%s%s%s", authorIcon, comment.Author, dateIcon, timestamp)
			headerLine := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render(commentHeader)

			contentLines := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal)).Render(strings.Join(lines, "\n"))

			previewLines = append(previewLines, headerLine+"\n"+contentLines)
		}

		previewContent = strings.Join(previewLines, "\n\n")
	}

	// Compose content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		previewContent,
	)

	noteZoneStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1, 1, 1).
		BorderTop(true).
		BorderStyle(lipgloss.Border{
			Top: "─",
		}).
		BorderForeground(lipgloss.Color(theme.Subtle))

	return noteZoneStyle.Render(content)
}
