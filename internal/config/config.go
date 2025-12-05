package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	KeyMappings KeyMappings `yaml:"key_mappings"`
	ColorScheme ColorScheme `yaml:"theme"`
}

// Load loads config from the user's config directory
// Returns default config if file doesn't exist
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		// Return default config if we can't determine config path
		return &Config{
			KeyMappings: DefaultKeyMappings(),
			ColorScheme: DefaultColorScheme(),
		}, nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{
			KeyMappings: DefaultKeyMappings(),
			ColorScheme: DefaultColorScheme(),
		}, nil
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
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configPath, data, 0o644)
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

// applyDefaults fills in missing configuration with defaults
func (c *Config) applyDefaults() {
	c.KeyMappings.applyDefaults()
	c.ColorScheme.ApplyDefaults()
}
