package state

import "github.com/thenoetrevino/paso/internal/config"

// DatabasePickerState manages the state of the database connection picker
type DatabasePickerState struct {
	// SavedDatabases holds the list of configured database connections from config
	SavedDatabases []config.DatabaseConfig

	// Cursor is the current selection index in DatabaseSelectMode
	// Indices 0 to len(SavedDatabases) are saved databases
	// Index len(SavedDatabases) is "Local SQLite"
	// Index len(SavedDatabases)+1 is "Create New Connection..."
	Cursor int

	// Create form fields (used in DatabaseCreateMode)
	FormName             string
	FormConnectionString string
	FormType             string // "sqlite" or "postgres"
	FormConfirm          bool

	// PendingConnection stores a connection waiting for user confirmation
	// This is set after successful form submission, before the confirmation dialog
	PendingConnection *config.DatabaseConfig

	// PendingDeleteName stores the name of the database connection pending deletion
	// This is set when user presses 'd' on a connection, before the confirmation dialog
	PendingDeleteName string

	// Err stores any error that occurred during connection attempt or form submission
	Err error

	// Connecting indicates if a connection attempt is in progress
	Connecting bool

	// ConnectingDBName stores the name of the database being connected to
	ConnectingDBName string

	// SpinnerFrame tracks the current animation frame (0 to frameCount-1)
	SpinnerFrame int
}

// NewDatabasePickerState creates a new DatabasePickerState with default values
func NewDatabasePickerState() *DatabasePickerState {
	return &DatabasePickerState{
		SavedDatabases:       []config.DatabaseConfig{},
		Cursor:               0,
		FormName:             "",
		FormConnectionString: "",
		FormType:             "postgres", // Default to postgres in form
		FormConfirm:          true,
		PendingConnection:    nil,
		Err:                  nil,
		Connecting:           false,
		ConnectingDBName:     "",
		SpinnerFrame:         0,
	}
}

// Reset clears the picker state to defaults
func (s *DatabasePickerState) Reset() {
	s.SavedDatabases = []config.DatabaseConfig{}
	s.Cursor = 0
	s.FormName = ""
	s.FormConnectionString = ""
	s.FormType = "postgres"
	s.FormConfirm = true
	s.PendingConnection = nil
	s.PendingDeleteName = ""
	s.Err = nil
	s.Connecting = false
	s.ConnectingDBName = ""
	s.SpinnerFrame = 0
}

// SetError updates the error state
func (s *DatabasePickerState) SetError(err error) {
	s.Err = err
}

// ClearError clears any error state
func (s *DatabasePickerState) ClearError() {
	s.Err = nil
}

// SetConnecting updates the connection status
func (s *DatabasePickerState) SetConnecting(connecting bool) {
	s.Connecting = connecting
}

// StartConnecting initiates the connecting state with the database name
func (s *DatabasePickerState) StartConnecting(dbName string) {
	s.Connecting = true
	s.ConnectingDBName = dbName
	s.SpinnerFrame = 0
}

// StopConnecting clears the connecting state
func (s *DatabasePickerState) StopConnecting() {
	s.Connecting = false
	s.ConnectingDBName = ""
	s.SpinnerFrame = 0
}

// AdvanceSpinnerFrame increments the frame index, wrapping at frameCount
func (s *DatabasePickerState) AdvanceSpinnerFrame() {
	const frameCount = 12 // Match spinnerFrames length
	s.SpinnerFrame = (s.SpinnerFrame + 1) % frameCount
}

// IsConnecting returns true if currently connecting
func (s *DatabasePickerState) IsConnecting() bool {
	return s.Connecting
}

// MoveCursorUp moves the cursor up if possible
func (s *DatabasePickerState) MoveCursorUp() {
	if s.Cursor > 0 {
		s.Cursor--
	}
}

// MoveCursorDown moves the cursor down to the next available option
// Total options: SavedDatabases + LocalSQLite + CreateNew (1 + 1)
func (s *DatabasePickerState) MoveCursorDown() {
	maxCursor := len(s.SavedDatabases) + 2 // +2 for LocalSQLite and CreateNew
	if s.Cursor < maxCursor-1 {
		s.Cursor++
	}
}

// SelectedOption returns which option is currently selected
// Returns "create" for Create New, "local" for Local SQLite, or the database name for saved databases
func (s *DatabasePickerState) SelectedOption() string {
	if s.Cursor < len(s.SavedDatabases) {
		return s.SavedDatabases[s.Cursor].Name
	}

	if s.Cursor == len(s.SavedDatabases) {
		return "local"
	}

	return "create"
}

// IsSelectedLocal returns true if "Local SQLite" is selected
func (s *DatabasePickerState) IsSelectedLocal() bool {
	return s.Cursor == len(s.SavedDatabases)
}

// IsSelectedCreateNew returns true if "Create New Connection" is selected
func (s *DatabasePickerState) IsSelectedCreateNew() bool {
	return s.Cursor == len(s.SavedDatabases)+1
}

// IsSelectedSavedDatabase returns true if a saved database is selected
func (s *DatabasePickerState) IsSelectedSavedDatabase() bool {
	return s.Cursor < len(s.SavedDatabases)
}

// GetSelectedConnectionString returns the connection string for the selected option
// Returns empty string if Local SQLite or Create New is selected
func (s *DatabasePickerState) GetSelectedConnectionString() string {
	if s.IsSelectedLocal() || s.IsSelectedCreateNew() {
		return ""
	}

	if s.IsSelectedSavedDatabase() {
		return s.SavedDatabases[s.Cursor].ConnectionString
	}

	return ""
}

// GetSelectedDatabaseName returns the friendly name for the selected option
func (s *DatabasePickerState) GetSelectedDatabaseName() string {
	if s.IsSelectedLocal() {
		return "Local"
	}

	if s.IsSelectedCreateNew() {
		return "Create New Connection"
	}

	if s.IsSelectedSavedDatabase() {
		return s.SavedDatabases[s.Cursor].Name
	}

	return ""
}
