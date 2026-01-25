package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

const (
	descriptionMaxLines  = 6
	commentPreviewLength = 80
)

// RenderDetailPanel renders the complete detail panel for a task
func RenderDetailPanel(task *models.TaskDetail, width, height int) string {
	if task == nil {
		return renderEmptyDetailPanel(width, height)
	}

	contentWidth := width - 4 // Account for padding (2 left + 2 right)

	var sections []string

	sections = append(sections, renderDetailHeader(task, contentWidth))
	sections = append(sections, renderDetailMetadata(task, contentWidth))

	if len(task.Labels) > 0 {
		sections = append(sections, renderDetailLabels(task.Labels, contentWidth))
	}

	if task.Description != "" {
		sections = append(sections, renderDetailDescription(task.Description, contentWidth))
	}

	if len(task.ParentTasks) > 0 || len(task.ChildTasks) > 0 {
		sections = append(sections, renderDetailRelations(task, contentWidth))
	}

	if len(task.Comments) > 0 {
		sections = append(sections, renderDetailComments(task.Comments, contentWidth))
	}

	sections = append(sections, renderDetailTimestamps(task, contentWidth))
	sections = append(sections, renderDetailFooter(contentWidth))

	divider := SubtleStyle.Render(strings.Repeat("─", contentWidth))
	content := strings.Join(sections, "\n"+divider+"\n")

	return DetailPanelStyle.
		Width(width).
		Height(height).
		Render(content)
}

// RenderDetailPanelLoading renders the detail panel in a loading state
func RenderDetailPanelLoading(width, height int, spinnerFrame int) string {
	spinnerFrame = spinnerFrame % GetSpinnerFrameCount()
	spinnerIcon := []string{
		"󰋙", "󰫃", "󰫄", "󰫅", "󰫆", "󰫇", "󰫈", "󰫇", "󰫆", "󰫅", "󰫄", "󰫃",
	}[spinnerFrame]

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	content := loadingStyle.Render(fmt.Sprintf("Loading task details %s", spinnerIcon))

	return DetailPanelStyle.
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// renderEmptyDetailPanel renders an empty state when no task is selected
func renderEmptyDetailPanel(width, height int) string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	content := emptyStyle.Render("Select a task to view details")

	return DetailPanelStyle.
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// renderDetailHeader renders the ticket number and title
func renderDetailHeader(task *models.TaskDetail, width int) string {
	ticketStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	ticketNumber := fmt.Sprintf("%s-%d", task.ProjectName, task.TicketNumber)
	header := ticketStyle.Render(ticketNumber)

	titleStyle := lipgloss.NewStyle().Bold(true)
	title := task.Title

	// Wrap title if needed
	maxTitleWidth := width - len(ticketNumber) - 3
	if len(title) > maxTitleWidth {
		wrappedTitle := wordwrap.String(title, width)
		lines := strings.Split(wrappedTitle, "\n")
		title = lines[0]
		if len(lines) > 1 {
			header += "\n" + titleStyle.Render(title)
			for i := 1; i < len(lines); i++ {
				header += "\n" + titleStyle.Render(lines[i])
			}
			return header
		}
	}

	return header + "\n" + titleStyle.Render(title)
}

// renderDetailMetadata renders status, type, and priority
func renderDetailMetadata(task *models.TaskDetail, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	var parts []string

	// Status/Column
	statusLine := labelStyle.Render("Status: ") + task.ColumnName
	parts = append(parts, statusLine)

	// Type
	if task.TypeDescription != "" {
		typeLine := labelStyle.Render("Type: ") + task.TypeDescription
		parts = append(parts, typeLine)
	}

	// Priority
	if task.PriorityDescription != "" {
		priorityStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(task.PriorityColor))
		priorityLine := labelStyle.Render("Priority: ") + priorityStyle.Render(task.PriorityDescription)
		parts = append(parts, priorityLine)
	}

	// Blocked status
	if task.IsBlocked {
		blockedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Blocked)).
			Bold(true)
		parts = append(parts, blockedStyle.Render("BLOCKED"))
	}

	return strings.Join(parts, "  ")
}

