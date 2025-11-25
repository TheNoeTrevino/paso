package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/thenoetrevino/paso/internal/tui/components"
)

// View renders the current state of the application
// This implements the "View" part of the Model-View-Update pattern
func (m Model) View() string {
	// Wait for terminal size to be initialized
	if m.width == 0 {
		return "Loading..."
	}

	// Handle ticket form mode: show huh form in centered dialog
	if m.mode == TicketFormMode && m.ticketForm != nil {
		formView := m.ticketForm.View()

		// Wrap form in a styled container
		formBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(1, 2).
			Width(m.width / 2).
			Height(m.height / 2).
			Render(formView)

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			formBox,
		)
	}

	// Handle project form mode: show huh form in centered dialog
	if m.mode == ProjectFormMode && m.projectForm != nil {
		formView := m.projectForm.View()

		// Wrap form in a styled container with green border for creation
		formBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("42")). // Green for creation
			Padding(1, 2).
			Width(m.width / 2).
			Height(m.height / 3).
			Render("New Project\n\n" + formView)

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			formBox,
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
  space View task details
  l     Edit labels (when viewing task)

COLUMNS
  C     Create new column (after current)
  R     Rename current column
  X     Delete current column

NAVIGATION
  h     Move to previous column
  l     Move to next column
  k     Move to previous task
  j     Move to next task
  [     Scroll viewport left
  ]     Scroll viewport right
  {     Move to next project
  }     Move to prev project

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

	// Handle label picker mode: show picker over task view
	if m.mode == LabelPickerMode {
		// Render the label picker content
		var pickerContent string
		if m.labelPickerCreateMode {
			// Show color picker
			pickerContent = RenderLabelColorPicker(
				GetDefaultLabelColors(),
				m.labelPickerColorIdx,
				m.formLabelName,
				m.width/2-8,
			)
		} else {
			// Show label list
			pickerContent = RenderLabelPicker(
				m.labelPickerItems,
				m.labelPickerCursor,
				m.labelPickerFilter,
				true, // show create option
				m.width/2-8,
				m.height/2-4,
			)
		}

		// Wrap in styled container
		borderColor := lipgloss.Color("170") // Purple
		if m.labelPickerCreateMode {
			borderColor = lipgloss.Color("42") // Green for creation
		}

		pickerBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Width(m.width / 2).
			Height(m.height / 2).
			Render(pickerContent)

		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			pickerBox,
		)
	}

	if m.mode == ViewTaskMode && m.viewingTask != nil {
		return components.RenderTaskView(components.TaskViewProps{
			Task:         m.viewingTask,
			DB:           m.db,
			PopupWidth:   m.width / 2,
			PopupHeight:  m.height / 2,
			ScreenWidth:  m.width,
			ScreenHeight: m.height,
		})
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

	// Calculate column height: terminal height minus tabs, header, status bar, footer, and margins
	// Tab bar (3) + empty line (1) + Header (1) + status bar (1) + empty line (1) + empty line (1) + footer (1) + margins (2) = ~11
	columnHeight := m.height - 11
	if columnHeight < 10 {
		columnHeight = 10 // Minimum height
	}

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

		columns = append(columns, RenderColumn(col, tasks, isSelected, selectedTaskIdx, columnHeight))
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

	// Create project tabs from actual project data
	var projectTabs []string
	for _, project := range m.projects {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	tabBar := RenderTabs(projectTabs, m.selectedProject, m.width)

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
	elements := []string{tabBar, "", header, statusBar}
	if errorBanner != "" {
		elements = append(elements, errorBanner)
	}
	elements = append(elements, "", board, "", footer)

	return lipgloss.JoinVertical(lipgloss.Left, elements...)
}
