package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderFormTitleDescriptionZone renders the top-left zone with title and description fields
func (m Model) renderFormTitleDescriptionZone(width, height int) string {
	if m.FormState.TicketForm == nil {
		return ""
	}

	// Render the form view (which includes title and description)
	formView := m.FormState.TicketForm.View()

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
	if m.FormState.EditingTaskID == 0 {
		createdStr = subtleStyle.Render("(not created yet)")
		updatedStr = subtleStyle.Render("(not created yet)")
	} else {
		// In edit mode, show actual timestamps from FormState
		createdStr = m.FormState.FormCreatedAt.Format("Jan 2, 2006 3:04 PM")
		updatedStr = m.FormState.FormUpdatedAt.Format("Jan 2, 2006 3:04 PM")
	}

	// Edited indicator (unsaved changes)
	parts = append(parts, labelHeaderStyle.Render("Status"))
	if m.FormState.HasTicketFormChanges() {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight))
		parts = append(parts, warningStyle.Render("● Unsaved Changes"))
	} else {
		parts = append(parts, subtleStyle.Render("○ No Changes"))
	}
	parts = append(parts, "")

	// Type section
	parts = append(parts, labelHeaderStyle.Render("Type"))
	if m.FormState.FormTypeDescription != "" {
		parts = append(parts, m.FormState.FormTypeDescription)
	} else {
		parts = append(parts, subtleStyle.Render("task"))
	}
	parts = append(parts, "")

	// Priority section
	parts = append(parts, labelHeaderStyle.Render("Priority"))
	if m.FormState.FormPriorityDescription != "" && m.FormState.FormPriorityColor != "" {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.FormState.FormPriorityColor))
		parts = append(parts, priorityStyle.Render(m.FormState.FormPriorityDescription))
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
	if len(m.FormState.FormLabelIDs) == 0 {
		parts = append(parts, subtleStyle.Render("No labels"))
	} else {
		// Get label objects from IDs
		labelMap := make(map[int]*models.Label)
		for _, label := range m.AppState.Labels() {
			labelMap[label.ID] = label
		}

		for _, labelID := range m.FormState.FormLabelIDs {
			if label, ok := labelMap[labelID]; ok {
				parts = append(parts, components.RenderLabelChip(label, ""))
			}
		}
	}
	parts = append(parts, "")

	// Parent Tasks section (moved from associations zone)
	parts = append(parts, labelHeaderStyle.Render("Parent Tasks"))
	if len(m.FormState.FormParentRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No parents"))
	} else {
		for _, parent := range m.FormState.FormParentRefs {
			taskLine := fmt.Sprintf("#%d - %s", parent.TicketNumber, parent.Title)
			parts = append(parts, taskLine)
		}
	}
	parts = append(parts, "")

	// Child Tasks section (moved from associations zone)
	parts = append(parts, labelHeaderStyle.Render("Child Tasks"))
	if len(m.FormState.FormChildRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No children"))
	} else {
		for _, child := range m.FormState.FormChildRefs {
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

// renderCommentSimple renders a single comment.
// The format is:
//
//	{author}   {date}
//	{content - wrapped to width}
func renderCommentSimple(comment *models.Comment, width int) string {
	timestamp := comment.CreatedAt.Format("Jan 2 15:04")

	// Author icon and date header
	authorIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("󰀄 ")
	dateIcon := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle)).Render("  ")

	header := fmt.Sprintf("%s%s%s%s", authorIcon, comment.Author, dateIcon, timestamp)
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	// Content (wrapped)
	// Reserve space for padding
	contentWidth := max(width-4, 20)

	wrapped := wordwrap.String(comment.Message, contentWidth)
	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Normal))

	return headerStyle.Render(header) + "\n" + contentStyle.Render(wrapped)
}

// renderFormNotesZone renders the bottom zone with notes/comments using a viewport
func (m *Model) renderFormNotesZone(width, height int) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	// Notes header with count
	noteCount := len(m.FormState.FormComments)
	header := headerStyle.Render(fmt.Sprintf("Notes (%d)", noteCount))

	// Calculate available height for viewport
	// Account for: header (1 line), help text (2 lines), padding/borders
	availableHeight := max(height-5, 1)

	var viewportContent string

	if noteCount == 0 {
		viewportContent = subtleStyle.Render("No notes. Press Ctrl+N to add one.")
	} else {
		// Initialize viewport if not ready
		if !m.FormState.ViewportReady {
			vp := viewport.New()
			vp.SetWidth(width - 2)
			vp.SetHeight(availableHeight)
			vp.Style = lipgloss.NewStyle()
			vp.MouseWheelEnabled = true
			m.FormState.CommentsViewport = vp
			m.FormState.ViewportReady = true
		}

		// Update viewport dimensions in case terminal was resized
		m.FormState.CommentsViewport.SetWidth(width - 2)
		m.FormState.CommentsViewport.SetHeight(availableHeight)

		// Render all comments into viewport content
		var commentLines []string
		for _, comment := range m.FormState.FormComments {
			commentLines = append(commentLines, renderCommentSimple(comment, width-2))
		}
		allComments := strings.Join(commentLines, "\n\n")

		// Set content and scroll to bottom (most recent comment)
		m.FormState.CommentsViewport.SetContent(allComments)
		if !m.FormState.ViewportFocused {
			// Auto-scroll to bottom when not focused (to show most recent)
			m.FormState.CommentsViewport.GotoBottom()
		}

		viewportContent = m.FormState.CommentsViewport.View()
	}

	// Compose content - just header and viewport, separated by blank line
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		"",
		header,
		"",
		viewportContent,
	)

	// Determine border color based on focus (kept for potential future use)
	borderColor := theme.Subtle
	if m.FormState.ViewportFocused {
		borderColor = theme.Highlight
	}
	_ = borderColor // Suppress unused warning for now

	noteZoneStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1, 1, 1)

	return noteZoneStyle.Render(content)
}
