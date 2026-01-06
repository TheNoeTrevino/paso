package tui

import "github.com/thenoetrevino/paso/internal/config"

// Message types for database operations
type connectionConfirmation struct {
	connect bool
	config  config.DatabaseConfig
}

// Database update handlers have been organized into logical modules for better maintainability.
//
// Module organization:
//
// - database_selection.go: Database picker UI and selection handling
//   - handleConnectRemote: Show database connection picker
//   - updateDatabaseSelect: Handle keyboard input in database selection mode
//
// - database_creation.go: Database connection form and validation
//   - validateConnectionString: Validate connection string format
//   - updateDatabaseCreate: Handle database creation form input
//
// - database_switching.go: Database connection and switching logic
//   - tickConnectingSpinner: Animate spinner during connection
//   - switchToDatabase: Switch to database by connection string
//   - switchToDatabaseConfig: Switch using DatabaseConfig
//   - handleDisconnectDatabase: Switch back to local SQLite
//   - reloadAllData: Reload all data after database switch
//
// - database_confirmation.go: Connection and deletion confirmations
//   - updateDatabaseConnectConfirm: Handle database connection confirmation
//   - updateDatabaseDeleteConfirm: Handle database deletion confirmation
