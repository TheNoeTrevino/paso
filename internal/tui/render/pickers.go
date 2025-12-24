package render

import (
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/components"
	"github.com/thenoetrevino/paso/internal/tui/modelops"
	"github.com/thenoetrevino/paso/internal/tui/renderers"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// ViewLabelPicker renders the label picker modal
func (w *Wrapper) ViewLabelPicker() string {
	// Render the label picker content
	var pickerContent string
	if w.LabelPickerState.CreateMode {
		// Show color picker
		pickerContent = renderers.RenderLabelColorPicker(
			renderers.GetDefaultLabelColors(),
			w.LabelPickerState.ColorIdx,
			w.FormState.FormLabelName,
			w.UiState.Width()*3/4-8,
		)
	} else {
		// Show label list (use filtered items from state)
		ops := modelops.New(w.Model)
		pickerContent = renderers.RenderLabelPicker(
			ops.GetFilteredLabelPickerItems(),
			w.LabelPickerState.Cursor,
			w.LabelPickerState.Filter,
			true, // show create option
			w.UiState.Width()*3/4-8,
			w.UiState.Height()*3/4-4,
		)
	}

	// Wrap in styled container - use different style for create mode
	var pickerBox string
	if w.LabelPickerState.CreateMode {
		pickerBox = components.LabelPickerCreateBoxStyle.
			Width(w.UiState.Width() * 3 / 4).
			Height(w.UiState.Height() * 3 / 4).
			Render(pickerContent)
	} else {
		pickerBox = components.LabelPickerBoxStyle.
			Width(w.UiState.Width() * 3 / 4).
			Height(w.UiState.Height() * 3 / 4).
			Render(pickerContent)
	}

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// ViewParentPicker renders the parent task picker modal.
// Parent tasks are tasks that depend on (block on) the current task.
// The picker displays all tasks in the project with checkboxes indicating current selections.
func (w *Wrapper) ViewParentPicker() string {
	pickerContent := renderers.RenderTaskPicker(
		w.ParentPickerState.GetFilteredItems(),
		w.ParentPickerState.Cursor,
		w.ParentPickerState.Filter,
		"Parent Issues",
		w.UiState.Width()*3/4-8,
		w.UiState.Height()*3/4-4,
		true, // isParentPicker
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(w.UiState.Width() * 3 / 4).
		Height(w.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// ViewChildPicker renders the child task picker modal.
// Child tasks are tasks that the current task depends on (must be completed first).
// The picker displays all tasks in the project with checkboxes indicating current selections.
func (w *Wrapper) ViewChildPicker() string {
	pickerContent := renderers.RenderTaskPicker(
		w.ChildPickerState.GetFilteredItems(),
		w.ChildPickerState.Cursor,
		w.ChildPickerState.Filter,
		"Child Issues",
		w.UiState.Width()*3/4-8,
		w.UiState.Height()*3/4-4,
		false, // isParentPicker
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(w.UiState.Width() * 3 / 4).
		Height(w.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// ViewPriorityPicker renders the priority picker popup
func (w *Wrapper) ViewPriorityPicker() string {
	pickerContent := renderers.RenderPriorityPicker(
		renderers.GetPriorityOptions(),
		w.PriorityPickerState.SelectedPriorityID(),
		w.PriorityPickerState.Cursor(),
		w.UiState.Width()*3/4-8,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(w.UiState.Width() * 3 / 4).
		Height(w.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// ViewTypePicker renders the type picker popup
func (w *Wrapper) ViewTypePicker() string {
	pickerContent := renderers.RenderTypePicker(
		renderers.GetTypeOptions(),
		w.TypePickerState.SelectedTypeID(),
		w.TypePickerState.Cursor(),
		w.UiState.Width()*3/4-8,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(w.UiState.Width() * 3 / 4).
		Height(w.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// ViewRelationTypePicker renders the relation type picker popup
func (w *Wrapper) ViewRelationTypePicker() string {
	// Determine picker type based on return mode
	pickerType := "parent"
	if w.RelationTypePickerState.ReturnMode() == state.ChildPickerMode {
		pickerType = "child"
	}

	pickerContent := renderers.RenderRelationTypePicker(
		renderers.GetRelationTypeOptions(),
		w.RelationTypePickerState.SelectedRelationTypeID(),
		w.RelationTypePickerState.Cursor(),
		w.UiState.Width()*3/4-8,
		pickerType,
	)

	// Wrap in styled container (reuse LabelPickerBoxStyle)
	pickerBox := components.LabelPickerBoxStyle.
		Width(w.UiState.Width() * 3 / 4).
		Height(w.UiState.Height() * 3 / 4).
		Render(pickerContent)

	return lipgloss.Place(
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}

// ViewStatusPicker renders the status/column selection picker.
func (w *Wrapper) ViewStatusPicker() string {
	var items []string
	columns := w.StatusPickerState.Columns()
	cursor := w.StatusPickerState.Cursor()

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
		w.UiState.Width(), w.UiState.Height(),
		lipgloss.Center, lipgloss.Center,
		pickerBox,
	)
}
