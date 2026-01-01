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
	if m.UiState.Width() == 0 {
		view.Content = "Loading..."
		return view
	}

	// Check if current mode uses layer-based rendering
	usesLayers := m.UiState.Mode() == state.TicketFormMode ||
		m.UiState.Mode() == state.ProjectFormMode ||
		m.UiState.Mode() == state.AddColumnFormMode ||
		m.UiState.Mode() == state.EditColumnFormMode ||
		m.UiState.Mode() == state.CommentFormMode ||
		m.UiState.Mode() == state.CommentsViewMode ||
		m.UiState.Mode() == state.HelpMode ||
		m.UiState.Mode() == state.TaskFormHelpMode ||
		m.UiState.Mode() == state.LabelPickerMode ||
		m.UiState.Mode() == state.ParentPickerMode ||
		m.UiState.Mode() == state.ChildPickerMode ||
		m.UiState.Mode() == state.PriorityPickerMode ||
		m.UiState.Mode() == state.TypePickerMode ||
		m.UiState.Mode() == state.RelationTypePickerMode ||
		m.UiState.Mode() == state.StatusPickerMode ||
		m.UiState.Mode() == state.NormalMode ||
		m.UiState.Mode() == state.SearchMode

	if usesLayers {
		// Layer-based rendering: always show base board with modal overlays
		baseView := m.viewKanbanBoard()

		// Start layer stack with base view
		layers := []*lipgloss.Layer{
			lipgloss.NewLayer(baseView),
		}

		// Add modal overlay based on mode
		var modalLayer *lipgloss.Layer
		switch m.UiState.Mode() {
		case state.TicketFormMode:
			modalLayer = m.renderTaskFormLayer()
		case state.ProjectFormMode:
			modalLayer = m.renderProjectFormLayer()
		case state.AddColumnFormMode, state.EditColumnFormMode:
			modalLayer = m.renderColumnFormLayer()
		case state.CommentFormMode:
			// Stack both task form AND comment form
			layers = append(layers, m.renderTaskFormLayer())
			modalLayer = m.renderCommentFormLayer()
		case state.CommentsViewMode:
			// Stack both task form AND comments view
			layers = append(layers, m.renderTaskFormLayer())
			modalLayer = m.renderCommentsViewLayer()
		case state.HelpMode:
			modalLayer = m.renderHelpLayer()
		case state.TaskFormHelpMode:
			// Stack both task form AND help menu
			layers = append(layers, m.renderTaskFormLayer())
			modalLayer = m.renderTaskFormHelpLayer()
		case state.LabelPickerMode:
			// Stack task form if picker was opened from task form
			if m.LabelPickerState.ReturnMode == state.TicketFormMode {
				layers = append(layers, m.renderTaskFormLayer())
			}
			modalLayer = m.renderLabelPickerLayer()
		case state.ParentPickerMode:
			// Stack task form if picker was opened from task form
			if m.ParentPickerState.ReturnMode == state.TicketFormMode {
				layers = append(layers, m.renderTaskFormLayer())
			}
			modalLayer = m.renderParentPickerLayer()
		case state.ChildPickerMode:
			// Stack task form if picker was opened from task form
			if m.ChildPickerState.ReturnMode == state.TicketFormMode {
				layers = append(layers, m.renderTaskFormLayer())
			}
			modalLayer = m.renderChildPickerLayer()
		case state.PriorityPickerMode:
			// Stack task form if picker was opened from task form
			if m.PriorityPickerState.ReturnMode() == state.TicketFormMode {
				layers = append(layers, m.renderTaskFormLayer())
			}
			modalLayer = m.renderPriorityPickerLayer()
		case state.TypePickerMode:
			// Stack task form if picker was opened from task form
			if m.TypePickerState.ReturnMode() == state.TicketFormMode {
				layers = append(layers, m.renderTaskFormLayer())
			}
			modalLayer = m.renderTypePickerLayer()
		case state.RelationTypePickerMode:
			// Stack task form and parent/child picker if opened from them
			if m.RelationTypePickerState.ReturnMode() == state.ParentPickerMode ||
				m.RelationTypePickerState.ReturnMode() == state.ChildPickerMode {
				layers = append(layers, m.renderTaskFormLayer())
				// Add the parent/child picker layer based on return mode
				if m.RelationTypePickerState.ReturnMode() == state.ParentPickerMode {
					layers = append(layers, m.renderParentPickerLayer())
				} else {
					layers = append(layers, m.renderChildPickerLayer())
				}
			}
			modalLayer = m.renderRelationTypePickerLayer()
		case state.StatusPickerMode:
			modalLayer = m.renderStatusPickerLayer()
		}

		if modalLayer != nil {
			layers = append(layers, modalLayer)
		}

		// Notifications are now rendered inline with tabs, no need for floating layers

		// Combine all layers into canvas
		canvas := lipgloss.NewCanvas(layers...)
		view.Content = canvas.Render()
	} else {
		// Legacy full-screen rendering for modes not yet converted to layers
		var content string
		switch m.UiState.Mode() {
		case state.DiscardConfirmMode:
			content = m.viewDiscardConfirm()
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
