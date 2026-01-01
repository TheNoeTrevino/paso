package tui

import (
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thenoetrevino/paso/internal/config"
	"github.com/thenoetrevino/paso/internal/testutil"
)

// TestModelInitialization tests that a Model can be created without panic
func TestModelInitialization(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, app := testutil.SetupCLITest(t)
	_ = testutil.CreateTestProject(t, db, "Test Project")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	// Create model with InitialModel which loads data from database
	model := InitialModel(ctx, app, cfg, nil)

	if model.App == nil {
		t.Fatal("Model should have app initialized")
	}
}

// TestModelImplementsTeaModel verifies Model implements tea.Model interface
func TestModelImplementsTeaModel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, app := testutil.SetupCLITest(t)
	_ = testutil.CreateTestProject(t, db, "Test Project")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	model := InitialModel(ctx, app, cfg, nil)

	// Verify implements tea.Model
	var _ tea.Model = model

	// Verify no panic on Update
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Verify View generates output
	view := model.View()
	if view.Content == "" {
		t.Error("View should generate non-empty output")
	}
}

// TestModelViewGeneratesOutput verifies model rendering produces output
func TestModelViewGeneratesOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, app := testutil.SetupCLITest(t)
	projectID := testutil.CreateTestProject(t, db, "Test Project")
	_ = testutil.CreateTestColumn(t, db, projectID, "Todo")

	cfg := &config.Config{
		ColorScheme: config.DefaultColorScheme(),
		KeyMappings: config.DefaultKeyMappings(),
	}

	model := InitialModel(ctx, app, cfg, nil)

	// Verify rendering with different window sizes
	sizes := []struct {
		width  int
		height int
	}{
		{80, 24},
		{120, 30},
		{60, 20},
	}

	for _, size := range sizes {
		model.Update(tea.WindowSizeMsg{Width: size.width, Height: size.height})
		view := model.View()
		if view.Content == "" {
			t.Errorf("View should generate output for size %dx%d", size.width, size.height)
		}
	}
}
