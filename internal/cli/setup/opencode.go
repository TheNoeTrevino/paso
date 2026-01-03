package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// OpenCodeCmd returns the setup opencode subcommand
func OpenCodeCmd() *cobra.Command {
	var projectFlag bool
	var checkFlag bool
	var removeFlag bool

	cmd := &cobra.Command{
		Use:   "opencode",
		Short: "Setup OpenCode integration",
		Long: `Install the opencode-paso plugin for automatic context injection.

This ensures AI agents always have paso workflow context at session start.

Examples:
  # Install globally (default)
  paso setup opencode

  # Install for current project only
  paso setup opencode --project

  # Check installation status
  paso setup opencode --check

  # Remove plugin
  paso setup opencode --remove
`,
		Run: func(cmd *cobra.Command, args []string) {
			if checkFlag {
				CheckOpenCode()
				return
			}

			if removeFlag {
				RemoveOpenCode(projectFlag)
				return
			}

			InstallOpenCode(projectFlag)
		},
	}

	cmd.Flags().BoolVar(&projectFlag, "project", false, "Install for current project only")
	cmd.Flags().BoolVar(&checkFlag, "check", false, "Check installation status")
	cmd.Flags().BoolVar(&removeFlag, "remove", false, "Remove plugin")

	return cmd
}

// InstallOpenCode installs the OpenCode plugin
func InstallOpenCode(project bool) {
	var configPath string

	if project {
		configPath = "opencode.json"
		fmt.Println("Installing OpenCode plugin for this project...")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		configPath = filepath.Join(home, ".config/opencode/opencode.json")
		fmt.Println("Installing OpenCode plugin globally...")
	}

	// Ensure parent directory exists
	if err := EnsureDir(filepath.Dir(configPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load or create config
	var config map[string]any
	data, err := os.ReadFile(configPath)
	if err != nil {
		config = make(map[string]any)
	} else {
		if err := json.Unmarshal(data, &config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to parse opencode.json: %v\n", err)
			os.Exit(1)
		}
	}

	// Get or create plugin array
	plugins, ok := config["plugin"].([]any)
	if !ok {
		plugins = []any{}
	}

	// Check if already installed
	pluginName := "opencode-paso"
	for _, p := range plugins {
		if p == pluginName {
			fmt.Println("✓ Plugin already installed")
			return
		}
	}

	// Add plugin
	plugins = append(plugins, pluginName)
	config["plugin"] = plugins

	// Write back
	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshal config: %v\n", err)
		os.Exit(1)
	}

	if err := atomicWriteFile(configPath, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: write config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ OpenCode plugin installed\n")
	fmt.Printf("  Config: %s\n", configPath)
	fmt.Println("\nNote: You may need to install the plugin package:")
	fmt.Println("  bun add opencode-paso")
}

// CheckOpenCode checks if OpenCode plugin is installed
func CheckOpenCode() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	globalConfig := filepath.Join(home, ".config/opencode/opencode.json")
	projectConfig := "opencode.json"

	globalInstalled := hasPasoPlugin(globalConfig)
	projectInstalled := hasPasoPlugin(projectConfig)

	if globalInstalled {
		fmt.Println("✓ Plugin installed globally:", globalConfig)
	} else if projectInstalled {
		fmt.Println("✓ Plugin installed for project:", projectConfig)
	} else {
		fmt.Println("✗ Plugin not installed")
		fmt.Println("  Run: paso setup opencode")
		os.Exit(1)
	}
}

// RemoveOpenCode removes the OpenCode plugin
func RemoveOpenCode(project bool) {
	var configPath string

	if project {
		configPath = "opencode.json"
		fmt.Println("Removing OpenCode plugin from project...")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		configPath = filepath.Join(home, ".config/opencode/opencode.json")
		fmt.Println("Removing OpenCode plugin globally...")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("No config file found")
		return
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse opencode.json: %v\n", err)
		os.Exit(1)
	}

	plugins, ok := config["plugin"].([]any)
	if !ok {
		fmt.Println("No plugins configured")
		return
	}

	// Filter out paso plugin
	var filtered []any
	removed := false
	for _, p := range plugins {
		if p == "opencode-paso" {
			removed = true
			continue
		}
		filtered = append(filtered, p)
	}

	if !removed {
		fmt.Println("Plugin not found in config")
		return
	}

	config["plugin"] = filtered

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshal config: %v\n", err)
		os.Exit(1)
	}

	if err := atomicWriteFile(configPath, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: write config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ OpenCode plugin removed")
}

// hasPasoPlugin checks if opencode.json has the paso plugin
func hasPasoPlugin(configPath string) bool {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return false
	}

	plugins, ok := config["plugin"].([]any)
	if !ok {
		return false
	}

	for _, p := range plugins {
		if p == "opencode-paso" {
			return true
		}
	}

	return false
}
