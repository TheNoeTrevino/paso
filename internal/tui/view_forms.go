package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/state"
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

// renderFormAssociationsZone renders the bottom-left zone with parent and child tasks
func (m Model) renderFormAssociationsZone(width, height int) string {
	var parts []string

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	taskStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Normal))

	// Parent Tasks section
	parts = append(parts, headerStyle.Render("Parent Tasks"))
	if len(m.FormState.FormParentRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No Parent Tasks Found"))
	} else {
		for _, parent := range m.FormState.FormParentRefs {
			// Render relation label with color if available
			var relationLabel string
			if parent.RelationLabel != "" && parent.RelationColor != "" {
				labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(parent.RelationColor))
				relationLabel = labelStyle.Render(parent.RelationLabel)
			} else {
				// Fallback to default if no relation type
				relationLabel = subtleStyle.Render("Parent")
			}
			taskLine := fmt.Sprintf("#%d - %s - %s", parent.TicketNumber, relationLabel, parent.Title)
			parts = append(parts, taskStyle.Render(taskLine))
		}
	}
	parts = append(parts, "")

	// Child Tasks section
	parts = append(parts, headerStyle.Render("Child Tasks"))
	if len(m.FormState.FormChildRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No Child Tasks Found"))
	} else {
		for _, child := range m.FormState.FormChildRefs {
			// Render relation label with color if available
			var relationLabel string
			if child.RelationLabel != "" && child.RelationColor != "" {
				labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(child.RelationColor))
				relationLabel = labelStyle.Render(child.RelationLabel)
			} else {
				// Fallback to default if no relation type
				relationLabel = subtleStyle.Render("Child")
			}
			taskLine := fmt.Sprintf("#%d - %s - %s", child.TicketNumber, relationLabel, child.Title)
			parts = append(parts, taskStyle.Render(taskLine))
		}
	}

	content := strings.Join(parts, "\n")

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		BorderTop(true).
		BorderStyle(lipgloss.Border{
			Top: "─",
		}).
		BorderForeground(lipgloss.Color(theme.Subtle))

	return style.Render(content)
}

// renderFormNotesZone renders the bottom-left zone with notes/comments
func (m Model) renderFormNotesZone(width, height int) string {
	var parts []string

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	noteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Normal))

	// Notes header with count
	noteCount := len(m.FormState.FormComments)
	parts = append(parts, headerStyle.Render(fmt.Sprintf("Notes (%d)", noteCount)))

	if noteCount == 0 {
		parts = append(parts, subtleStyle.Render("No notes. Press Ctrl+N to add one."))
	} else {
		// Display notes (newest first, already sorted by created_at DESC)
		for i, comment := range m.FormState.FormComments {
			// Show timestamp and truncated message
			timestamp := comment.CreatedAt.Format("Jan 2 15:04")

			// Truncate message if too long for display
			message := comment.Message
			maxLen := width - 20 // Leave room for timestamp
			if len(message) > maxLen {
				message = message[:maxLen-3] + "..."
			}

			noteLine := fmt.Sprintf("[%s] %s", timestamp, message)

			// Highlight if this is the cursor position in NoteEditMode
			if m.UiState.Mode() == state.NoteEditMode && m.NoteState.Cursor == i {
				highlightStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(theme.Highlight)).
					Bold(true)
				parts = append(parts, highlightStyle.Render("▶ "+noteLine))
			} else {
				parts = append(parts, noteStyle.Render("  "+noteLine))
			}
		}
		parts = append(parts, "")
		parts = append(parts, subtleStyle.Render("Ctrl+N: manage notes"))
	}

	content := strings.Join(parts, "\n")

	noteZoneStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		BorderTop(true).
		BorderStyle(lipgloss.Border{
			Top: "─",
		}).
		BorderForeground(lipgloss.Color(theme.Subtle))

	return noteZoneStyle.Render(content)
}
