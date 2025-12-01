package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// View renders the current state of the application
// This implements the "View" part of the Model-View-Update pattern
func (m Model) View() string {
	// Wait for terminal size to be initialized
	if m.uiState.Width() == 0 {
		return "Loading..."
	}

	// Handle ticket form mode: show huh form in centered dialog
	if m.uiState.Mode() == state.TicketFormMode && m.formState.TicketForm != nil {
		formView := m.formState.TicketForm.View()

		// Wrap form in a styled container
		formBox := FormBoxStyle.
			Width(m.uiState.Width() / 2).
			Height(m.uiState.Height() / 2).
			Render(formView)

		return lipgloss.Place(
			m.uiState.Width(), m.uiState.Height(),
			lipgloss.Center, lipgloss.Center,
			formBox,
		)
	}

	// Handle project form mode: show huh form in centered dialog
	if m.uiState.Mode() == state.ProjectFormMode && m.formState.ProjectForm != nil {
		formView := m.formState.ProjectForm.View()

		// Wrap form in a styled container with green border for creation
		formBox := ProjectFormBoxStyle.
			Width(m.uiState.Width() / 2).
			Height(m.uiState.Height() / 3).
			Render("New Project\n\n" + formView)

		return lipgloss.Place(
			m.uiState.Width(), m.uiState.Height(),
			lipgloss.Center, lipgloss.Center,
			formBox,
		)
	}

	// Handle column creation mode: show centered dialog with green border
	if m.uiState.Mode() == state.AddColumnMode {
		inputBox := CreateInputBoxStyle.
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.inputState.Prompt, m.inputState.Buffer))

		return lipgloss.Place(
			m.uiState.Width(), m.uiState.Height(),
			lipgloss.Center, lipgloss.Center,
			inputBox,
		)
	}

	// Handle column edit mode: show centered dialog with blue border
	if m.uiState.Mode() == state.EditColumnMode {
		inputBox := EditInputBoxStyle.
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.inputState.Prompt, m.inputState.Buffer))

		return lipgloss.Place(
			m.uiState.Width(), m.uiState.Height(),
			lipgloss.Center, lipgloss.Center,
			inputBox,
		)
	}

	// Handle delete confirmation mode: show centered dialog
	if m.uiState.Mode() == state.DeleteConfirmMode {
		task := m.getCurrentTask()
		if task != nil {
			confirmBox := DeleteConfirmBoxStyle.
				Width(50).
				Render(fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", task.Title))

			return lipgloss.Place(
				m.uiState.Width(), m.uiState.Height(),
				lipgloss.Center, lipgloss.Center,
				confirmBox,
			)
		}
	}

	// Handle delete column confirmation mode: show centered dialog with task count warning
	if m.uiState.Mode() == state.DeleteColumnConfirmMode {
		column := m.getCurrentColumn()
		if column != nil {
			var content string
			taskCount := m.inputState.DeleteColumnTaskCount
			if taskCount > 0 {
				content = fmt.Sprintf(
					"Delete column '%s'?\nThis will also delete %d task(s).\n\n[y]es  [n]o",
					column.Name,
					taskCount,
				)
			} else {
				content = fmt.Sprintf("Delete column '%s'?\n\n[y]es  [n]o", column.Name)
			}

			confirmBox := DeleteConfirmBoxStyle.
				Width(50).
				Render(content)

			return lipgloss.Place(
				m.uiState.Width(), m.uiState.Height(),
				lipgloss.Center, lipgloss.Center,
				confirmBox,
			)
		}
	}

	// Handle help mode: show keyboard shortcuts
	if m.uiState.Mode() == state.HelpMode {
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

		helpBox := HelpBoxStyle.
			Width(50).
			Render(helpContent)

		return lipgloss.Place(
			m.uiState.Width(), m.uiState.Height(),
			lipgloss.Center, lipgloss.Center,
			helpBox,
		)
	}

	// Handle label picker mode: show picker over task view
	if m.uiState.Mode() == state.LabelPickerMode {
		// Render the label picker content
		var pickerContent string
		if m.labelPickerState.CreateMode {
			// Show color picker
			pickerContent = RenderLabelColorPicker(
				GetDefaultLabelColors(),
				m.labelPickerState.ColorIdx,
				m.formState.FormLabelName,
				m.uiState.Width()/2-8,
			)
		} else {
			// Show label list (use filtered items from state)
			pickerContent = RenderLabelPicker(
				m.getFilteredLabelPickerItems(),
				m.labelPickerState.Cursor,
				m.labelPickerState.Filter,
				true, // show create option
				m.uiState.Width()/2-8,
				m.uiState.Height()/2-4,
			)
		}

		// Wrap in styled container - use different style for create mode
		var pickerBox string
		if m.labelPickerState.CreateMode {
			pickerBox = LabelPickerCreateBoxStyle.
				Width(m.uiState.Width() / 2).
				Height(m.uiState.Height() / 2).
				Render(pickerContent)
		} else {
			pickerBox = LabelPickerBoxStyle.
				Width(m.uiState.Width() / 2).
				Height(m.uiState.Height() / 2).
				Render(pickerContent)
		}

		return lipgloss.Place(
			m.uiState.Width(), m.uiState.Height(),
			lipgloss.Center, lipgloss.Center,
			pickerBox,
		)
	}

	if m.uiState.Mode() == state.ViewTaskMode && m.uiState.ViewingTask() != nil {
		// Find the column name for the task
		columnName := "Unknown"
		for _, col := range m.appState.Columns() {
			if col.ID == m.uiState.ViewingTask().ColumnID {
				columnName = col.Name
				break
			}
		}

		return components.RenderTaskView(components.TaskViewProps{
			Task:         m.uiState.ViewingTask(),
			ColumnName:   columnName,
			PopupWidth:   m.uiState.Width() / 2,
			PopupHeight:  m.uiState.Height() / 2,
			ScreenWidth:  m.uiState.Width(),
			ScreenHeight: m.uiState.Height(),
		})
	}

	// Normal mode: render kanban board
	// Handle empty column list edge case
	if len(m.appState.Columns()) == 0 {
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
	endIdx := min(m.uiState.ViewportOffset()+m.uiState.ViewportSize(), len(m.appState.Columns()))
	visibleColumns := m.appState.Columns()[m.uiState.ViewportOffset():endIdx]

	// Calculate column height: terminal height minus tabs, header, status bar, footer, and margins
	// Tab bar (3) + empty line (1) + Header (1) + status bar (1) + empty line (1) + empty line (1) + footer (1) + margins (2) = ~11
	columnHeight := m.uiState.Height() - 11
	if columnHeight < 10 {
		columnHeight = 10 // Minimum height
	}

	// Render only visible columns
	var columns []string
	for i, col := range visibleColumns {
		// Calculate global index for selection check
		globalIndex := m.uiState.ViewportOffset() + i

		tasks := m.appState.Tasks()[col.ID]

		// Determine selection state for this column
		isSelected := (globalIndex == m.uiState.SelectedColumn())

		// Determine which task is selected (only for the selected column)
		selectedTaskIdx := -1
		if isSelected {
			selectedTaskIdx = m.uiState.SelectedTask()
		}

		columns = append(columns, RenderColumn(col, tasks, isSelected, selectedTaskIdx, columnHeight))
	}

	// Add scroll indicators
	leftArrow := " "
	rightArrow := " "
	if m.uiState.ViewportOffset() > 0 {
		leftArrow = "◀"
	}
	if m.uiState.ViewportOffset()+m.uiState.ViewportSize() < len(m.appState.Columns()) {
		rightArrow = "▶"
	}

	// Layout columns horizontally with scroll indicators
	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	board := lipgloss.JoinHorizontal(lipgloss.Top, leftArrow, " ", columnsView, " ", rightArrow)

	// Get cached total task count
	totalTasks := m.appState.TotalTaskCount()

	// Create project tabs from actual project data
	var projectTabs []string
	for _, project := range m.appState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	tabBar := RenderTabs(projectTabs, m.appState.SelectedProject(), m.uiState.Width())

	// Create header with status bar
	header := TitleStyle.Render("PASO - Your Tasks")
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("%d columns  |  %d tasks  |  Press ? for help", len(m.appState.Columns()), totalTasks))

	// Create error banner if there's an error
	var errorBanner string
	if m.errorState.HasError() {
		errorBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Bold(true).
			Padding(0, 1).
			Render("⚠ " + m.errorState.Get())
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
