package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
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

	// Assignee section
	parts = append(parts, labelHeaderStyle.Render("Assignee"))
	if m.Forms.Form.FormAssigneeName != "" {
		assigneeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight))
		parts = append(parts, assigneeStyle.Render("@"+m.Forms.Form.FormAssigneeName))
	} else {
		parts = append(parts, subtleStyle.Render("unassigned"))
	}
	parts = append(parts, "")

	// Estimate section
	parts = append(parts, labelHeaderStyle.Render("Estimate"))
	if m.Forms.Form.FormEstimate != "" {
		estimateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight))
		parts = append(parts, estimateStyle.Render(m.Forms.Form.FormEstimate))
	} else {
		parts = append(parts, subtleStyle.Render("none"))
	}
	parts = append(parts, "")

	// Due Date section
	parts = append(parts, labelHeaderStyle.Render("Due Date"))
	if m.Forms.Form.FormDueDate != nil {
		dueDateText, dueDateColor := components.FormatRelativeDueDate(m.Forms.Form.FormDueDate)
		dueDateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(dueDateColor))
		parts = append(parts, dueDateStyle.Render(dueDateText))
	} else {
		parts = append(parts, subtleStyle.Render("none"))
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
			taskLine := fmt.Sprintf("#%d - %s", parent.TaskNumber, parent.Title)
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
			taskLine := fmt.Sprintf("#%d - %s", child.TaskNumber, child.Title)
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

// renderFormCommentsPreview renders a read-only preview of recent activity (events + comments)
// Users press Ctrl+N to open the full activity view
func (m *Model) renderFormCommentsPreview(width, height int) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	activityCount := len(m.Forms.Form.FormActivities)

	// Header: "Activity · {count} total · ctrl+n to open"
	var headerText string
	if activityCount == 0 {
		headerText = "Activity · ctrl+n to add"
	} else {
		headerText = fmt.Sprintf("Recent Activity · %d total · ctrl+n to open", activityCount)
	}
	header := headerStyle.Render(headerText)

	// Calculate available height for preview (excluding header, border, padding)
	// Account for: header (1), blank line (1), top border (1), padding (2) = 5 lines
	availableHeight := max(height-5, 1)

	var previewContent string

	if activityCount == 0 {
		previewContent = subtleStyle.Render("No activity yet · ctrl+n to add comment")
	} else {
		previewContent = components.RenderActivityPreviews(
			m.Forms.Form.FormActivities,
			width,
			availableHeight,
		)
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
