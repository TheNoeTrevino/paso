package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/spinner"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

const (
	descriptionMaxLines = 6
	// detailPanelPadding accounts for DetailPanelStyle padding (1 top + 1 bottom)
	detailPanelPadding = 2
	// detailDividerHeight is the height of each divider between sections (1 line)
	detailDividerHeight = 1
)

// RenderDetailPanel renders the complete detail panel for a task
func RenderDetailPanel(task *models.TaskDetail, width, height int) string {
	if task == nil {
		return renderEmptyDetailPanel(width, height)
	}

	contentWidth := width - 4 // Account for padding (2 left + 2 right)

	// Build fixed sections (everything except comments)
	var fixedSections []string

	fixedSections = append(fixedSections, renderDetailHeader(task, contentWidth))
	fixedSections = append(fixedSections, renderDetailMetadata(task))

	if len(task.Labels) > 0 {
		fixedSections = append(fixedSections, renderDetailLabels(task.Labels))
	}

	if task.Description != "" {
		fixedSections = append(fixedSections, renderDetailDescription(task.Description, contentWidth))
	}

	if len(task.ParentTasks) > 0 || len(task.ChildTasks) > 0 {
		fixedSections = append(fixedSections, renderDetailRelations(task, contentWidth))
	}

	timestampsSection := renderDetailTimestamps(task, contentWidth)
	footerSection := renderDetailFooter(contentWidth)

	divider := SubtleStyle.Render(strings.Repeat("─", contentWidth))
	fixedContent := strings.Join(fixedSections, "\n"+divider+"\n")

	fixedHeight := lipgloss.Height(fixedContent)
	timestampsHeight := lipgloss.Height(timestampsSection)
	footerHeight := lipgloss.Height(footerSection)

	// Dividers count: fixed sections have internal dividers already in fixedContent.
	// When comments exist: +1 divider before comments, +1 after comments, +1 before footer = 3
	// When no comments: +1 divider before timestamps, +1 before footer = 2
	extraDividers := 3
	if len(task.Comments) == 0 {
		extraDividers = 2
	}

	usedHeight := fixedHeight + timestampsHeight + footerHeight +
		(extraDividers * detailDividerHeight) + detailPanelPadding

	remainingHeight := max(height-usedHeight, 1)

	// Build final content with comments filling remaining space
	var allSections []string
	allSections = append(allSections, fixedSections...)

	if len(task.Comments) > 0 {
		allSections = append(allSections, renderDetailComments(task.Comments, contentWidth, remainingHeight))
	}

	allSections = append(allSections, timestampsSection)
	allSections = append(allSections, footerSection)

	content := strings.Join(allSections, "\n"+divider+"\n")

	return DetailPanelStyle.
		Width(width).
		Height(height).
		Render(content)
}

// RenderDetailPanelLoading renders the detail panel in a loading state
func RenderDetailPanelLoading(width, height int, spinnerFrame int) string {
	spinnerFrame = spinnerFrame % spinner.FrameCount()
	spinnerIcon := spinner.Frames[spinnerFrame]

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
	var header strings.Builder
	header.WriteString(ticketStyle.Render(ticketNumber))

	titleStyle := lipgloss.NewStyle().Bold(true)
	title := task.Title

	// Wrap title if needed
	maxTitleWidth := width - len(ticketNumber) - 3
	if len(title) > maxTitleWidth {
		wrappedTitle := wordwrap.String(title, width)
		lines := strings.Split(wrappedTitle, "\n")
		title = lines[0]
		if len(lines) > 1 {
			header.WriteString("\n" + titleStyle.Render(title))
			for i := 1; i < len(lines); i++ {
				header.WriteString("\n" + titleStyle.Render(lines[i]))
			}
			return header.String()
		}
	}

	return header.String() + "\n" + titleStyle.Render(title)
}

// renderDetailMetadata renders status, type, and priority
func renderDetailMetadata(task *models.TaskDetail) string {
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

	// Assignee
	if task.AssigneeName != nil {
		assigneeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Highlight))
		assigneeLine := labelStyle.Render("Assignee: ") + assigneeStyle.Render("@"+*task.AssigneeName)
		parts = append(parts, assigneeLine)
	}

	// Estimate
	if task.Estimate != nil && *task.Estimate != "" {
		estimateLine := labelStyle.Render("Estimate: ") + *task.Estimate
		parts = append(parts, estimateLine)
	}

	// Due Date
	if task.DueDate != nil {
		dueDateText, dueDateColor := FormatRelativeDueDate(task.DueDate)
		dueDateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(dueDateColor))
		dueDateLine := labelStyle.Render("Due: ") + dueDateStyle.Render(dueDateText)
		parts = append(parts, dueDateLine)
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
func renderDetailLabels(labels []*models.Label) string {
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

// renderDetailComments renders a height-aware comments section using compact preview format.
// Converts comments to ActivityItems and delegates to RenderActivityPreviews.
// Shows as many recent comments as fit in the available height.
func renderDetailComments(comments []*models.Comment, width, availableHeight int) string {
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

	// Convert comments to ActivityItems for shared rendering.
	// Comments are stored oldest-first; reverse so most recent appear first.
	items := make([]models.ActivityItem, len(comments))
	for i, c := range comments {
		items[len(comments)-1-i] = models.ActivityItem{
			ID:        c.ID,
			TaskID:    c.TaskID,
			Type:      models.ActivityTypeComment,
			Content:   c.Message,
			Author:    c.Author,
			CreatedAt: c.CreatedAt,
		}
	}

	// Subtract header line from available height
	previewHeight := max(availableHeight-1, 1)
	previews := RenderActivityPreviews(items, width, previewHeight)

	return header + "\n" + previews
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
