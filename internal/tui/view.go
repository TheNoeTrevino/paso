package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/models"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/layers"
	"github.com/thenoetrevino/paso/internal/tui/notifications"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// View renders the current state of the application
// This implements the "View" part of the Model-View-Update pattern
func (m Model) View() tea.View {
	var view tea.View
	view.AltScreen = true                                   // Use alternate screen buffer
	view.BackgroundColor = lipgloss.Color(theme.Background) // Set root background color

	// Wait for terminal size to be initialized
	if m.uiState.Width() == 0 {
		view.Content = "Loading..."
		return view
	}

	// Check if current mode uses layer-based rendering
	usesLayers := m.uiState.Mode() == state.TicketFormMode ||
		m.uiState.Mode() == state.ProjectFormMode ||
		m.uiState.Mode() == state.AddColumnMode ||
		m.uiState.Mode() == state.EditColumnMode ||
		m.uiState.Mode() == state.HelpMode ||
		m.uiState.Mode() == state.NormalMode ||
		m.uiState.Mode() == state.SearchMode

	if usesLayers {
		// Layer-based rendering: always show base board with modal overlays
		baseView := m.viewKanbanBoard()

		// Start layer stack with base view
		layers := []*lipgloss.Layer{
			lipgloss.NewLayer(baseView),
		}

		// Add modal overlay based on mode
		var modalLayer *lipgloss.Layer
		switch m.uiState.Mode() {
		case state.TicketFormMode:
			modalLayer = m.renderTicketFormLayer()
		case state.ProjectFormMode:
			modalLayer = m.renderProjectFormLayer()
		case state.AddColumnMode, state.EditColumnMode:
			modalLayer = m.renderColumnInputLayer()
		case state.HelpMode:
			modalLayer = m.renderHelpLayer()
		}

		if modalLayer != nil {
			layers = append(layers, modalLayer)
		}

		// Add notification layers (always on top)
		if m.notificationState.HasAny() {
			notificationLayers := m.notificationState.GetLayers(notifications.RenderFromState)
			layers = append(layers, notificationLayers...)
		}

		// Combine all layers into canvas
		canvas := lipgloss.NewCanvas(layers...)
		view.Content = canvas.Render()
	} else {
		// Legacy full-screen rendering for modes not yet converted to layers
		var content string
		switch m.uiState.Mode() {
		case state.DiscardConfirmMode:
			content = m.viewDiscardConfirm()
		case state.DeleteConfirmMode:
			content = m.viewDeleteTaskConfirm()
		case state.DeleteColumnConfirmMode:
			content = m.viewDeleteColumnConfirm()
		case state.LabelPickerMode:
			content = m.viewLabelPicker()
		case state.ParentPickerMode:
			content = m.viewParentPicker()
		case state.ChildPickerMode:
			content = m.viewChildPicker()
		case state.PriorityPickerMode:
			content = m.viewPriorityPicker()
		case state.StatusPickerMode:
			content = m.viewStatusPicker()
		default:
			content = m.viewKanbanBoard()
		}
		view.Content = content
	}

	return view
}

