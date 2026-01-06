package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/tui/huhforms"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// handleConnectRemote shows the database connection picker
// Allows switching between any saved databases (local SQLite or remote PostgreSQL)
func (m Model) handleConnectRemote() (tea.Model, tea.Cmd) {
	// Always show the database picker - user can select local to "disconnect" from remote
	m.DatabasePicker.Reset()
	m.DatabasePicker.SavedDatabases = m.Config.Databases
	m.UIState.SetMode(state.DatabaseSelectMode)

	return m, nil
}

// updateDatabaseSelect handles keyboard input while in DatabaseSelectMode
func (m Model) updateDatabaseSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.DatabasePicker.MoveCursorUp()
		return m, nil

	case "down", "j":
		m.DatabasePicker.MoveCursorDown()
		return m, nil

	case "d", "D":
		// Only allow delete on saved databases (not Local or Create New)
		if m.DatabasePicker.IsSelectedSavedDatabase() {
			selectedDB := m.DatabasePicker.SavedDatabases[m.DatabasePicker.Cursor]
			m.DatabasePicker.PendingDeleteName = selectedDB.Name
			m.UIState.SetMode(state.DatabaseDeleteConfirmMode)
			return m, nil
		}
		return m, nil

	case "enter":
		// Handle selection based on cursor position
		if m.DatabasePicker.IsSelectedLocal() {
			// User selected local SQLite - disconnect from remote
			return m.handleDisconnectDatabase()
		}

		if m.DatabasePicker.IsSelectedCreateNew() {
			// User selected create new - transition to DatabaseCreateMode
			m.UIState.SetMode(state.DatabaseCreateMode)

			// Initialize the form with empty fields
			m.Forms.Form.FormDatabaseName = ""
			m.Forms.Form.FormDatabaseConnString = ""
			m.Forms.Form.FormDatabaseType = "postgres"
			m.Forms.Form.FormDatabaseConfirm = true

			// Create the form with theme
			m.Forms.Form.DatabaseForm = huhforms.CreateDatabaseConnectionForm(
				&m.Forms.Form.FormDatabaseName,
				&m.Forms.Form.FormDatabaseConnString,
				&m.Forms.Form.FormDatabaseType,
				&m.Forms.Form.FormDatabaseConfirm,
			).WithTheme(huhforms.CreatePasoTheme(m.Config.ColorScheme))

			// Snapshot initial values for change detection
			m.Forms.Form.SnapshotDatabaseFormInitialValues()

			return m, m.Forms.Form.DatabaseForm.Init()
		}

		if m.DatabasePicker.IsSelectedSavedDatabase() {
			// User selected a saved database - switch to it
			selectedDB := m.DatabasePicker.SavedDatabases[m.DatabasePicker.Cursor]

			// IMMEDIATELY start showing spinner - this must be first for instant UI feedback
			m.DatabasePicker.StartConnecting(selectedDB.Name)

			// Clear any previous errors
			m.DatabasePicker.ClearError()

			// Start connection process
			return m, m.switchToDatabaseConfig(selectedDB)
		}

		return m, nil

	case "esc":
		// Cancel and return to normal mode
		m.UIState.SetMode(state.NormalMode)
		m.DatabasePicker.Reset()
		return m, nil
	}

	return m, nil
}
