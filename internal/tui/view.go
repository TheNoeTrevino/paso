package tui

import (
	tea "charm.land/bubbletea/v2"
)

// View renders the current state of the application.
// This implements the "View" part of the Model-View-Update pattern.
//
// ARCHITECTURE NOTE (Legacy Stub):
// This method is a stub that exists only for interface compatibility.
// Production code uses: core.App.View() → render.View() (see internal/tui/render/view.go).
//
// All rendering logic has been moved to the render/ package.
// This stub cannot delegate to render.View() due to import cycle restrictions
// (render imports tui for *tui.Model, so tui cannot import render back).
//
// No tests currently call Model.View() directly (verified via grep).
// If you need to add tests that require rendering, use core.App instead.
func (m Model) View() tea.View {
	var view tea.View
	view.AltScreen = true
	view.Content = "Model.View() is deprecated. Use core.App for rendering."
	return view
}
