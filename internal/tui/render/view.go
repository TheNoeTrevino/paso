package render

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/thenoetrevino/paso/internal/tui"
	"github.com/thenoetrevino/paso/internal/tui/state"
	"github.com/thenoetrevino/paso/internal/tui/theme"
)

// View is the main view dispatcher that renders the current state of the application.
// This implements the "View" part of the Model-View-Update pattern.
func View(m *tui.Model) tea.View {
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
		m.UiState.Mode() == state.AddColumnMode ||
		m.UiState.Mode() == state.EditColumnMode ||
		m.UiState.Mode() == state.HelpMode ||
		m.UiState.Mode() == state.NormalMode ||
		m.UiState.Mode() == state.SearchMode

	if usesLayers {
		// Layer-based rendering: always show base board with modal overlays
		baseView := ViewKanbanBoard(m)

		// Start layer stack with base view
		layers := []*lipgloss.Layer{
			lipgloss.NewLayer(baseView),
		}

		// Add modal overlay based on mode
		var modalLayer *lipgloss.Layer
		switch m.UiState.Mode() {
		case state.TicketFormMode:
			modalLayer = RenderTicketFormLayer(m)
		case state.ProjectFormMode:
			modalLayer = RenderProjectFormLayer(m)
		case state.AddColumnMode, state.EditColumnMode:
			modalLayer = RenderColumnInputLayer(m)
		case state.HelpMode:
			modalLayer = RenderHelpLayer(m)
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
			content = ViewDiscardConfirm(m)
		case state.DeleteConfirmMode:
			content = ViewDeleteTaskConfirm(m)
		case state.DeleteColumnConfirmMode:
			content = ViewDeleteColumnConfirm(m)
		case state.LabelPickerMode:
			content = ViewLabelPicker(m)
		case state.ParentPickerMode:
			content = ViewParentPicker(m)
		case state.ChildPickerMode:
			content = ViewChildPicker(m)
		case state.PriorityPickerMode:
			content = ViewPriorityPicker(m)
		case state.TypePickerMode:
			content = ViewTypePicker(m)
		case state.RelationTypePickerMode:
			content = ViewRelationTypePicker(m)
		case state.StatusPickerMode:
			content = ViewStatusPicker(m)
		default:
			content = ViewKanbanBoard(m)
		}
		view.Content = content
	}

	return view
}
