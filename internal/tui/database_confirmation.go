package tui

import (
	"fmt"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// updateDatabaseConnectConfirm handles y/n input for database connection confirmation
func (m Model) updateDatabaseConnectConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.DatabasePicker.PendingConnection == nil {
		// Safety: if no pending connection, return to select mode
		m.UIState.Mode = state.DatabaseSelectMode
		return m, nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed - connect to the database
		cfg := *m.DatabasePicker.PendingConnection
		m.DatabasePicker.PendingConnection = nil
		m.UIState.Mode = state.NormalMode
		m.DatabasePicker.StartConnecting(cfg.Name)
		return m, m.switchToDatabaseConfig(cfg)

	case "n", "N", "esc":
		// User cancelled - return to database select mode without connecting
		// Connection is already saved, just don't connect to it
		m.DatabasePicker.PendingConnection = nil
		m.UIState.Mode = state.DatabaseSelectMode
		return m, nil
	}

	return m, nil
}

// updateDatabaseDeleteConfirm handles y/n input for database deletion confirmation
func (m Model) updateDatabaseDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.DatabasePicker.PendingDeleteName == "" {
		// Safety: if no pending delete, return to select mode
		m.UIState.Mode = state.DatabaseSelectMode
		return m, nil
	}

	switch msg.String() {
	case "y", "Y":
		// User confirmed - delete the database
		dbName := m.DatabasePicker.PendingDeleteName

		// Remove from config
		if err := m.Config.RemoveDatabase(dbName); err != nil {
			slog.Error("failed to delete database connection", "name", dbName, "error", err)
			m.DatabasePicker.SetError(fmt.Errorf("failed to delete: %w", err))
		} else {
			slog.Info("deleted database connection", "name", dbName)

			// Reload config to refresh the list
			cfg, err := config.Load()
			if err != nil {
				slog.Error("failed to reload config after delete", "error", err)
			} else {
				m.Config = cfg
				m.DatabasePicker.SavedDatabases = cfg.Databases
			}
		}

		// Clear pending delete and return to select mode
		m.DatabasePicker.PendingDeleteName = ""
		m.UIState.Mode = state.DatabaseSelectMode
		return m, nil

	case "n", "N", "esc":
		// User cancelled - return to select mode without deleting
		m.DatabasePicker.PendingDeleteName = ""
		m.UIState.Mode = state.DatabaseSelectMode
		return m, nil
	}

	return m, nil
}
