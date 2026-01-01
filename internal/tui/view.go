package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// View renders the current state of the application.
// This implements the "View" part of the Model-View-Update pattern.
func (m Model) View() tea.View {
	var view tea.View
	view.AltScreen = true                                   // Use alternate screen buffer
	view.BackgroundColor = lipgloss.Color(theme.Background) // Set root background color

	// Wait for terminal size to be initialized
	if m.UIState.Width() == 0 {
		view.Content = "Loading..."
		return view
	}

	// Check if current mode uses layer-based rendering
	if m.UIState.Mode().UsesLayers() {
		// Layer-based rendering: always show base board with modal overlays
		baseView := m.viewKanbanBoard()

		// Start layer stack with base view
		layers := []*lipgloss.Layer{
			lipgloss.NewLayer(baseView),
		}

		// Add modal overlay based on mode
		var modalLayer *lipgloss.Layer
		switch m.UIState.Mode() {
		case state.TicketFormMode:
			modalLayer = m.renderTaskFormLayer()
		case state.ProjectFormMode:
			modalLayer = m.renderProjectFormLayer()
		case state.AddColumnFormMode, state.EditColumnFormMode:
			modalLayer = m.renderColumnFormLayer()
		case state.CommentFormMode:
			layers = append(layers, m.renderTaskFormLayer())
			modalLayer = m.renderCommentFormLayer()
		case state.CommentsViewMode:
			layers = append(layers, m.renderTaskFormLayer())
			modalLayer = m.renderCommentsViewLayer()
		case state.HelpMode:
			modalLayer = m.renderHelpLayer()
		case state.DiscardConfirmMode:
			layers = append(layers, m.renderTaskFormLayer())
			modalLayer = m.renderDiscardConfirmLayer()
		case state.TaskFormHelpMode:
			layers = append(layers, m.renderTaskFormLayer())
			modalLayer = m.renderTaskFormHelpLayer()
		case state.LabelPickerMode:
			layers = m.buildPickerLayers(layers, m.LabelPickerState.ReturnMode, m.renderLabelPickerLayer())
		case state.ParentPickerMode:
			layers = m.buildPickerLayers(layers, m.ParentPickerState.ReturnMode, m.renderParentPickerLayer())
		case state.ChildPickerMode:
			layers = m.buildPickerLayers(layers, m.ChildPickerState.ReturnMode, m.renderChildPickerLayer())
		case state.PriorityPickerMode:
			layers = m.buildPickerLayers(layers, m.PriorityPickerState.ReturnMode, m.renderPriorityPickerLayer())
		case state.TypePickerMode:
			layers = m.buildPickerLayers(layers, m.TypePickerState.ReturnMode, m.renderTypePickerLayer())
		case state.RelationTypePickerMode:
			// RelationTypePicker is only accessible from ParentPicker or ChildPicker,
			// so returnMode will always be one of those two modes (see update_pickers.go:329, 484).
			// We render the appropriate intermediate layer to maintain the picker stack.
			var intermediateLayer *lipgloss.Layer
			returnMode := m.RelationTypePickerState.ReturnMode
			switch returnMode {
			case state.ParentPickerMode:
				intermediateLayer = m.renderParentPickerLayer()
			case state.ChildPickerMode:
				intermediateLayer = m.renderChildPickerLayer()
			}
			layers = m.buildPickerLayers(layers, returnMode, m.renderRelationTypePickerLayer(), intermediateLayer)
		case state.StatusPickerMode:
			modalLayer = m.renderStatusPickerLayer()
		}

		if modalLayer != nil {
			layers = append(layers, modalLayer)
		}

		canvas := lipgloss.NewCanvas(layers...)
		view.Content = canvas.Render()
	} else {
		// Legacy full-screen rendering for modes not yet converted to layers
		var content string
		switch m.UIState.Mode() {
		case state.DeleteConfirmMode:
			content = m.viewDeleteTaskConfirm()
		case state.DeleteColumnConfirmMode:
			content = m.viewDeleteColumnConfirm()
		default:
			content = m.viewKanbanBoard()
		}
		view.Content = content
	}

	return view
}

// shouldStackTaskForm determines if the task form should be stacked below a picker.
// The task form is stacked when:
//   - The picker will return directly to TicketFormMode (picker opened from task form)
//   - The picker will return to another picker that was opened from task form
//     (e.g., RelationTypePicker returns to ParentPicker which was opened from task form)
func (m Model) shouldStackTaskForm(returnMode state.Mode) bool {
	return returnMode == state.TicketFormMode ||
		returnMode == state.ParentPickerMode ||
		returnMode == state.ChildPickerMode
}

// buildPickerLayers is a helper method that builds layer stacks for picker modes.
// It handles the common pattern of stacking the task form layer when a picker was
// opened from TicketFormMode, and supports stacking additional intermediate layers.
//
// Parameters:
//   - layers: the base layer stack to append to
//   - returnMode: the mode to return to when the picker closes
//   - pickerLayer: the picker layer to render as the top modal
//   - intermediateLayers: optional layers to stack between task form and picker
//     (used by RelationTypePicker to stack parent/child picker)
//
// Returns the updated layers slice with picker layers appended.
func (m Model) buildPickerLayers(layers []*lipgloss.Layer, returnMode state.Mode, pickerLayer *lipgloss.Layer, intermediateLayers ...*lipgloss.Layer) []*lipgloss.Layer {
	if m.shouldStackTaskForm(returnMode) {
		layers = append(layers, m.renderTaskFormLayer())
		layers = append(layers, intermediateLayers...)
	}

	if pickerLayer != nil {
		layers = append(layers, pickerLayer)
	}

	return layers
}
