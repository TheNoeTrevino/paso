package core

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/events"
	"github.com/thenoetrevino/paso/internal/tui"
)

// App wraps the TUI Model and implements the tea.Model interface.
// This is the single entry point for the Bubble Tea application.
// It delegates all operations to the underlying Model and subpackages.
type App struct {
	model *tui.Model
}

// New creates a new App with an initialized Model.
// This is the constructor that should be used instead of tui.InitialModel.
func New(ctx context.Context, repo database.DataStore, cfg *config.Config, eventClient events.EventPublisher) *App {
	model := tui.InitialModel(ctx, repo, cfg, eventClient)
	return &App{model: &model}
}

// Init initializes the Bubble Tea application.
// Implements tea.Model interface.
func (a *App) Init() tea.Cmd {
	return a.model.Init()
}

// Update handles all messages and updates the model.
// Implements tea.Model interface.
// Currently delegates to Model.Update() - Phase 4 will refactor to use handlers package.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := a.model.Update(msg)
	// Type assertion: Update returns tea.Model but we know it's tui.Model
	if updatedModel, ok := m.(tui.Model); ok {
		a.model = &updatedModel
	}
	return a, cmd
}

// View renders the current state of the application.
// Implements tea.Model interface.
// Currently delegates to Model.View() - Phase 4 will refactor to use render package.
func (a *App) View() tea.View {
	return a.model.View()
}

// GetModel returns the underlying Model.
// This is primarily useful for testing purposes.
func (a *App) GetModel() *tui.Model {
	return a.model
}
