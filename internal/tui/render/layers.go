package render

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/layers"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderTicketFormLayer renders the ticket creation/edit form modal as a layer
func (w *Wrapper) RenderTicketFormLayer() *lipgloss.Layer {
	if w.FormState.TicketForm == nil {
		return nil
	}

	// Calculate layer dimensions (80% of screen)
	layerWidth := w.UiState.Width() * 4 / 5
	layerHeight := w.UiState.Height() * 4 / 5

	// Calculate zone dimensions
	leftColumnWidth := layerWidth * 7 / 10  // 70% of layer width
	rightColumnWidth := layerWidth * 3 / 10 // 30% of layer width
	topHeight := layerHeight * 6 / 10       // 60% of layer height
	bottomHeight := layerHeight * 4 / 10    // 40% of layer height

	// Render the three zones
	topLeftZone := w.renderFormTitleDescriptionZone(leftColumnWidth, topHeight)
	bottomLeftZone := w.renderFormAssociationsZone(leftColumnWidth, bottomHeight)
	rightZone := w.renderFormMetadataZone(rightColumnWidth, layerHeight)

	// Compose left column (top + bottom)
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, topLeftZone, bottomLeftZone)

	// Compose full content (left + right)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightZone)

	// Add form title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Highlight))
	var formTitle string
	if w.FormState.EditingTaskID == 0 {
		formTitle = titleStyle.Render("Create New Task")
	} else {
		formTitle = titleStyle.Render("Edit Task")
	}

	// Add help text for shortcuts
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Subtle))
	helpText := helpStyle.Render("Ctrl+L: edit labels  Ctrl+P: edit parents  Ctrl+C: edit children  Ctrl+R: edit priority Ctrl+T edit type")

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

	return layers.CreateCenteredLayer(formBox, w.UiState.Width(), w.UiState.Height())
}

// RenderProjectFormLayer renders the project creation form modal as a layer
func (w *Wrapper) RenderProjectFormLayer() *lipgloss.Layer {
	if w.FormState.ProjectForm == nil {
		return nil
	}

	formView := w.FormState.ProjectForm.View()

	// Wrap form in a styled container with green border for creation
	formBox := components.ProjectFormBoxStyle.
		Width(w.UiState.Width() * 3 / 4).
		Height(w.UiState.Height() / 3).
		Render("New Project\n\n" + formView)

	return layers.CreateCenteredLayer(formBox, w.UiState.Width(), w.UiState.Height())
}

// RenderColumnInputLayer renders the column name input dialog (create or edit mode) as a layer
func (w *Wrapper) RenderColumnInputLayer() *lipgloss.Layer {
	var inputBox string
	if w.UiState.Mode() == state.AddColumnMode {
		inputBox = components.CreateInputBoxStyle.
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", w.InputState.Prompt, w.InputState.Buffer))
	} else {
		// EditColumnMode
		inputBox = components.EditInputBoxStyle.
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", w.InputState.Prompt, w.InputState.Buffer))
	}

	return layers.CreateCenteredLayer(inputBox, w.UiState.Width(), w.UiState.Height())
}

// RenderHelpLayer renders the keyboard shortcuts help screen as a layer
func (w *Wrapper) RenderHelpLayer() *lipgloss.Layer {
	helpBox := components.HelpBoxStyle.
		Width(50).
		Render(w.generateHelpText())

	return layers.CreateCenteredLayer(helpBox, w.UiState.Width(), w.UiState.Height())
}

// generateHelpText creates help text based on current key mappings
func (w *Wrapper) generateHelpText() string {
	km := w.Config.KeyMappings
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
