package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// validateConnectionString validates a connection string matches its declared type
func validateConnectionString(connStr string, expectedType string) error {
	switch expectedType {
	case "postgres":
		if !strings.HasPrefix(connStr, "postgres://") &&
			!strings.HasPrefix(connStr, "postgresql://") {
			return fmt.Errorf("PostgreSQL connection string must start with postgres:// or postgresql://")
		}
	case "sqlite":
		if strings.Contains(connStr, "://") {
			return fmt.Errorf("SQLite connection string should be a file path, not a URL")
		}
	}
	return nil
}

// updateDatabaseCreate handles the huh form while in DatabaseCreateMode
func (m Model) updateDatabaseCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If form doesn't exist, return
	if m.Forms.Form.DatabaseForm == nil {
		return m, nil
	}

	// Check for keyboard escape to cancel form
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			// Clear form and return to database select mode
			m.Forms.Form.ClearDatabaseForm()
			m.UIState.SetMode(state.DatabaseSelectMode)
			return m, nil
		}
	}

	// Let huh form handle the message
	model, cmd := m.Forms.Form.DatabaseForm.Update(msg)
	if form, ok := model.(*huh.Form); ok {
		m.Forms.Form.DatabaseForm = form
	}

	// Check if form is complete (user submitted)
	if m.Forms.Form.DatabaseForm != nil && m.Forms.Form.DatabaseForm.State == huh.StateCompleted {
		// Form was submitted
		// Validate connection string matches type
		if err := validateConnectionString(m.Forms.Form.FormDatabaseConnString, m.Forms.Form.FormDatabaseType); err != nil {
			m.DatabasePicker.SetError(err)
			m.Forms.Form.ClearDatabaseForm()
			m.UIState.SetMode(state.DatabaseSelectMode)
			return m, nil
		}

		// Create DatabaseConfig
		cfg := config.DatabaseConfig{
			Name:             m.Forms.Form.FormDatabaseName,
			ConnectionString: m.Forms.Form.FormDatabaseConnString,
			Type:             m.Forms.Form.FormDatabaseType,
		}

		// Save to config
		if err := m.Config.AddDatabase(cfg.Name, cfg.ConnectionString, cfg.Type); err != nil {
			m.DatabasePicker.SetError(fmt.Errorf("failed to save connection: %w", err))
			m.Forms.Form.ClearDatabaseForm()
			m.UIState.SetMode(state.DatabaseSelectMode)
			return m, nil
		}

		// Refresh SavedDatabases state immediately after saving
		m.DatabasePicker.SavedDatabases = m.Config.Databases

		// Store pending connection and show confirmation dialog
		m.DatabasePicker.PendingConnection = &cfg
		m.Forms.Form.ClearDatabaseForm()
		m.UIState.SetMode(state.DatabaseConnectConfirmMode)

		return m, nil
	}

	// Check for form abort (user pressed No on confirmation)
	if m.Forms.Form.DatabaseForm.State == huh.StateAborted {
		m.Forms.Form.ClearDatabaseForm()
		m.UIState.SetMode(state.DatabaseSelectMode)
		return m, nil
	}

	return m, cmd
}
