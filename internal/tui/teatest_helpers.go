package tui

import (
	"context"
	"database/sql"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/app"
	"github.com/thenoetrevino/paso/internal/config"
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
	t.Cleanup(func() {
		_ = db.Close()
	})

	// Create app container with all services
	appContainer := &app.App{
		TaskService:    task.NewService(db, nil),
		ColumnService:  column.NewService(db, nil),
		LabelService:   label.NewService(db, nil),
		ProjectService: project.NewService(db, nil),
	}

	// Create test project and columns
	ctx := context.Background()
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columns, err := appContainer.ColumnService.GetColumnsByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	// Create initial model
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	m := InitialModel(ctx, appContainer, cfg, nil)

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
		time.Sleep(10 * time.Millisecond)
	}
	return m
}

// SendSpecialKeyToModel sends a special key (arrow, escape, etc.) to the model
func SendSpecialKeyToModel(m *Model, code rune) *Model {
	msg := tea.KeyPressMsg(tea.Key{Code: code})
	updatedModel, _ := m.Update(msg)
	*m = updatedModel.(Model)
	time.Sleep(10 * time.Millisecond)
	return m
}

// TypeStringToModel types a string into a model character by character
func TypeStringToModel(m *Model, s string) *Model {
	for _, r := range s {
		msg := tea.KeyPressMsg(tea.Key{Text: string(r), Code: r})
		updatedModel, _ := m.Update(msg)
		*m = updatedModel.(Model)
		time.Sleep(5 * time.Millisecond)
	}
	return m
}

// WaitForModeChange waits for the model's mode to change to the expected state
func WaitForModeChange(t *testing.T, m *Model, expectedMode state.Mode, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if m.UIState.Mode() == expectedMode {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("Timeout waiting for mode %v (timeout: %v)", expectedMode, timeout)
}
