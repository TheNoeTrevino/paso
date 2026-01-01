package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/layers"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderTaskFormLayer renders the task creation/edit form modal as a layer
func (m Model) renderTaskFormLayer() *lipgloss.Layer {
	if m.FormState.TaskForm == nil {
		return nil
	}

	// Calculate layer dimensions (80% of screen)
	layerWidth := m.UiState.Width() * 8 / 10
	layerHeight := m.UiState.Height() * 8 / 10

	// Account for chrome: border (2) + padding (2) + title (1) + blanks (1) = 6 lines
	chromeHeight := 6
	innerHeight := layerHeight - chromeHeight

	// Calculate zone dimensions based on requirements:
	// - Top left (title/desc): 60% width, 70% height
	// - Bottom left (comments): 60% width, 30% height
	// - Right (metadata): 40% width, 100% height
	leftColumnWidth := layerWidth * 6 / 10   // 60% of layer width
	rightColumnWidth := layerWidth * 4 / 10  // 40% of layer width
	topLeftHeight := innerHeight * 7 / 10    // 70% of inner height
	bottomLeftHeight := innerHeight * 3 / 10 // 30% of inner height
	rightColumnHeight := innerHeight         // 100% of inner height

	// Calculate dynamic description field height based on available space
	// Chrome overhead: Title field (~3) + Confirmation (~3) + spacing (~3) = 9 lines
	const (
		descChromeOverhead = 9
		minDescLines       = 5
		maxDescLines       = 15
	)

	descriptionLines := topLeftHeight - descChromeOverhead
	if descriptionLines < minDescLines {
		descriptionLines = minDescLines
	}
	if descriptionLines > maxDescLines {
		descriptionLines = maxDescLines
	}

	// Store in FormState for use during form creation
	m.FormState.CalculatedDescriptionLines = descriptionLines

	// Render the three zones
	topLeftZone := m.renderFormTitleDescriptionZone(leftColumnWidth, topLeftHeight)
	bottomLeftZone := m.renderFormCommentsPreview(leftColumnWidth, bottomLeftHeight)
	rightZone := m.renderFormMetadataZone(rightColumnWidth, rightColumnHeight)

	// Compose left column (top + bottom)
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, topLeftZone, bottomLeftZone)

	// Compose full content (left column + right metadata)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightZone)

	// Add form title with help hint
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	helpHintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))

	var formTitle string
	if m.FormState.EditingTaskID == 0 {
		formTitle = titleStyle.Render("Create New Task")
	} else {
		formTitle = titleStyle.Render("Edit Task")
	}

	titleWithHint := lipgloss.JoinHorizontal(
		lipgloss.Left,
		formTitle,
		"  ",
		helpHintStyle.Render("|"),
		"  ",
		helpHintStyle.Render("Ctrl+H: help"),
	)

	// Combine title + content
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleWithHint,
		"",
		content,
	)

	// Wrap in form box style
	formBox := components.FormBoxStyle.
		Width(layerWidth).
		Height(layerHeight).
		Render(fullContent)

	return layers.CreateCenteredLayer(formBox, m.UiState.Width(), m.UiState.Height())
}

// renderProjectFormLayer renders the project creation form modal as a layer
func (m Model) renderProjectFormLayer() *lipgloss.Layer {
	if m.FormState.ProjectForm == nil {
		return nil
	}

	formView := m.FormState.ProjectForm.View()

	// Wrap form in a styled container with green border for creation
	formBox := components.ProjectFormBoxStyle.
		Width(m.UiState.Width() * 3 / 4).
		Height(m.UiState.Height() / 3).
		Render("New Project\n\n" + formView)

	return layers.CreateCenteredLayer(formBox, m.UiState.Width(), m.UiState.Height())
}

// renderColumnFormLayer renders the column creation/rename form modal as a layer
func (m Model) renderColumnFormLayer() *lipgloss.Layer {
	if m.FormState.ColumnForm == nil {
		return nil
	}

	formView := m.FormState.ColumnForm.View()

	// Determine title based on mode
	var title string
	if m.UiState.Mode() == state.AddColumnFormMode {
		title = "New Column"
	} else {
		title = "Rename Column"
	}

	// Wrap form in a styled container
	formBox := components.CreateInputBoxStyle.
		Width(50).
		Render(title + "\n\n" + formView)

	return layers.CreateCenteredLayer(formBox, m.UiState.Width(), m.UiState.Height())
}

