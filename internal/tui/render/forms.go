package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderFormTitleDescriptionZone renders the top-left zone with title and description fields
func (w *Wrapper) renderFormTitleDescriptionZone(width, height int) string {
	if w.FormState.TicketForm == nil {
		return ""
	}

	// Render the form view (which includes title and description)
	formView := w.FormState.TicketForm.View()

	style := lipgloss.NewStyle().
		Width(width).
		Height(height)

	return style.Render(formView)
}

// renderFormMetadataZone renders the right column with metadata
func (w *Wrapper) renderFormMetadataZone(width, height int) string {
	var parts []string

	labelHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	// Get current timestamps - for create mode, show placeholders
	var createdStr, updatedStr string
	if w.FormState.EditingTaskID == 0 {
		createdStr = subtleStyle.Render("(not created yet)")
		updatedStr = subtleStyle.Render("(not created yet)")
	} else {
		// In edit mode, show actual timestamps from FormState
		createdStr = w.FormState.FormCreatedAt.Format("Jan 2, 2006 3:04 PM")
		updatedStr = w.FormState.FormUpdatedAt.Format("Jan 2, 2006 3:04 PM")
	}

	// Edited indicator (unsaved changes)
	parts = append(parts, labelHeaderStyle.Render("Status"))
	if w.FormState.HasTicketFormChanges() {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight))
		parts = append(parts, warningStyle.Render("● Unsaved Changes"))
	} else {
		parts = append(parts, subtleStyle.Render("○ No Changes"))
	}
	parts = append(parts, "")

	// Type section
	parts = append(parts, labelHeaderStyle.Render("Type"))
	if w.FormState.FormTypeDescription != "" {
		parts = append(parts, w.FormState.FormTypeDescription)
	} else {
		parts = append(parts, subtleStyle.Render("task"))
	}
	parts = append(parts, "")

	// Priority section
	parts = append(parts, labelHeaderStyle.Render("Priority"))
	if w.FormState.FormPriorityDescription != "" && w.FormState.FormPriorityColor != "" {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(w.FormState.FormPriorityColor))
		parts = append(parts, priorityStyle.Render(w.FormState.FormPriorityDescription))
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
	if len(w.FormState.FormLabelIDs) == 0 {
		parts = append(parts, subtleStyle.Render("No labels"))
	} else {
		// Get label objects from IDs
		labelMap := make(map[int]*models.Label)
		for _, label := range w.AppState.Labels() {
			labelMap[label.ID] = label
		}

		for _, labelID := range w.FormState.FormLabelIDs {
			if label, ok := labelMap[labelID]; ok {
				parts = append(parts, components.RenderLabelChip(label, ""))
			}
		}
	}
	parts = append(parts, "")

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
func (w *Wrapper) renderFormAssociationsZone(width, height int) string {
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
	if len(w.FormState.FormParentRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No Parent Tasks Found"))
	} else {
		for _, parent := range w.FormState.FormParentRefs {
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
	if len(w.FormState.FormChildRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No Child Tasks Found"))
	} else {
		for _, child := range w.FormState.FormChildRefs {
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
