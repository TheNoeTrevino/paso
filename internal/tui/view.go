package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current state of the application
// This implements the "View" part of the Model-View-Update pattern
func (m Model) View() string {
	// Wait for terminal size to be initialized
	if m.width == 0 {
		return "Loading..."
	}

	// Handle input modes: show centered dialog
	if m.mode == AddTaskMode || m.mode == EditTaskMode {
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(1).
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.inputPrompt, m.inputBuffer))

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			inputBox,
		)
	}

	// Handle column creation mode: show centered dialog with green border
	if m.mode == AddColumnMode {
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("42")). // Green for creation
			Padding(1).
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.inputPrompt, m.inputBuffer))

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			inputBox,
		)
	}

	// Handle column edit mode: show centered dialog with blue border
	if m.mode == EditColumnMode {
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")). // Blue for editing
			Padding(1).
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.inputPrompt, m.inputBuffer))

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			inputBox,
		)
	}

	// Handle delete confirmation mode: show centered dialog
	if m.mode == DeleteConfirmMode {
		task := m.getCurrentTask()
		if task != nil {
			confirmBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")).
				Padding(1).
				Width(50).
				Render(fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", task.Title))

			return lipgloss.Place(
				m.width, m.height,
				lipgloss.Center, lipgloss.Center,
				confirmBox,
			)
		}
	}

	// Handle delete column confirmation mode: show centered dialog with task count warning
	if m.mode == DeleteColumnConfirmMode {
		column := m.getCurrentColumn()
		if column != nil {
			var content string
			if m.deleteColumnTaskCount > 0 {
				content = fmt.Sprintf(
					"Delete column '%s'?\nThis will also delete %d task(s).\n\n[y]es  [n]o",
					column.Name,
					m.deleteColumnTaskCount,
				)
			} else {
				content = fmt.Sprintf("Delete column '%s'?\n\n[y]es  [n]o", column.Name)
			}

			confirmBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")). // Red for deletion
				Padding(1).
				Width(50).
				Render(content)

			return lipgloss.Place(
				m.width, m.height,
				lipgloss.Center, lipgloss.Center,
				confirmBox,
			)
		}
	}

	// Handle help mode: show keyboard shortcuts
	if m.mode == HelpMode {
		helpContent := `PASO - Keyboard Shortcuts

TASKS
  a     Add new task
  e     Edit selected task
  d     Delete selected task
  <     Move task to previous column
  >     Move task to next column

COLUMNS
  C     Create new column (after current)
  R     Rename current column
  X     Delete current column

NAVIGATION
  h/←   Move to previous column
  l/→   Move to next column
  k/↑   Move to previous task
  j/↓   Move to next task
  [     Scroll viewport left
  ]     Scroll viewport right

OTHER
  ?     Show this help screen
  q     Quit application

Press any key to close`

		helpBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")). // Blue for info
			Padding(1, 2).
			Width(50).
			Render(helpContent)

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			helpBox,
		)
	}

	// Normal mode: render kanban board
	// Handle empty column list edge case
	if len(m.columns) == 0 {
		header := TitleStyle.Render("PASO - Your Tasks")
		emptyMsg := "No columns found. Please check database initialization."
		footer := "[q] quit"
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			emptyMsg,
			"",
			footer,
		)
	}

	// Calculate visible columns based on viewport
	endIdx := min(m.viewportOffset+m.viewportSize, len(m.columns))
	visibleColumns := m.columns[m.viewportOffset:endIdx]

	// Render only visible columns
	var columns []string
	for i, col := range visibleColumns {
		// Calculate global index for selection check
		globalIndex := m.viewportOffset + i

		tasks := m.tasks[col.ID]

		// Determine selection state for this column
		isSelected := (globalIndex == m.selectedColumn)

		// Determine which task is selected (only for the selected column)
		selectedTaskIdx := -1
		if isSelected {
			selectedTaskIdx = m.selectedTask
		}

		columns = append(columns, RenderColumn(col, tasks, isSelected, selectedTaskIdx))
	}

	// Add scroll indicators
	leftArrow := " "
	rightArrow := " "
	if m.viewportOffset > 0 {
		leftArrow = "◀"
	}
	if m.viewportOffset+m.viewportSize < len(m.columns) {
		rightArrow = "▶"
	}

	// Layout columns horizontally with scroll indicators
	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	board := lipgloss.JoinHorizontal(lipgloss.Top, leftArrow, " ", columnsView, " ", rightArrow)

	// Calculate total task count
	totalTasks := 0
	for _, tasks := range m.tasks {
		totalTasks += len(tasks)
	}

	// Create header with status bar
	header := TitleStyle.Render("PASO - Your Tasks")
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("%d columns  |  %d tasks  |  Press ? for help", len(m.columns), totalTasks))

	// Create error banner if there's an error
	var errorBanner string
	if m.errorMessage != "" {
		errorBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Bold(true).
			Padding(0, 1).
			Render("⚠ " + m.errorMessage)
	}

	// Create footer with keyboard shortcuts
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("[a]dd  [e]dit  [d]elete  [C]ol  [R]ename  [X]delete  [hjkl]nav  [[]]scroll  [?]help  [q]uit")

	// Combine all elements vertically
	elements := []string{header, statusBar}
	if errorBanner != "" {
		elements = append(elements, errorBanner)
	}
	elements = append(elements, "", board, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, elements...)
}