// renderHelpLayer renders the keyboard shortcuts help screen as a layer
func (m Model) renderHelpLayer() *lipgloss.Layer {
	helpBox := components.HelpBoxStyle.
		Width(50).
		Render(m.generateHelpText())

	return layers.CreateCenteredLayer(helpBox, m.UiState.Width(), m.UiState.Height())
}

// generateHelpText creates help text based on current key mappings
func (m Model) generateHelpText() string {
	km := m.Config.KeyMappings
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

PROJECTS
  %s     Switch to previous project
  %s     Switch to next project
  %s     Create new project

VIEW
  %s     Toggle between kanban and list view
  %s     Change status (list view)
  %s     Toggle sort order (list view)
  /         Search tasks

OTHER
  %s     Show this help
  %s     Quit

Press any key to close`,
		km.AddTask,
		km.EditTask,
		km.DeleteTask,
		km.MoveTaskLeft,
		km.MoveTaskRight,
		km.MoveTaskUp,
		km.MoveTaskDown,
		km.ViewTask,
		km.CreateColumn,
		km.RenameColumn,
		km.DeleteColumn,
		km.PrevColumn,
		km.NextColumn,
		km.PrevTask,
		km.NextTask,
		km.ScrollViewportLeft,
		km.ScrollViewportRight,
		km.PrevProject,
		km.NextProject,
		km.CreateProject,
		km.ToggleView,
		km.ChangeStatus,
		km.SortList,
		km.ShowHelp,
		km.Quit,
	)
}

// renderCommentsViewLayer renders the comments view modal as a full-screen layer
func (m Model) renderCommentsViewLayer() *lipgloss.Layer {
	// Calculate layer dimensions (80% of screen, same as task form)
	layerWidth := m.UiState.Width() * 8 / 10
	layerHeight := m.UiState.Height() * 8 / 10

	// Render comments view content
	content := m.renderCommentsViewContent(layerWidth, layerHeight)

	// Wrap in styled box
	commentsBox := components.HelpBoxStyle.
		Width(layerWidth).
		Height(layerHeight).
		Render(content)

	return layers.CreateCenteredLayer(commentsBox, m.UiState.Width(), m.UiState.Height())
}

// renderCommentFormLayer renders the comment creation/edit form modal as a layer
func (m Model) renderCommentFormLayer() *lipgloss.Layer {
	if m.FormState.CommentForm == nil {
		return nil
	}

	formView := m.FormState.CommentForm.View()

	// Determine title based on mode
	var title string
	if m.FormState.EditingCommentID == 0 {
		title = "New Comment"
	} else {
		title = "Edit Comment"
	}

	// Wrap form in a styled container
	formBox := components.CreateInputBoxStyle.
		Width(m.UiState.Width() * 3 / 4).
		Height(m.UiState.Height() * 2 / 3).
		Render(title + "\n\n" + formView)

	return layers.CreateCenteredLayer(formBox, m.UiState.Width(), m.UiState.Height())
}

// renderTaskFormHelpLayer renders the task form keyboard shortcuts help screen as a layer
func (m Model) renderTaskFormHelpLayer() *lipgloss.Layer {
	helpContent := `TASK FORM - Keyboard Shortcuts

FORM NAVIGATION
  Tab             Navigate between form fields
  Shift+Tab       Navigate backwards
  Ctrl+S          Save task and close form
  Esc             Close form (will prompt if unsaved)

TEXT EDITING (Title/Description)
  Shift+Enter     New line
  Alt+Enter       New line
  Ctrl+J          New line
  Ctrl+E          Open editor
  Enter           Next field

COMMENTS SECTION
  Ctrl+↓          Focus comments section
  Down            Auto-focus comments (when not focused)
  ↑↓              Scroll comments (when focused)
  Mouse wheel     Scroll comments (when focused)
  Tab/Shift+Tab   Return to form fields

QUICK ACTIONS
  Ctrl+N          Create new comment
  Ctrl+L          Manage labels
  Ctrl+P          Select parent tasks
  Ctrl+C          Select child tasks
  Ctrl+R          Change priority
  Ctrl+T          Change task type

HELP
  Ctrl+/          Toggle this help menu
  Esc             Close help menu

Press Ctrl+/ or Esc to close`

	helpBox := components.HelpBoxStyle.
		Width(m.UiState.Width() * 3 / 8).
		Render(helpContent)

	return layers.CreateCenteredLayer(helpBox, m.UiState.Width(), m.UiState.Height())
}

// renderLabelPickerLayer renders the label picker modal as a layer
func (m Model) renderLabelPickerLayer() *lipgloss.Layer {
	var pickerContent string
	var pickerWidth, pickerHeight int

	if m.LabelPickerState.CreateMode {
		// CreateMode: Show color picker
		// Use dynamic dimensions for color picker (similar to normal mode but with reasonable defaults)
		pickerWidth, pickerHeight = layers.CalculatePickerDimensions(
			10, // reasonable default for color picker (shows ~10 colors)
			false,
			m.UiState.Width(),
			m.UiState.Height(),
			40,
			60,
		)

		// Render color picker content (account for border + padding: -8 width)
		pickerContent = renderers.RenderLabelColorPicker(
			renderers.GetDefaultLabelColors(),
			m.LabelPickerState.ColorIdx,
			m.FormState.FormLabelName,
			pickerWidth-8,
		)
	} else {
		// Normal mode: Show label list
		filteredItems := m.getFilteredLabelPickerItems()
		hasFilter := m.LabelPickerState.Filter != ""

		// Calculate dynamic dimensions based on filtered item count
		// Add 1 for the "Create new label" option
		itemCount := len(filteredItems) + 1
		pickerWidth, pickerHeight = layers.CalculatePickerDimensions(
			itemCount,
			hasFilter,
			m.UiState.Width(),
			m.UiState.Height(),
			40,
			60,
		)

		// Render label picker content (account for border + padding: -8 width, -4 height)
		pickerContent = renderers.RenderLabelPicker(
			filteredItems,
			m.LabelPickerState.Cursor,
			m.LabelPickerState.Filter,
			true, // show create option
			pickerWidth-8,
			pickerHeight-4,
		)
	}

	// Wrap in styled container - use different style for create mode
	var pickerBox string
	if m.LabelPickerState.CreateMode {
		pickerBox = components.LabelPickerCreateBoxStyle.
			Width(pickerWidth).
			Height(pickerHeight).
			Render(pickerContent)
	} else {
		pickerBox = components.LabelPickerBoxStyle.
			Width(pickerWidth).
			Height(pickerHeight).
			Render(pickerContent)
	}

	return layers.CreateCenteredLayer(pickerBox, m.UiState.Width(), m.UiState.Height())
}

// renderParentPickerLayer renders the parent task picker modal as a layer
func (m Model) renderParentPickerLayer() *lipgloss.Layer {
	// Get filtered items from state
	filteredItems := m.ParentPickerState.GetFilteredItems()
	hasFilter := m.ParentPickerState.Filter != ""

	// Calculate dynamic dimensions based on filtered item count
	itemCount := len(filteredItems)
	pickerWidth, pickerHeight := layers.CalculatePickerDimensions(
		itemCount,
		hasFilter,
		m.UiState.Width(),
		m.UiState.Height(),
		50,
		70,
	)

	// Render parent picker content (account for border + padding: -8 width, -4 height)
	pickerContent := renderers.RenderTaskPicker(
		filteredItems,
		m.ParentPickerState.Cursor,
		m.ParentPickerState.Filter,
		"Parent Issues",
		pickerWidth-8,
		pickerHeight-4,
		true, // isParentPicker
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(pickerWidth).
		Height(pickerHeight).
		Render(pickerContent)

	return layers.CreateCenteredLayer(pickerBox, m.UiState.Width(), m.UiState.Height())
}

// renderChildPickerLayer renders the child task picker modal as a layer
func (m Model) renderChildPickerLayer() *lipgloss.Layer {
	// Get filtered items from state
	filteredItems := m.ChildPickerState.GetFilteredItems()
	hasFilter := m.ChildPickerState.Filter != ""

	// Calculate dynamic dimensions based on filtered item count
	itemCount := len(filteredItems)
	pickerWidth, pickerHeight := layers.CalculatePickerDimensions(
		itemCount,
		hasFilter,
		m.UiState.Width(),
		m.UiState.Height(),
		50,
		70,
	)

	// Render child picker content (account for border + padding: -8 width, -4 height)
	pickerContent := renderers.RenderTaskPicker(
		filteredItems,
		m.ChildPickerState.Cursor,
		m.ChildPickerState.Filter,
		"Child Issues",
		pickerWidth-8,
		pickerHeight-4,
		false, // isParentPicker
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(pickerWidth).
		Height(pickerHeight).
		Render(pickerContent)

	return layers.CreateCenteredLayer(pickerBox, m.UiState.Width(), m.UiState.Height())
}

// renderPriorityPickerLayer renders the priority picker modal as a layer
func (m Model) renderPriorityPickerLayer() *lipgloss.Layer {
	// Fixed small size for priority picker (5 options)
	pickerWidth := 40
	pickerHeight := 12 // 5 options + chrome (title, footer, padding, border)

	pickerContent := renderers.RenderPriorityPicker(
		renderers.GetPriorityOptions(),
		m.PriorityPickerState.SelectedPriorityID(),
		m.PriorityPickerState.Cursor(),
		pickerWidth-8, // Account for padding/border
	)

	// Wrap in styled container
	pickerBox := components.LabelPickerBoxStyle.
		Width(pickerWidth).
		Height(pickerHeight).
		Render(pickerContent)

	return layers.CreateCenteredLayer(pickerBox, m.UiState.Width(), m.UiState.Height())
}

// renderTypePickerLayer renders the type picker modal as a layer
func (m Model) renderTypePickerLayer() *lipgloss.Layer {
	// Fixed small size for type picker (2 options: task/feature)
	pickerWidth := 40
	pickerHeight := 9 // 2 options + chrome (title, footer, padding, border)

	pickerContent := renderers.RenderTypePicker(
		renderers.GetTypeOptions(),
		m.TypePickerState.SelectedTypeID(),
		m.TypePickerState.Cursor(),
		pickerWidth-8, // Account for padding/border
	)

	// Wrap in styled container
	pickerBox := components.LabelPickerBoxStyle.
		Width(pickerWidth).
		Height(pickerHeight).
		Render(pickerContent)

	return layers.CreateCenteredLayer(pickerBox, m.UiState.Width(), m.UiState.Height())
}

// renderRelationTypePickerLayer renders the relation type picker modal as a layer
func (m Model) renderRelationTypePickerLayer() *lipgloss.Layer {
	// Fixed small size for relation type picker (3 options)
	pickerWidth := 45
	pickerHeight := 11 // 3 options + chrome (title, footer, padding, border)

	// Determine picker type based on return mode
	pickerType := "parent"
	if m.RelationTypePickerState.ReturnMode() == state.ChildPickerMode {
		pickerType = "child"
	}

	pickerContent := renderers.RenderRelationTypePicker(
		renderers.GetRelationTypeOptions(),
		m.RelationTypePickerState.SelectedRelationTypeID(),
		m.RelationTypePickerState.Cursor(),
		pickerWidth-8, // Account for padding/border
		pickerType,
	)

	// Wrap in styled container
	pickerBox := components.LabelPickerBoxStyle.
		Width(pickerWidth).
		Height(pickerHeight).
		Render(pickerContent)

	return layers.CreateCenteredLayer(pickerBox, m.UiState.Width(), m.UiState.Height())
}

// renderStatusPickerLayer renders the status/column selection picker modal as a layer
func (m Model) renderStatusPickerLayer() *lipgloss.Layer {
	columns := m.StatusPickerState.Columns()
	cursor := m.StatusPickerState.Cursor()

	// Build column list items
	var items []string
	for i, col := range columns {
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		items = append(items, prefix+col.Name)
	}

	content := "Select Status:\n\n" + lipgloss.JoinVertical(lipgloss.Left, items...) + "\n\nEnter: confirm  Esc: cancel"

	// Fixed width, dynamic height based on column count
	pickerWidth := 40
	pickerHeight := len(columns) + 6 // Columns + chrome (title, spacing, footer)

	// Wrap in styled container
	pickerBox := components.LabelPickerBoxStyle.
		Width(pickerWidth).
		Height(pickerHeight).
		Render(content)

	return layers.CreateCenteredLayer(pickerBox, m.UiState.Width(), m.UiState.Height())
}
