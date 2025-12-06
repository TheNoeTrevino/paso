package config

import (
	"os"
	"testing"
)

func TestThemeFileLoading(t *testing.T) {
	// Create a temporary theme file
	themeContent := []byte(`theme:
  accent: "#FF0000"
  create: "#00FF00"
  edit: "#0000FF"
`)
	tmpFile, err := os.CreateTemp("", "paso-theme-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(themeContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Set environment variable
	os.Setenv("PASO_THEME_FILE", tmpFile.Name())
	defer os.Unsetenv("PASO_THEME_FILE")

	// Load config
	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify theme was merged
	if config.ColorScheme.Accent != "#FF0000" {
		t.Errorf("Expected accent to be #FF0000, got %s", config.ColorScheme.Accent)
	}
	if config.ColorScheme.Create != "#00FF00" {
		t.Errorf("Expected create to be #00FF00, got %s", config.ColorScheme.Create)
	}
	if config.ColorScheme.Edit != "#0000FF" {
		t.Errorf("Expected edit to be #0000FF, got %s", config.ColorScheme.Edit)
	}

	// Verify other colors still have defaults
	if config.ColorScheme.Delete == "" {
		t.Error("Expected delete to have default value")
	}
}
