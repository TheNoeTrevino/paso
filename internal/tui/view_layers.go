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

	// Account for chrome: border (2) + padding (2) + title (1) + blanks (2) + help (1) = 8 lines
	chromeHeight := 8
	innerHeight := layerHeight - chromeHeight

	// Calculate zone dimensions based on requirements:
	// - Top left (title/desc): 60% width, 70% height
	// - Top right (metadata): 40% width, 70% height
	// - Bottom (comments): 100% width, 30% height
	leftColumnWidth := layerWidth * 6 / 10  // 60% of layer width
	rightColumnWidth := layerWidth * 4 / 10 // 40% of layer width
	topHeight := innerHeight * 7 / 10       // 70% of inner height
	bottomHeight := innerHeight * 3 / 10    // 30% of inner height

	// Render the three zones
	topLeftZone := m.renderFormTitleDescriptionZone(leftColumnWidth, topHeight)
	topRightZone := m.renderFormMetadataZone(rightColumnWidth, topHeight)
	bottomZone := m.renderFormNotesZone(layerWidth, bottomHeight)

	// Compose top row (left + right)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topLeftZone, topRightZone)

	// Compose full content (top row + bottom comments)
	content := lipgloss.JoinVertical(lipgloss.Top, topRow, bottomZone)

	// Add form title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	var formTitle string
	if m.FormState.EditingTaskID == 0 {
		formTitle = titleStyle.Render("Create New Task")
	} else {
		formTitle = titleStyle.Render("Edit Task")
	}

	// Add help text for shortcuts
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	helpText := helpStyle.Render("Ctrl+L: labels  Ctrl+P: parents  Ctrl+C: children  Ctrl+R: priority  Ctrl+T: type  Ctrl+N: notes")

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
