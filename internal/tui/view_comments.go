package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderCommentsViewContent renders the full comments view content (without layer wrapping)
func (m Model) renderCommentsViewContent(width, height int) string {
	if m.Forms.Comment.TaskID == 0 {
		return "Error: No task selected"
	}

	// Get task title for header
	task := m.getCurrentTask()
	taskTitle := "Unknown Task"
	if task != nil {
		taskTitle = task.Title
	}

	// Title bar
	itemCount := len(m.Forms.Comment.Items)
	titleText := fmt.Sprintf("Task Activity - \"%s\" (%d items)", taskTitle, itemCount)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Highlight))
	titleBar := titleStyle.Render(titleText)

	// Empty state
	if m.Forms.Comment.IsEmpty() {
		emptyContent := renderEmptyCommentsState(width, height-4) // Reserve for title + help
		helpText := renderEmptyCommentsHelpText()
		return lipgloss.JoinVertical(lipgloss.Left, titleBar, "", emptyContent, "", helpText)
	}

	// Calculate card dimensions (80% of width, responsive)
	cardWidth := (width * 8) / 10
	if cardWidth < 60 {
		cardWidth = 60
	}
	if cardWidth > 100 {
		cardWidth = 100
	}

	// Calculate which comments to render based on scroll position
	availableHeight := height - 4 // Reserve for title + help
	startIdx, endIdx, _ := calculateVisibleCommentRange(
		m.Forms.Comment.ScrollOffset,
		len(m.Forms.Comment.Items),
		availableHeight,
	)

	// Render only visible activity cards
	var cards []string
	visibleItems := m.Forms.Comment.Items[startIdx:endIdx]
	for i, item := range visibleItems {
		actualIdx := startIdx + i
		selected := (actualIdx == m.Forms.Comment.Cursor)
		card := components.RenderActivityCard(item.Activity, selected, cardWidth)
		cards = append(cards, card)
	}

	// Join cards with 1 blank line spacing
	cardsContent := strings.Join(cards, "\n\n")

	// Calculate scroll indicators
	scrollIndicators := calculateCommentsScrollIndicators(
		m.Forms.Comment.ScrollOffset,
		startIdx,
		endIdx,
		len(m.Forms.Comment.Items),
		cardWidth,
	)

	// Combine content
	selectedActivity := m.Forms.Comment.GetSelectedActivity()
	helpText := renderCommentsHelpText(selectedActivity)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleBar,
		"",
		scrollIndicators.Top,
		cardsContent,
		scrollIndicators.Bottom,
		"",
		helpText,
	)

	return content
}

// renderEmptyCommentsState renders the empty state when there are no comments
func renderEmptyCommentsState(width, height int) string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true).
		Align(lipgloss.Center).
		Width(width)

	lines := []string{
		"",
		"No activity yet.",
		"",
		"Press 'a' to add your first comment.",
		"",
		"Activity includes comments and task events",
		"to help track progress and context.",
		"",
	}

	return emptyStyle.Render(strings.Join(lines, "\n"))
}

// renderEmptyCommentsHelpText renders help text for the empty state.
// Only shows options that are available when there are no items.
func renderEmptyCommentsHelpText() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	return helpStyle.Render("[a: add comment | Esc: close]")
}

// renderCommentsHelpText renders the help text at the bottom of the comments view.
// It shows different help text based on whether an event or comment is selected.
func renderCommentsHelpText(selectedActivity *models.ActivityItem) string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	// Show limited options when an event is selected (events are read-only)
	if selectedActivity != nil && selectedActivity.Type == models.ActivityTypeEvent {
		return helpStyle.Render("[↑↓: navigate | a: add comment | Esc: close] (events are read-only)")
	}

	// Show full options when a comment is selected or no selection
	return helpStyle.Render("[↑↓: navigate | e: edit | a: add comment | d: delete | Esc: close]")
}

// ScrollIndicators holds the top and bottom scroll indicator strings
type ScrollIndicators struct {
	Top    string
	Bottom string
}

// commentsViewAvailableHeight calculates the available height for comments content
// based on the total terminal height. This accounts for the layer scaling (80%)
// and reserves space for the title bar and help text.
func commentsViewAvailableHeight(totalHeight int) int {
	layerHeight := totalHeight * 8 / 10
	return layerHeight - 4 // Reserve for title + help
}

// calculateVisibleCommentRange determines which comments should be rendered
// based on available height and scroll offset.
//
// Returns:
//   - startIdx: index of first comment to render
//   - endIdx: index after last comment to render (for slicing)
//   - maxVisible: maximum number of comments that can fit
func calculateVisibleCommentRange(scrollOffset int, totalComments int, availableHeight int) (startIdx, endIdx, maxVisible int) {
	// Estimate how many comments fit in available height
	// Reserve some overhead for scroll indicators
	const EstimatedCommentCardHeight = 7 // 4 lines card + 2 lines spacing + buffer
	const IndicatorOverhead = 2          // Space for scroll indicators

	workingHeight := availableHeight - IndicatorOverhead
	maxVisible = max(workingHeight/EstimatedCommentCardHeight, 1)

	startIdx = max(0, scrollOffset)
	endIdx = min(startIdx+maxVisible, totalComments)

	return startIdx, endIdx, maxVisible
}

// calculateCommentsScrollIndicators determines if scroll indicators should be shown
func calculateCommentsScrollIndicators(scrollOffset int, startIdx int, endIdx int, totalComments int, width int) ScrollIndicators {
	indicators := ScrollIndicators{
		Top:    "",
		Bottom: "",
	}

	indicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Align(lipgloss.Center).
		Width(width)

	// Show top indicator if scrolled down
	if scrollOffset > 0 {
		indicators.Top = indicatorStyle.Render("▲ more above")
	}

	// Show bottom indicator if more comments below
	if endIdx < totalComments {
		indicators.Bottom = indicatorStyle.Render("▼ more below")
	}

	return indicators
}
