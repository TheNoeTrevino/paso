package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	"github.com/thenoetrevino/paso/internal/cli/styles"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if checkFlag {
				return CheckOpenCode()
			}

			if removeFlag {
				return RemoveOpenCode(projectFlag)
			}

			return InstallOpenCode(projectFlag)
		},
	}

	cmd.Flags().BoolVar(&projectFlag, "project", false, "Install for current project only")
	cmd.Flags().BoolVar(&checkFlag, "check", false, "Check installation status")
	cmd.Flags().BoolVar(&removeFlag, "remove", false, "Remove plugin")

	return cmd
}

// InstallOpenCode installs the OpenCode plugin
func InstallOpenCode(project bool) error {
	var configPath string

	if project {
		configPath = "opencode.json"
		fmt.Println("Installing OpenCode plugin for this project...")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get home directory: %v", err))
		}
		configPath = filepath.Join(home, ".config/opencode/opencode.json")
		fmt.Println("Installing OpenCode plugin globally...")
	}

	// Ensure parent directory exists
	if err := EnsureDir(filepath.Dir(configPath), 0o755); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("%v", err))
	}

	// Load or create config
	var config map[string]any
	data, err := os.ReadFile(configPath)
	if err != nil {
		config = make(map[string]any)
	} else {
		if err := json.Unmarshal(data, &config); err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to parse opencode.json: %v", err))
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
			cli.PrintSuccess("Plugin already installed")
			return nil
		}
	}

	// Add plugin
	plugins = append(plugins, pluginName)
	config["plugin"] = plugins

	// Write back
	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("marshal config: %v", err))
	}

	if err := atomicWriteFile(configPath, data); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("write config: %v", err))
	}

	fmt.Print(styles.RenderSuccessWithDetails("OpenCode plugin installed", []styles.Detail{
		{Key: "Config", Value: configPath},
	}, cli.GetColorScheme()))
	fmt.Println("\nNote: You may need to install the plugin package:")
	fmt.Println("  bun add opencode-paso")
	return nil
}

// CheckOpenCode checks if OpenCode plugin is installed
func CheckOpenCode() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get home directory: %v", err))
	}

	globalConfig := filepath.Join(home, ".config/opencode/opencode.json")
	projectConfig := "opencode.json"

	globalInstalled := hasPasoPlugin(globalConfig)
	projectInstalled := hasPasoPlugin(projectConfig)

	if globalInstalled {
		cli.PrintSuccess(fmt.Sprintf("Plugin installed globally: %s", globalConfig))
	} else if projectInstalled {
		cli.PrintSuccess(fmt.Sprintf("Plugin installed for project: %s", projectConfig))
	} else {
		fmt.Println("✗ Plugin not installed")
		fmt.Println("  Run: paso setup opencode")
		return cli.NewExitErr(cli.ExitError, "plugin not installed")
	}
	return nil
}

// RemoveOpenCode removes the OpenCode plugin
func RemoveOpenCode(project bool) error {
	var configPath string

	if project {
		configPath = "opencode.json"
		fmt.Println("Removing OpenCode plugin from project...")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get home directory: %v", err))
		}
		configPath = filepath.Join(home, ".config/opencode/opencode.json")
		fmt.Println("Removing OpenCode plugin globally...")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("No config file found")
		return nil
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to parse opencode.json: %v", err))
	}

	plugins, ok := config["plugin"].([]any)
	if !ok {
		fmt.Println("No plugins configured")
		return nil
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
		return nil
	}

	config["plugin"] = filtered

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("marshal config: %v", err))
	}

	if err := atomicWriteFile(configPath, data); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("write config: %v", err))
	}

	cli.PrintSuccess("OpenCode plugin removed")
	return nil
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
