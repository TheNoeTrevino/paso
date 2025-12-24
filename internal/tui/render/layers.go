package render

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/layers"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// RenderTicketFormLayer renders the ticket creation/edit form modal as a layer
func RenderTicketFormLayer(m *tui.Model) *lipgloss.Layer {
	if m.FormState.TicketForm == nil {
		return nil
	}

	// Calculate layer dimensions (80% of screen)
	layerWidth := m.UiState.Width() * 4 / 5
	layerHeight := m.UiState.Height() * 4 / 5

	// Calculate zone dimensions
	leftColumnWidth := layerWidth * 7 / 10  // 70% of layer width
	rightColumnWidth := layerWidth * 3 / 10 // 30% of layer width
	topHeight := layerHeight * 6 / 10       // 60% of layer height
	bottomHeight := layerHeight * 4 / 10    // 40% of layer height

	// Render the three zones
	topLeftZone := renderFormTitleDescriptionZone(m, leftColumnWidth, topHeight)
	bottomLeftZone := renderFormAssociationsZone(m, leftColumnWidth, bottomHeight)
	rightZone := renderFormMetadataZone(m, rightColumnWidth, layerHeight)

	// Compose left column (top + bottom)
	leftColumn := lipgloss.JoinVertical(lipgloss.Top, topLeftZone, bottomLeftZone)

	// Compose full content (left + right)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightZone)

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

	return layers.CreateCenteredLayer(formBox, m.UiState.Width(), m.UiState.Height())
}

// RenderProjectFormLayer renders the project creation form modal as a layer
func RenderProjectFormLayer(m *tui.Model) *lipgloss.Layer {
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

// RenderColumnInputLayer renders the column name input dialog (create or edit mode) as a layer
func RenderColumnInputLayer(m *tui.Model) *lipgloss.Layer {
	var inputBox string
	if m.UiState.Mode() == state.AddColumnMode {
		inputBox = components.CreateInputBoxStyle.
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.InputState.Prompt, m.InputState.Buffer))
	} else {
		// EditColumnMode
		inputBox = components.EditInputBoxStyle.
			Width(50).
			Render(fmt.Sprintf("%s\n> %s_", m.InputState.Prompt, m.InputState.Buffer))
	}

	return layers.CreateCenteredLayer(inputBox, m.UiState.Width(), m.UiState.Height())
}

// RenderHelpLayer renders the keyboard shortcuts help screen as a layer
func RenderHelpLayer(m *tui.Model) *lipgloss.Layer {
	helpBox := components.HelpBoxStyle.
		Width(50).
		Render(generateHelpText(m))

	return layers.CreateCenteredLayer(helpBox, m.UiState.Width(), m.UiState.Height())
}

// generateHelpText creates help text based on current key mappings
func generateHelpText(m *tui.Model) string {
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