// renderTicketFormLayer renders the ticket creation/edit form modal as a layer
func (m Model) renderTicketFormLayer() *lipgloss.Layer {
	if m.formState.TicketForm == nil {
		return nil
	}

	// Calculate layer dimensions (80% of screen)
	layerWidth := m.uiState.Width() * 4 / 5
	layerHeight := m.uiState.Height() * 4 / 5

	// Calculate zone dimensions
	leftColumnWidth := layerWidth * 7 / 10  // 70% of layer width
	rightColumnWidth := layerWidth * 3 / 10 // 30% of layer width
	topHeight := layerHeight * 6 / 10       // 60% of layer height
	bottomHeight := layerHeight * 4 / 10    // 40% of layer height

	// Render the three zones
	topLeftZone := m.renderFormTitleDescriptionZone(leftColumnWidth, topHeight)
	bottomLeftZone := m.renderFormAssociationsZone(leftColumnWidth, bottomHeight)
	rightZone := m.renderFormMetadataZone(rightColumnWidth, layerHeight)

	// Compose left column (top + bottom)
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, topLeftZone, bottomLeftZone)

	// Compose full content (left + right)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightZone)

	// Add form title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	var formTitle string
	if m.formState.EditingTaskID == 0 {
		formTitle = titleStyle.Render("Create New Task")
	} else {
		formTitle = titleStyle.Render("Edit Task")
	}

	// Add help text for shortcuts
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	helpText := helpStyle.Render("Ctrl+L: edit labels  Ctrl+P: edit parents  Ctrl+C: edit children  Ctrl+R: edit priority")

	// Combine title + content + help
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		formTitle,
		"",
		content,
		"",
		helpText,
	)

	// Wrap in form box style
	formBox := FormBoxStyle.
		Width(layerWidth).
		Height(layerHeight).
		Render(fullContent)

	return layers.CreateCenteredLayer(formBox, m.uiState.Width(), m.uiState.Height())
}

// renderProjectFormLayer renders the project creation form modal as a layer
func (m Model) renderProjectFormLayer() *lipgloss.Layer {
	if m.formState.ProjectForm == nil {
		return nil
	}

	formView := m.formState.ProjectForm.View()

	// Wrap form in a styled container with green border for creation
	formBox := ProjectFormBoxStyle.
		Width(m.uiState.Width() * 3 / 4).
		Height(m.uiState.Height() / 3).
		Render("New Project\n\n" + formView)

	return layers.CreateCenteredLayer(formBox, m.uiState.Width(), m.uiState.Height())
}

// renderColumnInputLayer renders the column name input dialog (create or edit mode) as a layer
func (m Model) renderColumnInputLayer() *lipgloss.Layer {
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

	return layers.CreateCenteredLayer(inputBox, m.uiState.Width(), m.uiState.Height())
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

// viewDiscardConfirm renders the discard confirmation dialog with context-aware message
func (m Model) viewDiscardConfirm() string {
	ctx := m.uiState.DiscardContext()
	if ctx == nil {
		return ""
	}

	// Use context message for personalized prompt
	confirmBox := DeleteConfirmBoxStyle.
		Width(50).
		Render(fmt.Sprintf("%s\n\n[y]es  [n]o", ctx.Message))

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		confirmBox,
	)
}

