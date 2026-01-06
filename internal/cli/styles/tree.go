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

// CompletedDimIntensity controls how much completed tasks are dimmed (0.0-1.0)
const CompletedDimIntensity = 0.6

// RenderTreeConnector renders the tree connector with appropriate color
// isLast determines if this is the last child (uses └── vs ├──)
// inBlockingPath determines color (red if blocking)
func RenderTreeConnector(isLast bool, inBlockingPath bool, colors colors.ColorScheme) string {
	return RenderTreeConnectorWithCompletion(isLast, inBlockingPath, false, colors)
}

// RenderTreeConnectorWithCompletion renders the tree connector with dim-aware styling
// shouldDim controls whether the connector is dimmed (for completed paths)
func RenderTreeConnectorWithCompletion(isLast bool, inBlockingPath bool, shouldDim bool, colors colors.ColorScheme) string {
	connector := TreeBranch
	if isLast {
		connector = TreeLastBranch
	}

	color := colors.Subtle
	if inBlockingPath {
		color = colors.ErrorFg
	}
	if shouldDim {
		color = DimColor(color, CompletedDimIntensity)
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	if inBlockingPath {
		style = style.Bold(true)
	}
	return style.Render(connector)
}

// RenderTreeVertical renders the vertical continuation line with appropriate color
// inBlockingPath determines color (red if blocking)
// shouldDim controls whether the line is dimmed (for completed paths)
func RenderTreeVertical(inBlockingPath bool, shouldDim bool, colors colors.ColorScheme) string {
	color := colors.Subtle
	if inBlockingPath {
		color = colors.ErrorFg
	}
	if shouldDim {
		color = DimColor(color, CompletedDimIntensity)
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	if inBlockingPath {
		style = style.Bold(true)
	}
	return style.Render(TreeVertical)
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
func RenderTreeRootTask(ticketNumber int, title string, columnName string, isCompleted bool, colors colors.ColorScheme) string {
	color := colors.Title
	if isCompleted {
		color = DimColor(color, CompletedDimIntensity)
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(color)).
		Render(fmt.Sprintf("%d: %s - %s", ticketNumber, title, columnName))
}

// RenderTreeChildLine renders a complete child line in the tree
// Format: "│   ├── 39: [Blocker] Title - ColumnName"
// prefix: accumulated continuation characters from ancestors (e.g., "│   │   ")
// isLast: whether this node is the last child in its parent's children
// shouldDim: whether this line should be dimmed (typically when parent has no uncompleted descendants)
func RenderTreeChildLine(prefix string, isLast bool, node *models.TaskTreeNode, shouldDim bool, colors colors.ColorScheme) string {
	// The connector (├── or └──) - red if in blocking path
	connector := RenderTreeConnectorWithCompletion(isLast, node.InBlockingPath, shouldDim, colors)

	// Determine text color based on blocking status
	textColor := colors.Normal
	if node.IsBlocking {
		textColor = colors.ErrorFg
	}
	if shouldDim {
		textColor = DimColor(textColor, CompletedDimIntensity)
	}

	// Build the content parts
	ticketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))

	relationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(textColor)).
		Bold(node.IsBlocking)

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))

	ticketNum := ticketStyle.Render(fmt.Sprintf("%d: ", node.TicketNumber))
	relationChip := relationStyle.Render(node.RelationLabel)
	titleInfo := titleStyle.Render(fmt.Sprintf(" %s - %s", node.Title, node.ColumnName))

	return fmt.Sprintf("%s%s%s%s%s", prefix, connector, ticketNum, relationChip, titleInfo)
}