// renderDetailLabels renders label chips
func renderDetailLabels(labels []*models.Label, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	header := labelStyle.Render("Labels")

	var chips []string
	for _, label := range labels {
		chips = append(chips, RenderLabelChip(label, ""))
	}

	return header + "\n" + strings.Join(chips, " ")
}

// renderDetailDescription renders word-wrapped, truncated description
func renderDetailDescription(description string, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	header := labelStyle.Render("Description")

	wrapped := wordwrap.String(description, width)
	lines := strings.Split(wrapped, "\n")

	if len(lines) > descriptionMaxLines {
		lines = lines[:descriptionMaxLines]
		lines[descriptionMaxLines-1] = strings.TrimRight(lines[descriptionMaxLines-1], " ") +
			SubtleStyle.Render("...")
	}

	return header + "\n" + strings.Join(lines, "\n")
}

// renderDetailRelations renders parent/child task relationships
func renderDetailRelations(task *models.TaskDetail, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	header := labelStyle.Render("Relations")
	var lines []string

	// Parent tasks (tasks this depends on) - arrow pointing up
	for _, parent := range task.ParentTasks {
		arrow := "↑"
		relationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(parent.RelationColor))

		ticketNum := fmt.Sprintf("%s-%d", parent.ProjectName, parent.TicketNumber)
		line := fmt.Sprintf("  %s %s %s",
			arrow,
			relationStyle.Render(parent.RelationLabel+":"),
			ticketNum+" "+truncateString(parent.Title, width-len(ticketNum)-20))

		if parent.IsBlocking {
			blockedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Blocked))
			line = blockedStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Child tasks (tasks that depend on this) - arrow pointing down
	for _, child := range task.ChildTasks {
		arrow := "↓"
		relationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(child.RelationColor))

		ticketNum := fmt.Sprintf("%s-%d", child.ProjectName, child.TicketNumber)
		line := fmt.Sprintf("  %s %s %s",
			arrow,
			relationStyle.Render(child.RelationLabel+":"),
			ticketNum+" "+truncateString(child.Title, width-len(ticketNum)-20))

		if child.IsBlocking {
			blockedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Blocked))
			line = blockedStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return header + "\n" + strings.Join(lines, "\n")
}

// renderDetailComments renders comment count and preview
func renderDetailComments(comments []*models.Comment, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	countText := "comment"
	if len(comments) != 1 {
		countText = "comments"
	}
	header := labelStyle.Render(fmt.Sprintf("Comments (%d %s)", len(comments), countText))

	if len(comments) == 0 {
		return header
	}

	// Show preview of the most recent comment
	latestComment := comments[len(comments)-1]
	preview := latestComment.Message
	if len(preview) > commentPreviewLength {
		preview = preview[:commentPreviewLength] + SubtleStyle.Render("...")
	}

	authorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	return header + "\n" + authorStyle.Render(latestComment.Author+": ") + preview
}

// renderDetailTimestamps renders created/updated timestamps
func renderDetailTimestamps(task *models.TaskDetail, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Highlight)).
		Bold(true)

	timestampStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	createdAt := task.CreatedAt.Format("Jan 2, 2006 3:04 PM")
	updatedAt := task.UpdatedAt.Format("Jan 2, 2006 3:04 PM")

	created := labelStyle.Render("Created: ") + timestampStyle.Render(createdAt)
	updated := labelStyle.Render("Updated: ") + timestampStyle.Render(updatedAt)

	return created + "  " + updated
}

// renderDetailFooter renders the action hint
func renderDetailFooter(width int) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true).
		Align(lipgloss.Center).
		Width(width)

	return footerStyle.Render("Press Enter to view/edit")
}

// truncateString truncates s to maxLen characters, adding "..." suffix if truncated.
// Edge case: when maxLen <= 3, returns s[:maxLen] without ellipsis since "..." itself
// is 3 characters and would be longer than or equal to the allowed content.
func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
