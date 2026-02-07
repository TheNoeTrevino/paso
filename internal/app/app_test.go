package app

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func TestNew(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Create app with no options (defaults)
	app, err := New(db)
	require.NoError(t, err, "failed to create app")

	if app == nil {
		t.Fatal("Expected app to be created, got nil")
	}

	if app.TaskService == nil {
		t.Error("Expected TaskService to be initialized")
	}

	if app.ProjectService == nil {
		t.Error("Expected ProjectService to be initialized")
	}

	if app.ColumnService == nil {
		t.Error("Expected ColumnService to be initialized")
	}

	if app.LabelService == nil {
		t.Error("Expected LabelService to be initialized")
	}
}

func TestClose(t *testing.T) {
	db := testutil.SetupTestDB(t)

	app, err := New(db)
	require.NoError(t, err, "failed to create app")

	err = app.Close()
	if err != nil {
		t.Errorf("Expected Close to succeed, got error: %v", err)
	}
}
