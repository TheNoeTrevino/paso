package styles

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/config/colors"
	"github.com/thenoetrevino/paso/internal/models"
)

// Tree connector characters for visual hierarchy
const (
	TreeBranch     = "├── " // middle child connector
	TreeLastBranch = "└── " // last child connector
	TreeVertical   = "│   " // continuation line for non-last siblings
	TreeSpace      = "    " // continuation for last sibling (4 spaces)
)

// RenderTreeConnector renders the tree connector with appropriate color
// isLast determines if this is the last child (uses └── vs ├──)
// inBlockingPath determines color (red if blocking)
func RenderTreeConnector(isLast bool, inBlockingPath bool, colors colors.ColorScheme) string {
	return RenderTreeConnectorWithCompletion(isLast, inBlockingPath, false, colors)
}

// RenderTreeConnectorWithCompletion renders the tree connector with completion-aware styling
func RenderTreeConnectorWithCompletion(isLast bool, inBlockingPath bool, isCompleted bool, colors colors.ColorScheme) string {
	connector := TreeBranch
	if isLast {
		connector = TreeLastBranch
	}

	if inBlockingPath {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colors.ErrorFg)).
			Faint(isCompleted).
			Render(connector)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Subtle)).
		Faint(isCompleted).
		Render(connector)
}

// RenderTreeVertical renders the vertical continuation line with appropriate color
// inBlockingPath determines color (red if blocking)
func RenderTreeVertical(inBlockingPath bool, isCompleted bool, colors colors.ColorScheme) string {
	if inBlockingPath {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colors.ErrorFg)).
			Faint(isCompleted).
			Render(TreeVertical)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Subtle)).
		Faint(isCompleted).
		Render(TreeVertical)
}

// RenderRelationChip renders a relation type label as a chip like "[Blocker]"
func RenderRelationChip(label string, color string, isBlocking bool, colors colors.ColorScheme) string {
	if isBlocking {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colors.ErrorFg)).
			Render(label)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(label)
}

// RenderTreeTaskInfo renders task info for tree display
// Format: "PROJ-123: Title - ColumnName"
func RenderTreeTaskInfo(ticketNumber int, title string, columnName string, isBlocking bool, colors colors.ColorScheme) string {
	taskInfo := fmt.Sprintf("%d: %s - %s", ticketNumber, title, columnName)
	if isBlocking {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.ErrorFg)).
			Render(taskInfo)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Normal)).
		Render(taskInfo)
}

// RenderTreeRootTask renders a root task (no connector, no relation)
func RenderTreeRootTask(ticketNumber int, title string, columnName string, colors colors.ColorScheme) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.Title)).
		Render(fmt.Sprintf("%d: %s - %s", ticketNumber, title, columnName))
}

// RenderTreeChildLine renders a complete child line in the tree
// Format: "│   ├── 39: [Blocker] Title - ColumnName"
// prefix: accumulated continuation characters from ancestors (e.g., "│   │   ")
// isLast: whether this node is the last child in its parent's children
func RenderTreeChildLine(prefix string, isLast bool, node *models.TaskTreeNode, colors colors.ColorScheme) string {
	// The connector (├── or └──) - red if in blocking path
	connector := RenderTreeConnectorWithCompletion(isLast, node.InBlockingPath, node.IsCompleted, colors)

	// Determine text color based on blocking status
	textColor := colors.Normal
	if node.IsBlocking {
		textColor = colors.ErrorFg
	}

	// Build the content parts with faint applied if completed
	ticketStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(textColor)).
		Faint(node.IsCompleted)

	relationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(textColor)).
		Bold(node.IsBlocking).
		Faint(node.IsCompleted)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(textColor)).
		Faint(node.IsCompleted)

	ticketNum := ticketStyle.Render(fmt.Sprintf("%d: ", node.TicketNumber))
	relationChip := relationStyle.Render(node.RelationLabel)
	titleInfo := titleStyle.Render(fmt.Sprintf(" %s - %s", node.Title, node.ColumnName))

	return fmt.Sprintf("%s%s%s%s%s", prefix, connector, ticketNum, relationChip, titleInfo)
}
