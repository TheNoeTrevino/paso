package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/notifications"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// View renders the current state of the application
// This implements the "View" part of the Model-View-Update pattern
func (m Model) View() tea.View {
	var view tea.View
	view.AltScreen = true // Use alternate screen buffer

	// Wait for terminal size to be initialized
	if m.uiState.Width() == 0 {
		view.Content = "Loading..."
		return view
	}

	// Dispatch to appropriate view handler based on mode
	switch m.uiState.Mode() {
	case state.TicketFormMode:
		view.Content = m.viewTicketForm()
	case state.ProjectFormMode:
		view.Content = m.viewProjectForm()
	case state.AddColumnMode, state.EditColumnMode:
		view.Content = m.viewColumnInput()
	case state.DeleteConfirmMode:
		view.Content = m.viewDeleteTaskConfirm()
	case state.DeleteColumnConfirmMode:
		view.Content = m.viewDeleteColumnConfirm()
	case state.HelpMode:
		view.Content = m.viewHelp()
	case state.LabelPickerMode:
		view.Content = m.viewLabelPicker()
	case state.ViewTaskMode:
		view.Content = m.viewTaskDetail()
	default:
		view.Content = m.viewKanbanBoard()
	}

	return view
}

// viewTicketForm renders the ticket creation/edit form modal
func (m Model) viewTicketForm() string {
	if m.formState.TicketForm == nil {
		return ""
	}

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

// viewProjectForm renders the project creation form modal
func (m Model) viewProjectForm() string {
	if m.formState.ProjectForm == nil {
		return ""
	}

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

// viewColumnInput renders the column name input dialog (create or edit mode)
func (m Model) viewColumnInput() string {
	var inputBox string
	if m.uiState.Mode() == state.AddColumnMode {
		inputBox = CreateInputBoxStyle.
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.inputState.Prompt, m.inputState.Buffer))
	} else {
		// EditColumnMode
		inputBox = EditInputBoxStyle.
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.inputState.Prompt, m.inputState.Buffer))
	}

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		inputBox,
	)
}

// viewDeleteTaskConfirm renders the task deletion confirmation dialog
func (m Model) viewDeleteTaskConfirm() string {
	task := m.getCurrentTask()
	if task == nil {
		return ""
	}

	confirmBox := DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("Delete '%s'?\n\n[y]es  [n]o", task.Title))

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}

// viewDeleteColumnConfirm renders the column deletion confirmation with task count warning
func (m Model) viewDeleteColumnConfirm() string {
	column := m.getCurrentColumn()
	if column == nil {
		return ""
	}

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

// viewHelp renders the keyboard shortcuts help screen
func (m Model) viewHelp() string {
	helpBox := HelpBoxStyle.
		Width(50).
		Render(m.generateHelpText())

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		helpBox,
	)
}

// generateHelpText creates help text based on current key mappings
func (m Model) generateHelpText() string {
	km := m.config.KeyMappings
	return fmt.Sprintf(`PASO - Keyboard Shortcuts

TASKS
  %s     Add new task
  %s     Edit selected task
  %s     Delete selected task
  %s     Move task to previous column
  %s     Move task to next column
  %s     Move task up in column
  %s     Move task down in column
  %s     View task details
  %s     Edit labels (when viewing task)

COLUMNS
  %s     Create new column (after current)
  %s     Rename current column
  %s     Delete current column

NAVIGATION
  %s     Move to previous column
  %s     Move to next column
  %s     Move to previous task
  %s     Move to next task
  %s     Scroll viewport left
  %s     Scroll viewport right
  %s     Move to next project
  %s     Move to prev project

OTHER
  %s     Show this help screen
  %s     Quit application

Press any key to close`,
		km.AddTask, km.EditTask, km.DeleteTask,
		km.MoveTaskLeft, km.MoveTaskRight,
		km.MoveTaskUp, km.MoveTaskDown,
		formatKey(km.ViewTask), km.EditLabels,
		km.CreateColumn, km.RenameColumn, km.DeleteColumn,
		km.PrevColumn, km.NextColumn,
		km.PrevTask, km.NextTask,
		km.ScrollViewportLeft, km.ScrollViewportRight,
		km.NextProject, km.PrevProject,
		km.ShowHelp, km.Quit,
	)
}

// formatKey formats special keys for display
func formatKey(key string) string {
	if key == " " {
		return "space"
	}
	return key
}

// viewLabelPicker renders the label picker modal (select or create mode)
func (m Model) viewLabelPicker() string {
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

// viewTaskDetail renders the full task details modal
func (m Model) viewTaskDetail() string {
	if m.uiState.ViewingTask() == nil {
		return ""
	}

	// Find the column name for the task (O(1) lookup)
	columnName := "Unknown"
	if col := m.appState.GetColumnByID(m.uiState.ViewingTask().ColumnID); col != nil {
		columnName = col.Name
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

// viewKanbanBoard renders the main kanban board (normal mode)
func (m Model) viewKanbanBoard() string {
	// Handle empty column list edge case
	if len(m.appState.Columns()) == 0 {
		emptyMsg := "No columns found. Please check database initialization."
		footer := components.RenderStatusBar(components.StatusBarProps{
			Width: m.uiState.Width(),
		})
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			emptyMsg,
			"",
			footer,
		)
	}

	// Calculate visible columns based on viewport
	endIdx := min(m.uiState.ViewportOffset()+m.uiState.ViewportSize(), len(m.appState.Columns()))
	visibleColumns := m.appState.Columns()[m.uiState.ViewportOffset():endIdx]

	// Calculate column height: terminal height minus tabs, footer, and margins
	// Tab bar (3) + empty line (1) + empty line before footer (1) + footer (1) + margins (2) = ~8
	columnHeight := m.uiState.Height() - 8
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

	scrollIndicators := GetScrollIndicators(
		m.uiState.ViewportOffset(),
		m.uiState.ViewportSize(),
		len(m.appState.Columns()),
	)

	// Layout columns horizontally with scroll indicators
	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	board := lipgloss.JoinHorizontal(lipgloss.Top, scrollIndicators.Left, " ", columnsView, " ", scrollIndicators.Right)

	// Create project tabs from actual project data
	var projectTabs []string
	for _, project := range m.appState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	tabBar := RenderTabs(projectTabs, m.appState.SelectedProject(), m.uiState.Width())

	footer := components.RenderStatusBar(components.StatusBarProps{
		Width: m.uiState.Width(),
	})

	// Build base view
	baseView := lipgloss.JoinVertical(lipgloss.Left, tabBar, "", board, "", footer)

	// If no notifications, return base view directly
	if !m.notificationState.HasAny() {
		return baseView
	}

	// Start layer stack with base view
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(baseView),
	}

	// Add notification layers
	notificationLayers := m.notificationState.GetLayers(notifications.RenderFromState)
	layers = append(layers, notificationLayers...)

	// Combine all layers into canvas
	canvas := lipgloss.NewCanvas(layers...)
	return canvas.Render()
}
