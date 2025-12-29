package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/layers"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// renderTicketFormLayer renders the ticket creation/edit form modal as a layer
func (m Model) renderTicketFormLayer() *lipgloss.Layer {
	if m.FormState.TicketForm == nil {
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
	// Calculate layer dimensions (80% of screen, same as ticket form)
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

// renderNoteFormLayer renders the note creation/edit form modal as a layer
func (m Model) renderNoteFormLayer() *lipgloss.Layer {
	if m.FormState.CommentForm == nil {
		return nil
	}

	formView := m.FormState.CommentForm.View()

	// Determine title based on mode
	var title string
	if m.FormState.EditingCommentID == 0 {
		title = "New Note"
	} else {
		title = "Edit Note"
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
  Ctrl+N          Create new note/comment
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
