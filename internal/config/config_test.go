package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultKeyMappings(t *testing.T) {
	defaults := DefaultKeyMappings()

	// Test a few key bindings
	if defaults.Quit != "q" {
		t.Errorf("Default Quit key = %s, want q", defaults.Quit)
	}
	if defaults.AddTask != "a" {
		t.Errorf("Default AddTask key = %s, want a", defaults.AddTask)
	}
	if defaults.ViewTask != " " {
		t.Errorf("Default ViewTask key = %s, want space", defaults.ViewTask)
	}
}

func TestLoadConfigWithoutFile(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Set to a temp dir that doesn't have a config
	tempDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() without config file failed: %v", err)
	}

	// Should return default config
	if cfg.KeyMappings.Quit != "q" {
		t.Errorf("Loaded config Quit key = %s, want q (default)", cfg.KeyMappings.Quit)
	}
}

func TestLoadConfigWithFile(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Create temp dir with config
	tempDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir := filepath.Join(tempDir, "paso")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write custom config
	configContent := `key_mappings:
  quit: "x"
  add_task: "n"
  view_task: "v"
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with config file failed: %v", err)
	}

	// Should load custom values
	if cfg.KeyMappings.Quit != "x" {
		t.Errorf("Loaded Quit key = %s, want x", cfg.KeyMappings.Quit)
	}
	if cfg.KeyMappings.AddTask != "n" {
		t.Errorf("Loaded AddTask key = %s, want n", cfg.KeyMappings.AddTask)
	}
	if cfg.KeyMappings.ViewTask != "v" {
		t.Errorf("Loaded ViewTask key = %s, want v", cfg.KeyMappings.ViewTask)
	}

	// Unspecified values should use defaults
	if cfg.KeyMappings.EditTask != "e" {
		t.Errorf("Loaded EditTask key = %s, want e (default)", cfg.KeyMappings.EditTask)
	}
}

func TestSaveConfig(t *testing.T) {
	// Save original env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Create temp dir
	tempDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tempDir)

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
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tempDir, "paso", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file not created at %s", configPath)
	}

	// Load it back
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() failed: %v", err)
	}

	// Verify values match
	if cfg2.KeyMappings.Quit != "x" {
		t.Errorf("Reloaded Quit key = %s, want x", cfg2.KeyMappings.Quit)
	}
	if cfg2.KeyMappings.AddTask != "n" {
		t.Errorf("Reloaded AddTask key = %s, want n", cfg2.KeyMappings.AddTask)
	}
}
