package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/testutil"
)

func TestNew(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Create app with no options (defaults)
	app, err := New(db)
	require.NoError(t, err, "failed to create app")

	require.NotNil(t, app)

	assert.NotNil(t, app.TaskService)

	assert.NotNil(t, app.ProjectService)

	assert.NotNil(t, app.ColumnService)

	assert.NotNil(t, app.LabelService)
}

func TestClose(t *testing.T) {
	db := testutil.SetupTestDB(t)

	app, err := New(db)
	require.NoError(t, err, "failed to create app")

	err = app.Close()
	assert.NoError(t, err)
}
