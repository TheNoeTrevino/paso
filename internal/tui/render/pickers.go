package render

import (
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ViewLabelPicker renders the label picker modal
func ViewLabelPicker(m *tui.Model) string {
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
			modelops.GetFilteredLabelPickerItems(m),
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

// ViewParentPicker renders the parent task picker modal.
// Parent tasks are tasks that depend on (block on) the current task.
// The picker displays all tasks in the project with checkboxes indicating current selections.
func ViewParentPicker(m *tui.Model) string {
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

// ViewChildPicker renders the child task picker modal.
// Child tasks are tasks that the current task depends on (must be completed first).
// The picker displays all tasks in the project with checkboxes indicating current selections.
func ViewChildPicker(m *tui.Model) string {
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

// ViewPriorityPicker renders the priority picker popup
func ViewPriorityPicker(m *tui.Model) string {
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

// ViewTypePicker renders the type picker popup
func ViewTypePicker(m *tui.Model) string {
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

// ViewRelationTypePicker renders the relation type picker popup
func ViewRelationTypePicker(m *tui.Model) string {
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

// ViewStatusPicker renders the status/column selection picker.
func ViewStatusPicker(m *tui.Model) string {
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
