package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// RenderTabs renders a tab bar with the given tab names
// selectedIdx indicates which tab is active (0-indexed)
// width is the total width to fill with the tab gap
func RenderTabs(tabs []string, selectedIdx int, width int) string {
	var renderedTabs []string

	for i, tabName := range tabs {
		if i == selectedIdx {
			renderedTabs = append(renderedTabs, ActiveTabStyle.Render(tabName))
		} else {
			renderedTabs = append(renderedTabs, TabStyle.Render(tabName))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Fill remaining space with gap
	gapWidth := max(width-lipgloss.Width(row)-2, 0)
	gap := TabGapStyle.Render(strings.Repeat(" ", gapWidth))

	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
}

// RenderTask renders a single task as a card
// This is a pure, reusable component that displays task title and labels
//
// Format (as a card with border):
//
//	┌─────────────────────┐
//	│ {Task Title}        │
//	│ [label1] [label2]   │
//	└─────────────────────┘
//
// When selected is true, the task is highlighted with:
//   - Bold text
//   - Purple border color
//   - Brighter background
func RenderTask(task *models.TaskSummary, selected bool) string {
	// Format task content with title
	title := lipgloss.NewStyle().Bold(true).Render(task.Title)

	// Render label chips
	var labelChips string
	if len(task.Labels) > 0 {
		var chips []string
		for _, label := range task.Labels {
			chips = append(chips, components.RenderLabelChip(label))
		}
		labelChips = "\n" + strings.Join(chips, "")
	}

	content := title + labelChips

	// Apply selection styling if this task is selected
	style := TaskStyle
	if selected {
		style = style.
			BorderForeground(lipgloss.Color("170")). // Purple border when selected
			Background(lipgloss.Color("237")).       // Lighter background when selected
			BorderStyle(lipgloss.ThickBorder())
	}

	return style.Render(content)
}

// RenderColumn renders a complete column with its title and tasks
// This is a pure, reusable component that composes individual task components
//
// Layout:
//
//	{Column Name} ({count})
//
//	{Task 1}
//	{Task 2}
//	...
//
// Parameters:
//   - column: The column to render
//   - tasks: Task summaries in this column
//   - selected: Whether this column is currently selected
//   - selectedTaskIdx: Index of selected task in this column (-1 if not this column)
//   - height: Fixed height for the column (0 for auto)
func RenderColumn(column *models.Column, tasks []*models.TaskSummary, selected bool, selectedTaskIdx int, height int) string {
	// Render column title with task count
	header := fmt.Sprintf("%s (%d)", column.Name, len(tasks))
	content := TitleStyle.Render(header) + "\n\n"

	// Render all tasks in the column or show empty state
	if len(tasks) == 0 {
		// Empty column - show helpful message
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Padding(1, 0)
		content += emptyStyle.Render("No tasks")
	} else {
		var taskViews []string
		for i, task := range tasks {
			// Task is selected if this is the selected column and matches the index
			isTaskSelected := selected && i == selectedTaskIdx
			taskViews = append(taskViews, RenderTask(task, isTaskSelected))
		}

		// Join tasks with newlines
		content += strings.Join(taskViews, "\n")
	}

	// Apply column styling with selection highlight and fixed height
	style := ColumnStyle
	if selected {
		style = style.BorderForeground(lipgloss.Color("170"))
	}
	if height > 0 {
		style = style.Height(height)
	}

	return style.Render(content)
}

// ScrollIndicators holds the left and right scroll arrow indicators
type ScrollIndicators struct {
	Left  string
	Right string
}

// GetScrollIndicators returns the appropriate scroll arrows based on viewport position
func GetScrollIndicators(viewportOffset, viewportSize, columnCount int) ScrollIndicators {
	return ScrollIndicators{
		Left:  getLeftArrow(viewportOffset),
		Right: getRightArrow(viewportOffset, viewportSize, columnCount),
	}
}

// getLeftArrow returns "◀" if there are columns to the left, otherwise space
func getLeftArrow(viewportOffset int) string {
	if viewportOffset > 0 {
		return "◀"
	}
	return " "
}

// getRightArrow returns "▶" if there are columns to the right, otherwise space
func getRightArrow(viewportOffset, viewportSize, columnCount int) string {
	if viewportOffset+viewportSize < columnCount {
		return "▶"
	}
	return " "
}

// RenderInfoBanner renders an info notification banner with the given message
func RenderInfoBanner(message string) string {
	return InfoBannerStyle.Render("🔔 " + message)
}

// RenderWarningBanner renders a warning notification banner with the given message
func RenderWarningBanner(message string) string {
	return WarningBannerStyle.Render("⚠ " + message)
}

// RenderErrorBanner renders the error banner with the given error message
func RenderErrorBanner(message string) string {
	return ErrorBannerStyle.Render("✕ " + message)
}

// RenderNotificationBanner renders a notification banner based on its level
func RenderNotificationBanner(n state.Notification) string {
	switch n.Level {
	case state.LevelInfo:
		return RenderInfoBanner(n.Message)
	case state.LevelWarning:
		return RenderWarningBanner(n.Message)
	case state.LevelError:
		return RenderErrorBanner(n.Message)
	default:
		return RenderInfoBanner(n.Message)
	}
}
