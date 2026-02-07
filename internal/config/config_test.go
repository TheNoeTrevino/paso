package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultKeyMappings(t *testing.T) {
	defaults := DefaultKeyMappings()

	// Test a few key bindings
	assert.Equal(t, "q", defaults.Quit)
	assert.Equal(t, "a", defaults.AddTask)
	assert.Equal(t, " ", defaults.ViewTask)
}

func TestLoadConfigWithoutFile(t *testing.T) {
	// Set to a temp dir that doesn't have a config
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	require.NoError(t, err)

	// Should return default config
	assert.Equal(t, "q", cfg.KeyMappings.Quit)
}

func TestLoadConfigWithFile(t *testing.T) {
	// Create temp dir with config
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Join(tempDir, "paso")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Write custom config
	configContent := `key_mappings:
  quit: "x"
  add_task: "n"
  view_task: "v"
`
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	cfg, err := Load()
	require.NoError(t, err)

	// Should load custom values
	assert.Equal(t, "x", cfg.KeyMappings.Quit)
	assert.Equal(t, "n", cfg.KeyMappings.AddTask)
	assert.Equal(t, "v", cfg.KeyMappings.ViewTask)

	// Unspecified values should use defaults
	assert.Equal(t, "e", cfg.KeyMappings.EditTask)
}

func TestSaveConfig(t *testing.T) {
	// Create temp dir
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	cfg := &Config{
		KeyMappings: KeyMappings{
			Quit:     "x",
			AddTask:  "n",
			ViewTask: "v",
		},
	}

	// Apply defaults to fill missing fields
	cfg.applyDefaults()

	// Save config
	require.NoError(t, cfg.Save())

	// Verify file exists
	configPath := filepath.Join(tempDir, "paso", "config.yaml")
	_, err := os.Stat(configPath)
	require.False(t, os.IsNotExist(err))

	// Load it back
	cfg2, err := Load()
	require.NoError(t, err)

	// Verify values match
	assert.Equal(t, "x", cfg2.KeyMappings.Quit)
	assert.Equal(t, "n", cfg2.KeyMappings.AddTask)
}
