package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThemeFileLoading(t *testing.T) {
	// Create a temporary theme file
	themeContent := []byte(`theme:
  accent: "#FF0000"
  create: "#00FF00"
  edit: "#0000FF"
`)
	tmpFile, err := os.CreateTemp("", "paso-theme-*.yaml")
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	_, err = tmpFile.Write(themeContent)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// Set environment variable
	t.Setenv("PASO_THEME_FILE", tmpFile.Name())

	// Load config
	cfg, err := Load()
	require.NoError(t, err)

	// Verify theme was merged
	assert.Equal(t, "#FF0000", cfg.ColorScheme.Accent)
	assert.Equal(t, "#00FF00", cfg.ColorScheme.Create)
	assert.Equal(t, "#0000FF", cfg.ColorScheme.Edit)

	// Verify other colors still have defaults
	assert.NotEmpty(t, cfg.ColorScheme.Delete)
}
