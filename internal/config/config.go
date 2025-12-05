package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// KeyMappings defines all configurable key bindings
type KeyMappings struct {
	// Tasks
	AddTask        string `yaml:"add_task"`
	EditTask       string `yaml:"edit_task"`
	DeleteTask     string `yaml:"delete_task"`
	MoveTaskLeft   string `yaml:"move_task_left"`
	MoveTaskRight  string `yaml:"move_task_right"`
	MoveTaskUp     string `yaml:"move_task_up"`
	MoveTaskDown   string `yaml:"move_task_down"`
	ViewTask       string `yaml:"view_task"`
	EditLabels     string `yaml:"edit_labels"`

	// Columns
	CreateColumn string `yaml:"create_column"`
	RenameColumn string `yaml:"rename_column"`
	DeleteColumn string `yaml:"delete_column"`

	// Navigation
	PrevColumn          string `yaml:"prev_column"`
	NextColumn          string `yaml:"next_column"`
	PrevTask            string `yaml:"prev_task"`
	NextTask            string `yaml:"next_task"`
	ScrollViewportLeft  string `yaml:"scroll_viewport_left"`
	ScrollViewportRight string `yaml:"scroll_viewport_right"`
	NextProject         string `yaml:"next_project"`
	PrevProject         string `yaml:"prev_project"`

	// Other
	ShowHelp string `yaml:"show_help"`
	Quit     string `yaml:"quit"`
}

// Config represents the application configuration
type Config struct {
	KeyMappings KeyMappings `yaml:"key_mappings"`
}

// DefaultKeyMappings returns the default key mappings
func DefaultKeyMappings() KeyMappings {
	return KeyMappings{
		// Tasks
		AddTask:       "a",
		EditTask:      "e",
		DeleteTask:    "d",
		MoveTaskLeft:  "L",
		MoveTaskRight: "H",
		MoveTaskUp:    "K",
		MoveTaskDown:  "J",
		ViewTask:      " ",
		EditLabels:    "l",

		// Columns
		CreateColumn: "C",
		RenameColumn: "R",
		DeleteColumn: "X",

		// Navigation
		PrevColumn:          "h",
		NextColumn:          "l",
		PrevTask:            "k",
		NextTask:            "j",
		ScrollViewportLeft:  "[",
		ScrollViewportRight: "]",
		NextProject:         "{",
		PrevProject:         "}",

		// Other
		ShowHelp: "?",
		Quit:     "q",
	}
}

// Load loads config from the user's config directory
// Returns default config if file doesn't exist
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		// Return default config if we can't determine config path
		return &Config{KeyMappings: DefaultKeyMappings()}, nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{KeyMappings: DefaultKeyMappings()}, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Fill in any missing values with defaults
	config.applyDefaults()

	return &config, nil
}

// Save saves the config to the user's config directory
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	// Try XDG_CONFIG_HOME first
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "paso", "config.yaml"), nil
	}

	// Fall back to ~/.config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "paso", "config.yaml"), nil
}

// applyDefaults fills in missing key mappings with defaults
func (c *Config) applyDefaults() {
	defaults := DefaultKeyMappings()

	if c.KeyMappings.AddTask == "" {
		c.KeyMappings.AddTask = defaults.AddTask
	}
	if c.KeyMappings.EditTask == "" {
		c.KeyMappings.EditTask = defaults.EditTask
	}
	if c.KeyMappings.DeleteTask == "" {
		c.KeyMappings.DeleteTask = defaults.DeleteTask
	}
	if c.KeyMappings.MoveTaskLeft == "" {
		c.KeyMappings.MoveTaskLeft = defaults.MoveTaskLeft
	}
	if c.KeyMappings.MoveTaskRight == "" {
		c.KeyMappings.MoveTaskRight = defaults.MoveTaskRight
	}
	if c.KeyMappings.MoveTaskUp == "" {
		c.KeyMappings.MoveTaskUp = defaults.MoveTaskUp
	}
	if c.KeyMappings.MoveTaskDown == "" {
		c.KeyMappings.MoveTaskDown = defaults.MoveTaskDown
	}
	if c.KeyMappings.ViewTask == "" {
		c.KeyMappings.ViewTask = defaults.ViewTask
	}
	if c.KeyMappings.EditLabels == "" {
		c.KeyMappings.EditLabels = defaults.EditLabels
	}
	if c.KeyMappings.CreateColumn == "" {
		c.KeyMappings.CreateColumn = defaults.CreateColumn
	}
	if c.KeyMappings.RenameColumn == "" {
		c.KeyMappings.RenameColumn = defaults.RenameColumn
	}
	if c.KeyMappings.DeleteColumn == "" {
		c.KeyMappings.DeleteColumn = defaults.DeleteColumn
	}
	if c.KeyMappings.PrevColumn == "" {
		c.KeyMappings.PrevColumn = defaults.PrevColumn
	}
	if c.KeyMappings.NextColumn == "" {
		c.KeyMappings.NextColumn = defaults.NextColumn
	}
	if c.KeyMappings.PrevTask == "" {
		c.KeyMappings.PrevTask = defaults.PrevTask
	}
	if c.KeyMappings.NextTask == "" {
		c.KeyMappings.NextTask = defaults.NextTask
	}
	if c.KeyMappings.ScrollViewportLeft == "" {
		c.KeyMappings.ScrollViewportLeft = defaults.ScrollViewportLeft
	}
	if c.KeyMappings.ScrollViewportRight == "" {
		c.KeyMappings.ScrollViewportRight = defaults.ScrollViewportRight
	}
	if c.KeyMappings.NextProject == "" {
		c.KeyMappings.NextProject = defaults.NextProject
	}
	if c.KeyMappings.PrevProject == "" {
		c.KeyMappings.PrevProject = defaults.PrevProject
	}
	if c.KeyMappings.ShowHelp == "" {
		c.KeyMappings.ShowHelp = defaults.ShowHelp
	}
	if c.KeyMappings.Quit == "" {
		c.KeyMappings.Quit = defaults.Quit
	}
}
