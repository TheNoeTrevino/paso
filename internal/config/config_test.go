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

	assert.Equal(t, "q", defaults.General.Quit)
	assert.Equal(t, "a", defaults.Tasks.AddTask)
	assert.Equal(t, " ", defaults.Tasks.ViewTask)
	assert.Equal(t, "j", defaults.Navigation.MoveDown)
	assert.Equal(t, "C", defaults.Kanban.CreateColumn)
	assert.Equal(t, "P", defaults.Projects.CreateProject)
	assert.Equal(t, "ctrl+s", defaults.Forms.SaveForm)
	assert.Equal(t, "ctrl+d", defaults.Pickers.DeleteLabel)
	assert.Equal(t, "/", defaults.General.Search)
	assert.Equal(t, "ctrl+t", defaults.Tasks.EditType)
	assert.Equal(t, "ctrl+n", defaults.Forms.OpenCommentsView)
	assert.Equal(t, "f5", defaults.Forms.RefreshGitData)
	assert.Equal(t, "ctrl+l", defaults.Forms.EditLabels)
	assert.Equal(t, "ctrl+p", defaults.Forms.EditParentTask)
	assert.Equal(t, "ctrl+c", defaults.Forms.EditChildTask)
	assert.Equal(t, "ctrl+r", defaults.Forms.EditPriority)
	assert.Equal(t, "ctrl+a", defaults.Forms.EditAssignee)
	assert.Equal(t, "ctrl+e", defaults.Forms.EditEstimate)
	assert.Equal(t, "ctrl+d", defaults.Forms.EditDueDate)
	assert.Equal(t, "ctrl+/", defaults.Forms.ShowHelp)
}

func TestLoadConfigWithoutFile(t *testing.T) {
	// Set to a temp dir that doesn't have a config
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	require.NoError(t, err)

	// Should return default config
	assert.Equal(t, "q", cfg.KeyMappings.General.Quit)
}

func TestLoadConfigWithFile(t *testing.T) {
	// Create temp dir with config
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Join(tempDir, "paso")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Write custom config with nested structure
	configContent := `key_mappings:
  general:
    quit: "x"
  tasks:
    add_task: "n"
    view_task: "v"
`
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	cfg, err := Load()
	require.NoError(t, err)

	// Should load custom values
	assert.Equal(t, "x", cfg.KeyMappings.General.Quit)
	assert.Equal(t, "n", cfg.KeyMappings.Tasks.AddTask)
	assert.Equal(t, "v", cfg.KeyMappings.Tasks.ViewTask)

	// Unspecified values should use defaults
	assert.Equal(t, "e", cfg.KeyMappings.Tasks.EditTask)
	assert.Equal(t, "j", cfg.KeyMappings.Navigation.MoveDown)
	assert.Equal(t, "C", cfg.KeyMappings.Kanban.CreateColumn)
	assert.Equal(t, "P", cfg.KeyMappings.Projects.CreateProject)
	assert.Equal(t, "ctrl+s", cfg.KeyMappings.Forms.SaveForm)
	assert.Equal(t, "ctrl+d", cfg.KeyMappings.Pickers.DeleteLabel)
}

func TestSaveConfig(t *testing.T) {
	// Create temp dir
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	cfg := &Config{
		KeyMappings: KeyMappings{
			General: GeneralKeys{
				Quit: "x",
			},
			Tasks: TaskKeys{
				AddTask:  "n",
				ViewTask: "v",
			},
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
	assert.Equal(t, "x", cfg2.KeyMappings.General.Quit)
	assert.Equal(t, "n", cfg2.KeyMappings.Tasks.AddTask)
}
