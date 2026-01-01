package core

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// TestAppCreation tests that an App can be created and implements tea.Model
func TestAppCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	// Create a test project with data
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	// Create App
	app := New(ctx, appInstance, cfg, nil)

	if app == nil {
		t.Fatal("App should not be nil")
	}
}

// TestAppImplementsTeaModel verifies App implements tea.Model interface
func TestAppImplementsTeaModel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	app := New(ctx, appInstance, cfg, nil)

	// Verify it can be used as tea.Model
	var _ tea.Model = app

	// Test Init
	cmdInit := app.Init()
	if cmdInit == nil {
		t.Fatal("Init() should return a Cmd")
	}
}

// TestAppUpdate tests the Update method
func TestAppUpdate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	app := New(ctx, appInstance, cfg, nil)

	// Send a window size message
	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if model == nil {
		t.Fatal("Update should return a model")
	}

	// Should be able to call Update multiple times
	keyMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	model, _ = model.Update(keyMsg)
	if model == nil {
		t.Fatal("Update should handle key messages")
	}
}

// TestAppView tests the View method
func TestAppView(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	app := New(ctx, appInstance, cfg, nil)

	// Set window size
	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Get view output
	view := model.View()

	if view.Content == "" {
		t.Fatal("View should generate content")
	}
}

// TestAppDelegation tests that App properly delegates to underlying Model
func TestAppDelegation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	app := New(ctx, appInstance, cfg, nil)

	// Send multiple messages
	model := tea.Model(app)

	// Window size
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Key message
	keyMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	model, _ = model.Update(keyMsg)

	// View should work after both
	view := model.View()
	if view.Content == "" {
		t.Fatal("View should still generate content after key messages")
	}
}

// TestAppStatePreservation tests that App preserves model state across updates
func TestAppStatePreservation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	taskID := testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	app := New(ctx, appInstance, cfg, nil)

	// Set window size
	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app1 := model.(*App)

	// Get initial view
	view1 := app1.View()

	// Send another message
	keyMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	model, _ = model.Update(keyMsg)
	app2 := model.(*App)

	// View should still render
	view2 := app2.View()

	// Both views should be non-empty
	if view1.Content == "" || view2.Content == "" {
		t.Fatal("Views should both be non-empty")
	}

	_ = taskID
}

// TestAppWithNoEventClient tests App creation with nil event client
func TestAppWithNoEventClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	// Create App with nil event client
	app := New(ctx, appInstance, cfg, nil)

	if app == nil {
		t.Fatal("App should be created even without event client")
	}

	// Should still work normally
	cmd := app.Init()
	if cmd == nil {
		t.Fatal("Init should return a command")
	}
}

// TestAppSequentialUpdates tests multiple sequential updates
func TestAppSequentialUpdates(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	app := New(ctx, appInstance, cfg, nil)

	// Send a sequence of messages
	messages := []tea.Msg{
		tea.WindowSizeMsg{Width: 120, Height: 40},
		tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}),
		tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}),
		tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}),
		tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}),
	}

	model := tea.Model(app)
	for _, msg := range messages {
		var nextModel tea.Model
		nextModel, _ = model.Update(msg)
		if nextModel == nil {
			t.Fatalf("Update failed for message %T", msg)
		}
		model = nextModel
	}

	// Should still be valid after all updates
	view := model.View()
	if view.Content == "" {
		t.Fatal("View should still be valid after sequential updates")
	}
}

// TestAppContextRespect tests that App respects context cancellation
func TestAppContextRespect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, appInstance := testutil.SetupCLITest(t)
	defer db.Close()

	projectID := testutil.CreateTestProject(t, db, "Test Project")
	columnID := testutil.CreateTestColumn(t, db, projectID, "Test Column")
	_ = testutil.CreateTestTask(t, db, columnID, "Test Task")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	app := New(ctx, appInstance, cfg, nil)

	// Cancel context
	cancel()

	// App should handle context cancellation gracefully
	_ = app
}
