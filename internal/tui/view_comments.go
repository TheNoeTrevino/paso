package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderCommentsViewContent renders the full comments view content (without layer wrapping)
func (m Model) renderCommentsViewContent(width, height int) string {
	if m.CommentState.TaskID == 0 {
		return "Error: No task selected"
	}

	// Get task title for header
	task := m.getCurrentTask()
	taskTitle := "Unknown Task"
	if task != nil {
		taskTitle = task.Title
	}

	// Title bar
	commentCount := len(m.CommentState.Items)
	titleText := fmt.Sprintf("Task Comments - \"%s\" (%d comments)", taskTitle, commentCount)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Highlight))
	titleBar := titleStyle.Render(titleText)

	// Empty state
	if m.CommentState.IsEmpty() {
		emptyContent := renderEmptyCommentsState(width, height-4) // Reserve for title + help
		helpText := renderCommentsHelpText()
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
		m.CommentState.ScrollOffset,
		len(m.CommentState.Items),
		availableHeight,
	)

	// Render only visible comment cards
	var cards []string
	visibleItems := m.CommentState.Items[startIdx:endIdx]
	for i, item := range visibleItems {
		actualIdx := startIdx + i
		selected := (actualIdx == m.CommentState.Cursor)
		card := components.RenderCommentCard(item.Comment, selected, cardWidth)
		cards = append(cards, card)
	}

	// Join cards with 1 blank line spacing
	cardsContent := strings.Join(cards, "\n\n")

	// Calculate scroll indicators
	scrollIndicators := calculateCommentsScrollIndicators(
		m.CommentState.ScrollOffset,
		startIdx,
		endIdx,
		len(m.CommentState.Items),
	)

	// Combine content
	helpText := renderCommentsHelpText()

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
		"No comments yet.",
		"",
		"Press 'a' to add your first comment.",
		"",
		"Comments help you track context and reasoning",
		"as you work through tasks.",
		"",
	}

	return emptyStyle.Render(strings.Join(lines, "\n"))
}

// renderCommentsHelpText renders the help text at the bottom of the comments view
func renderCommentsHelpText() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	return helpStyle.Render("[↑↓: navigate | Enter/e: edit | a: add | d: delete | Esc: close]")
}

// ScrollIndicators holds the top and bottom scroll indicator strings
type ScrollIndicators struct {
	Top    string
	Bottom string
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
func calculateCommentsScrollIndicators(scrollOffset int, startIdx int, endIdx int, totalComments int) ScrollIndicators {
	indicators := ScrollIndicators{
		Top:    "",
		Bottom: "",
	}

	indicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Align(lipgloss.Center)

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