// renderHelpLayer renders the keyboard shortcuts help screen as a layer
func (m Model) renderHelpLayer() *lipgloss.Layer {
	helpBox := HelpBoxStyle.
		Width(50).
		Render(m.generateHelpText())

	return layers.CreateCenteredLayer(helpBox, m.uiState.Width(), m.uiState.Height())
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
  %s     Edit task details

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

VIEWS
  %s     Toggle list/kanban view
  %s     Change task status (list view)
  %s     Cycle sort options (list view)

OTHER
  %s     Show this help screen
  %s     Quit application

Press any key to close`,
		km.AddTask, km.EditTask, km.DeleteTask,
		km.MoveTaskLeft, km.MoveTaskRight,
		km.MoveTaskUp, km.MoveTaskDown,
		formatKey(km.ViewTask),
		km.CreateColumn, km.RenameColumn, km.DeleteColumn,
		km.PrevColumn, km.NextColumn,
		km.PrevTask, km.NextTask,
		km.ScrollViewportLeft, km.ScrollViewportRight,
		km.NextProject, km.PrevProject,
		km.ToggleView, km.ChangeStatus, km.SortList,
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
			m.uiState.Width()*3/4-8,
		)
	} else {
		// Show label list (use filtered items from state)
		pickerContent = RenderLabelPicker(
			m.getFilteredLabelPickerItems(),
			m.labelPickerState.Cursor,
			m.labelPickerState.Filter,
			true, // show create option
			m.uiState.Width()*3/4-8,
			m.uiState.Height()*3/4-4,
		)
	}

	// Wrap in styled container - use different style for create mode
	var pickerBox string
	if m.labelPickerState.CreateMode {
		pickerBox = LabelPickerCreateBoxStyle.
			Width(m.uiState.Width() * 3 / 4).
			Height(m.uiState.Height() * 3 / 4).
			Render(pickerContent)
	} else {
		pickerBox = LabelPickerBoxStyle.
			Width(m.uiState.Width() * 3 / 4).
			Height(m.uiState.Height() * 3 / 4).
			Render(pickerContent)
	}

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewParentPicker renders the parent task picker modal.
// Parent tasks are tasks that depend on (block on) the current task.
// The picker displays all tasks in the project with checkboxes indicating current selections.
func (m Model) viewParentPicker() string {
	pickerContent := RenderTaskPicker(
		m.parentPickerState.GetFilteredItems(),
		m.parentPickerState.Cursor,
		m.parentPickerState.Filter,
		"Parent Issues",
		m.uiState.Width()*3/4-8,
		m.uiState.Height()*3/4-4,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := LabelPickerBoxStyle.
		Width(m.uiState.Width() * 3 / 4).
		Height(m.uiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewChildPicker renders the child task picker modal.
// Child tasks are tasks that the current task depends on (must be completed first).
// The picker displays all tasks in the project with checkboxes indicating current selections.
func (m Model) viewChildPicker() string {
	pickerContent := RenderTaskPicker(
		m.childPickerState.GetFilteredItems(),
		m.childPickerState.Cursor,
		m.childPickerState.Filter,
		"Child Issues",
		m.uiState.Width()*3/4-8,
		m.uiState.Height()*3/4-4,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := LabelPickerBoxStyle.
		Width(m.uiState.Width() * 3 / 4).
		Height(m.uiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewPriorityPicker renders the priority picker popup
func (m Model) viewPriorityPicker() string {
	pickerContent := RenderPriorityPicker(
		GetPriorityOptions(),
		m.priorityPickerState.SelectedPriorityID(),
		m.priorityPickerState.Cursor(),
		m.uiState.Width()*3/4-8,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := LabelPickerBoxStyle.
		Width(m.uiState.Width() * 3 / 4).
		Height(m.uiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewKanbanBoard renders the main kanban board (normal mode)
func (m Model) viewKanbanBoard() string {
	// Check if list view is active
	if m.listViewState.IsListView() {
		return m.viewListView()
	}

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

	// Calculate fixed content height using shared method
	columnHeight := m.uiState.ContentHeight()

	// Render only visible columns
	var columns []string
	for i, col := range visibleColumns {
		// Calculate global index for selection check
		globalIndex := m.uiState.ViewportOffset() + i

		// Safe map access with defensive check
		tasks, ok := m.appState.Tasks()[col.ID]
		if !ok {
			tasks = []*models.TaskSummary{}
		}

		// Determine selection state for this column
		isSelected := (globalIndex == m.uiState.SelectedColumn())

		// Determine which task is selected (only for the selected column)
		selectedTaskIdx := -1
		if isSelected {
			selectedTaskIdx = m.uiState.SelectedTask()
		}

		// Get scroll offset for this column
		scrollOffset := m.uiState.TaskScrollOffset(col.ID)

		columns = append(columns, RenderColumn(col, tasks, isSelected, selectedTaskIdx, columnHeight, scrollOffset))
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
		Width:       m.uiState.Width(),
		SearchMode:  m.uiState.Mode() == state.SearchMode || m.searchState.IsActive,
		SearchQuery: m.searchState.Query,
	})

	// Build content (everything except footer)
	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, "", board, "")

	// Constrain content to fit terminal height, leaving room for footer
	contentLines := strings.Split(content, "\n")
	maxContentLines := m.uiState.Height() - 1 // Reserve 1 line for footer
	if maxContentLines < 1 {
		maxContentLines = 1
	}
	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}
	constrainedContent := strings.Join(contentLines, "\n")

	// Build base view with constrained content and footer always visible
	baseView := constrainedContent + "\n" + footer

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

// viewListView renders the list/table view of all tasks.
func (m Model) viewListView() string {
	// Build rows from all tasks across columns (with sorting applied)
	rows := m.buildListViewRows()

	// Calculate fixed content height using shared method
	listHeight := m.uiState.ContentHeight()

	// Render tab bar (same as kanban)
	var projectTabs []string
	for _, project := range m.appState.Projects() {
		projectTabs = append(projectTabs, project.Name)
	}
	if len(projectTabs) == 0 {
		projectTabs = []string{"No Projects"}
	}
	tabBar := RenderTabs(projectTabs, m.appState.SelectedProject(), m.uiState.Width())

	// Render list content with sort indicator
	listContent := RenderListView(
		rows,
		m.listViewState.SelectedRow(),
		m.listViewState.ScrollOffset(),
		m.listViewState.SortField(),
		m.listViewState.SortOrder(),
		m.uiState.Width(),
		listHeight,
	)

	// Render footer
	footer := components.RenderStatusBar(components.StatusBarProps{
		Width:       m.uiState.Width(),
		SearchMode:  m.uiState.Mode() == state.SearchMode || m.searchState.IsActive,
		SearchQuery: m.searchState.Query,
	})

	// Build content (everything except footer)
	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, "", listContent, "")

	// Constrain content to fit terminal height, leaving room for footer
	contentLines := strings.Split(content, "\n")
	maxContentLines := m.uiState.Height() - 1 // Reserve 1 line for footer
	if maxContentLines < 1 {
		maxContentLines = 1
	}
	if len(contentLines) > maxContentLines {
		contentLines = contentLines[:maxContentLines]
	}
	constrainedContent := strings.Join(contentLines, "\n")

	// Build base view with constrained content and footer always visible
	baseView := constrainedContent + "\n" + footer

	// Add notifications if any
	if !m.notificationState.HasAny() {
		return baseView
	}

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(baseView),
	}
	notificationLayers := m.notificationState.GetLayers(notifications.RenderFromState)
	layers = append(layers, notificationLayers...)
	canvas := lipgloss.NewCanvas(layers...)
	return canvas.Render()
}

// viewStatusPicker renders the status/column selection picker.
func (m Model) viewStatusPicker() string {
	var items []string
	columns := m.statusPickerState.Columns()
	cursor := m.statusPickerState.Cursor()

	for i, col := range columns {
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		items = append(items, prefix+col.Name)
	}

	content := "Select Status:\n\n" + strings.Join(items, "\n") + "\n\nEnter: confirm  Esc: cancel"

	// Wrap in styled container
	pickerBox := LabelPickerBoxStyle.
		Width(40).
		Height(len(columns) + 6).
		Render(content)

	return lipgloss.Place(
		m.uiState.Width(), m.uiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// renderFormTitleDescriptionZone renders the top-left zone with title and description fields
func (m Model) renderFormTitleDescriptionZone(width, height int) string {
	if m.formState.TicketForm == nil {
		return ""
	}

	// Render the form view (which includes title and description)
	formView := m.formState.TicketForm.View()

	style := lipgloss.NewStyle().
		Width(width).
		Height(height)

	return style.Render(formView)
}

// renderFormMetadataZone renders the right column with metadata
func (m Model) renderFormMetadataZone(width, height int) string {
	var parts []string

	labelHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle))

	// Get current timestamps - for create mode, show placeholders
	var createdStr, updatedStr string
	if m.formState.EditingTaskID == 0 {
		createdStr = subtleStyle.Render("(not created yet)")
		updatedStr = subtleStyle.Render("(not created yet)")
	} else {
		// In edit mode, show actual timestamps from FormState
		createdStr = m.formState.FormCreatedAt.Format("Jan 2, 2006 3:04 PM")
		updatedStr = m.formState.FormUpdatedAt.Format("Jan 2, 2006 3:04 PM")
	}

	// Edited indicator (unsaved changes)
	parts = append(parts, labelHeaderStyle.Render("Status"))
	if m.formState.HasTicketFormChanges() {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Highlight))
		parts = append(parts, warningStyle.Render("● Unsaved Changes"))
	} else {
		parts = append(parts, subtleStyle.Render("○ No Changes"))
	}
	parts = append(parts, "")

	// Type section
	parts = append(parts, labelHeaderStyle.Render("Type"))
	if m.formState.FormTypeDescription != "" {
		parts = append(parts, m.formState.FormTypeDescription)
	} else {
		parts = append(parts, subtleStyle.Render("task"))
	}
	parts = append(parts, "")

	// Priority section
	parts = append(parts, labelHeaderStyle.Render("Priority"))
	if m.formState.FormPriorityDescription != "" && m.formState.FormPriorityColor != "" {
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.formState.FormPriorityColor))
		parts = append(parts, priorityStyle.Render(m.formState.FormPriorityDescription))
	} else {
		// Default to medium priority if not set
		priorityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EAB308"))
		parts = append(parts, priorityStyle.Render("medium"))
	}
	parts = append(parts, "")

	// Created timestamp
	parts = append(parts, labelHeaderStyle.Render("Created"))
	parts = append(parts, createdStr)
	parts = append(parts, "")

	// Updated timestamp
	parts = append(parts, labelHeaderStyle.Render("Updated"))
	parts = append(parts, updatedStr)
	parts = append(parts, "")

	// Labels section
	parts = append(parts, labelHeaderStyle.Render("Labels"))
	if len(m.formState.FormLabelIDs) == 0 {
		parts = append(parts, subtleStyle.Render("No labels"))
	} else {
		// Get label objects from IDs
		labelMap := make(map[int]*models.Label)
		for _, label := range m.appState.Labels() {
			labelMap[label.ID] = label
		}

		for _, labelID := range m.formState.FormLabelIDs {
			if label, ok := labelMap[labelID]; ok {
				parts = append(parts, components.RenderLabelChip(label, ""))
			}
		}
	}
	parts = append(parts, "")

	content := strings.Join(parts, "\n")

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{
			Left: "│",
		}).
		BorderForeground(lipgloss.Color(theme.Subtle))

	return style.Render(content)
}

// renderFormAssociationsZone renders the bottom-left zone with parent and child tasks
func (m Model) renderFormAssociationsZone(width, height int) string {
	var parts []string

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Bold(true)

	subtleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Subtle)).
		Italic(true)

	taskStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Normal))

	// Parent Tasks section
	parts = append(parts, headerStyle.Render("Parent Tasks"))
	if len(m.formState.FormParentRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No Parent Tasks Found"))
	} else {
		for _, parent := range m.formState.FormParentRefs {
			taskLine := fmt.Sprintf("#%d - %s", parent.TicketNumber, parent.Title)
			parts = append(parts, taskStyle.Render(taskLine))
		}
	}
	parts = append(parts, "")

	// Child Tasks section
	parts = append(parts, headerStyle.Render("Child Tasks"))
	if len(m.formState.FormChildRefs) == 0 {
		parts = append(parts, subtleStyle.Render("No Child Tasks Found"))
	} else {
		for _, child := range m.formState.FormChildRefs {
			taskLine := fmt.Sprintf("#%d - %s", child.TicketNumber, child.Title)
			parts = append(parts, taskStyle.Render(taskLine))
		}
	}

	content := strings.Join(parts, "\n")

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1).
		BorderTop(true).
		BorderStyle(lipgloss.Border{
			Top: "─",
		}).
		BorderForeground(lipgloss.Color(theme.Subtle))

	return style.Render(content)
}
