package tui

import (
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// viewLabelPicker renders the label picker modal
func (m Model) viewLabelPicker() string {
	// Render the label picker content
	var pickerContent string
	if m.LabelPickerState.CreateMode {
		// Show color picker
		pickerContent = renderers.RenderLabelColorPicker(
			renderers.GetDefaultLabelColors(),
			m.LabelPickerState.ColorIdx,
			m.FormState.FormLabelName,
			m.UiState.Width()*3/4-8,
		)
	} else {
		// Show label list (use filtered items from state)
		pickerContent = renderers.RenderLabelPicker(
			m.getFilteredLabelPickerItems(),
			m.LabelPickerState.Cursor,
			m.LabelPickerState.Filter,
			true, // show create option
			m.UiState.Width()*3/4-8,
			m.UiState.Height()*3/4-4,
		)
	}

	// Wrap in styled container - use different style for create mode
	var pickerBox string
	if m.LabelPickerState.CreateMode {
		pickerBox = components.LabelPickerCreateBoxStyle.
			Width(m.UiState.Width() * 3 / 4).
			Height(m.UiState.Height() * 3 / 4).
			Render(pickerContent)
	} else {
		pickerBox = components.LabelPickerBoxStyle.
			Width(m.UiState.Width() * 3 / 4).
			Height(m.UiState.Height() * 3 / 4).
			Render(pickerContent)
	}

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewParentPicker renders the parent task picker modal.
// Parent tasks are tasks that depend on (block on) the current task.
// The picker displays all tasks in the project with checkboxes indicating current selections.
func (m Model) viewParentPicker() string {
	pickerContent := renderers.RenderTaskPicker(
		m.ParentPickerState.GetFilteredItems(),
		m.ParentPickerState.Cursor,
		m.ParentPickerState.Filter,
		"Parent Issues",
		m.UiState.Width()*3/4-8,
		m.UiState.Height()*3/4-4,
		true, // isParentPicker
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(m.UiState.Width() * 3 / 4).
		Height(m.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewChildPicker renders the child task picker modal.
// Child tasks are tasks that the current task depends on (must be completed first).
// The picker displays all tasks in the project with checkboxes indicating current selections.
func (m Model) viewChildPicker() string {
	pickerContent := renderers.RenderTaskPicker(
		m.ChildPickerState.GetFilteredItems(),
		m.ChildPickerState.Cursor,
		m.ChildPickerState.Filter,
		"Child Issues",
		m.UiState.Width()*3/4-8,
		m.UiState.Height()*3/4-4,
		false, // isParentPicker
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(m.UiState.Width() * 3 / 4).
		Height(m.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewPriorityPicker renders the priority picker popup
func (m Model) viewPriorityPicker() string {
	pickerContent := renderers.RenderPriorityPicker(
		renderers.GetPriorityOptions(),
		m.PriorityPickerState.SelectedPriorityID(),
		m.PriorityPickerState.Cursor(),
		m.UiState.Width()*3/4-8,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(m.UiState.Width() * 3 / 4).
		Height(m.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewTypePicker renders the type picker popup
func (m Model) viewTypePicker() string {
	pickerContent := renderers.RenderTypePicker(
		renderers.GetTypeOptions(),
		m.TypePickerState.SelectedTypeID(),
		m.TypePickerState.Cursor(),
		m.UiState.Width()*3/4-8,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(m.UiState.Width() * 3 / 4).
		Height(m.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewRelationTypePicker renders the relation type picker popup
func (m Model) viewRelationTypePicker() string {
	// Determine picker type based on return mode
	pickerType := "parent"
	if m.RelationTypePickerState.ReturnMode() == state.ChildPickerMode {
		pickerType = "child"
	}

	pickerContent := renderers.RenderRelationTypePicker(
		renderers.GetRelationTypeOptions(),
		m.RelationTypePickerState.SelectedRelationTypeID(),
		m.RelationTypePickerState.Cursor(),
		m.UiState.Width()*3/4-8,
		pickerType,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(m.UiState.Width() * 3 / 4).
		Height(m.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewStatusPicker renders the status/column selection picker.
func (m Model) viewStatusPicker() string {
	var items []string
	columns := m.StatusPickerState.Columns()
	cursor := m.StatusPickerState.Cursor()

	for i, col := range columns {
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		items = append(items, prefix+col.Name)
	}

	content := "Select Status:\n\n" + lipgloss.JoinVertical(lipgloss.Left, items...) + "\n\nEnter: confirm  Esc: cancel"

	// Wrap in styled container
	pickerBox := components.LabelPickerBoxStyle.
		Width(40).
		Height(len(columns) + 6).
		Render(content)

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// viewNoteEditor renders the note editor interface for managing task notes.
func (m Model) viewNoteEditor() string {
	var content string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7aa2f7")).
		Render("Notes Editor")

	// Help text - always show navigation mode since editing is done in form
	helpText := "j/k: navigate • Enter: edit • n: new note • Del: delete • Esc: close"

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565f89")).
		Render(helpText)

	// Build notes list
	var notesList []string

	if len(m.NoteState.Items) == 0 {
		// No notes
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89")).
			Italic(true).
			Render("No notes yet. Press n to add one.")
		notesList = append(notesList, emptyMsg)
	} else {
		// Display notes
		for i, item := range m.NoteState.Items {
			timestamp := item.Comment.CreatedAt.Format("Jan 2 15:04")

			// Display note
			prefix := "  "
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#c0caf5"))

			if i == m.NoteState.Cursor {
				prefix = "▶ "
				style = style.Bold(true).Foreground(lipgloss.Color("#7aa2f7"))
			}

			timeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#565f89"))

			noteLine := prefix + style.Render(item.Comment.Message) + " " + timeStyle.Render("["+timestamp+"]")
			notesList = append(notesList, noteLine)
		}
	}

	// Join all content
	content = lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		help,
		"",
		lipgloss.JoinVertical(lipgloss.Left, notesList...),
	)

	// Wrap in styled container
	pickerBox := components.LabelPickerBoxStyle.
		Width(m.UiState.Width() * 3 / 4).
		Height(m.UiState.Height() * 3 / 4).
		Render(content)

	return lipgloss.Place(
		m.UiState.Width(), m.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}
