package tui

import (
	"context"
	"database/sql"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/database"
	"github.com/thenoetrevino/paso/internal/services/column"
	"github.com/thenoetrevino/paso/internal/services/label"
	"github.com/thenoetrevino/paso/internal/services/project"
	"github.com/thenoetrevino/paso/internal/services/task"
	"github.com/thenoetrevino/paso/internal/testutil"
	"github.com/thenoetrevino/paso/internal/tui/state"
)

// SetupTestModelWithDB creates a test model with database and services
// Returns both the model and database for use in tests
func SetupTestModelWithDB(t *testing.T) (Model, *sql.DB) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	// Create app container with all services
	taskSvc, err := task.NewService(db, database.SQLite, nil, nil)
	require.NoError(t, err, "failed to create task service")
	columnSvc, err := column.NewService(db, database.SQLite, nil)
	require.NoError(t, err, "failed to create column service")
	labelSvc, err := label.NewService(db, database.SQLite, nil)
	require.NoError(t, err, "failed to create label service")
	projectSvc, err := project.NewService(db, database.SQLite, nil, nil)
	require.NoError(t, err, "failed to create project service")
	appContainer := &app.App{
		TaskService:    taskSvc,
		ColumnService:  columnSvc,
		LabelService:   labelSvc,
		ProjectService: projectSvc,
	}

	// Create test project and columns
	ctx := context.Background()
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	require.NoError(t, err, "Failed to get columns")

	// Create initial model
	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")
	m := InitialModel(ctx, appContainer, cfg, nil, db)

	// Set up initial state with project data
	m.AppState.SetColumns(columns)
	return m, db
}

// UpdateModelWithMessage updates the model with a message and returns the updated model
func UpdateModelWithMessage(m Model, msg tea.Msg) Model {
	updatedModel, _ := m.Update(msg)
	return updatedModel.(Model)
}

// SendKeysToModel sends multiple key presses to a model sequentially
func SendKeysToModel(m *Model, keys ...tea.Msg) *Model {
	for _, key := range keys {
		updatedModel, _ := m.Update(key)
		*m = updatedModel.(Model)
	}
	return m
}

// SendSpecialKeyToModel sends a special key (arrow, escape, etc.) to the model
func SendSpecialKeyToModel(m *Model, code rune) *Model {
	msg := tea.KeyPressMsg(tea.Key{Code: code})
	updatedModel, _ := m.Update(msg)
	*m = updatedModel.(Model)
	return m
}

// TypeStringToModel types a string into a model character by character
func TypeStringToModel(m *Model, s string) *Model {
	for _, r := range s {
		msg := tea.KeyPressMsg(tea.Key{Text: string(r), Code: r})
		updatedModel, _ := m.Update(msg)
		*m = updatedModel.(Model)
	}
	return m
}

// WaitForModeChange verifies the model's mode matches the expected state
// Note: Bubbletea model updates are synchronous, so this checks immediately
func WaitForModeChange(t *testing.T, m *Model, expectedMode state.Mode, timeout time.Duration) {
	t.Helper()
	require.Equal(t, expectedMode, m.UIState.Mode)
}
