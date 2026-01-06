package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/thenoetrevino/paso/internal/config/colors"
	"gopkg.in/yaml.v3"
)

// DatabaseConfig represents a saved database connection
type DatabaseConfig struct {
	Name             string `yaml:"name"`
	ConnectionString string `yaml:"connection_string"`
	Type             string `yaml:"type"` // "sqlite" or "postgres"
}

// Config represents the application configuration
type Config struct {
	KeyMappings KeyMappings        `yaml:"key_mappings"`
	ColorScheme colors.ColorScheme `yaml:"theme"`
	Databases   []DatabaseConfig   `yaml:"databases"`
}

// loadThemeFile loads and merges theme from PASO_THEME_FILE environment variable
func loadThemeFile(config *Config) {
	themeFile := os.Getenv("PASO_THEME_FILE")
	if themeFile == "" {
		return
	}

	if _, err := os.Stat(themeFile); err != nil {
		return
	}

	themeData, err := os.ReadFile(themeFile)
	if err != nil {
		return
	}

	var themeConfig struct {
		Theme colors.ColorScheme `yaml:"theme"`
	}

	if yaml.Unmarshal(themeData, &themeConfig) == nil {
		(&config.ColorScheme).MergeFrom(themeConfig.Theme)
	}
}

// Load loads config from the user's config directory
// Returns default config if file doesn't exist
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		// Return default config if we can't determine config path
		config := &Config{
			KeyMappings: DefaultKeyMappings(),
			ColorScheme: DefaultColorScheme(),
		}
		loadThemeFile(config)
		return config, nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		config := &Config{
			KeyMappings: DefaultKeyMappings(),
			ColorScheme: DefaultColorScheme(),
		}
		loadThemeFile(config)
		return config, nil
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

	// After config is loaded, check file permissions
	if info, err := os.Stat(configPath); err == nil {
		mode := info.Mode()
		// Check if group or other have any permissions (should be 0600)
		if mode.Perm()&0o077 != 0 {
			slog.Warn("config file has permissive permissions",
				"path", configPath,
				"current", fmt.Sprintf("%o", mode.Perm()),
				"recommended", "0600",
				"fix", fmt.Sprintf("chmod 600 %s", configPath))
		}
	}

	// Load theme from PASO_THEME_FILE if set
	loadThemeFile(&config)

	// Migration: auto-detect type for existing connections that don't have it set
	for i := range config.Databases {
		if config.Databases[i].Type == "" {
			if strings.HasPrefix(config.Databases[i].ConnectionString, "postgres") ||
				strings.HasPrefix(config.Databases[i].ConnectionString, "postgresql") {
				config.Databases[i].Type = "postgres"
			} else {
				config.Databases[i].Type = "sqlite"
			}
		}
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
	return os.WriteFile(configPath, data, 0o600)
}

// AddDatabase adds a new database connection to the config and saves it
func (c *Config) AddDatabase(name string, connectionString string, dbType string) error {
	// Check if database with this name already exists
	for i, db := range c.Databases {
		if db.Name == name {
			// Update existing
			c.Databases[i].ConnectionString = connectionString
			c.Databases[i].Type = dbType
			return c.Save()
		}
	}

	// Add new database
	c.Databases = append(c.Databases, DatabaseConfig{
		Name:             name,
		ConnectionString: connectionString,
		Type:             dbType,
	})

	return c.Save()
}

// RemoveDatabase removes a database connection from the config and saves it
func (c *Config) RemoveDatabase(name string) error {
	for i, db := range c.Databases {
		if db.Name == name {
			c.Databases = append(c.Databases[:i], c.Databases[i+1:]...)
			return c.Save()
		}
	}
	return nil // Silently succeed if database not found
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
