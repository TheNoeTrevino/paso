// Package setup holds all cli commands related to setting up
// integrations for agentic tools.
//
// This can include hooks, plugins, etc...
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

// ClaudeCmd returns the setup claude subcommand
func ClaudeCmd() *cobra.Command {
	var projectFlag bool
	var checkFlag bool
	var removeFlag bool

	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Setup Claude Code integration",
		Long: `Install hooks to automatically run 'paso tutorial' on SessionStart and PreCompact.

This ensures AI agents always have paso workflow context, even after context compaction.

Examples:
  # Install globally (default)
  paso setup claude

  # Install for current project only
  paso setup claude --project

  # Check installation status
  paso setup claude --check

  # Remove hooks
  paso setup claude --remove
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if checkFlag {
				return CheckClaude()
			}

			if removeFlag {
				return RemoveClaude(projectFlag)
			}

			return InstallClaude(projectFlag)
		},
	}

	cmd.Flags().BoolVar(&projectFlag, "project", false, "Install for current project only")
	cmd.Flags().BoolVar(&checkFlag, "check", false, "Check installation status")
	cmd.Flags().BoolVar(&removeFlag, "remove", false, "Remove hooks")

	return cmd
}

// InstallClaude installs Claude Code hooks
func InstallClaude(project bool) error {
	var settingsPath string

	if project {
		settingsPath = ".claude/settings.local.json"
		fmt.Println("Installing Claude hooks for this project...")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get home directory: %v", err))
		}
		settingsPath = filepath.Join(home, ".claude/settings.json")
		fmt.Println("Installing Claude hooks globally...")
	}

	// Ensure parent directory exists
	if err := EnsureDir(filepath.Dir(settingsPath), 0o755); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("%v", err))
	}

	// Load or create settings
	var settings map[string]any
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		settings = make(map[string]any)
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to parse settings.json: %v", err))
		}
	}

	// Get or create hooks section
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = make(map[string]any)
		settings["hooks"] = hooks
	}

	// Add SessionStart hook
	if addHookCommand(hooks, "SessionStart", "paso tutorial") {
		cli.PrintSuccess("Registered SessionStart hook")
	}

	// Add PreCompact hook
	if addHookCommand(hooks, "PreCompact", "paso tutorial") {
		cli.PrintSuccess("Registered PreCompact hook")
	}

	// Write back to file
	data, err = json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("marshal settings: %v", err))
	}

	if err := atomicWriteFile(settingsPath, data); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("write settings: %v", err))
	}

	fmt.Print(styles.RenderSuccessWithDetails("Claude Code integration installed", []styles.Detail{
		{Key: "Settings", Value: settingsPath},
	}, cli.GetColorScheme()))
	fmt.Println("\nRestart Claude Code for changes to take effect.")
	return nil
}

// CheckClaude checks if Claude integration is installed
func CheckClaude() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get home directory: %v", err))
	}

	globalSettings := filepath.Join(home, ".claude/settings.json")
	projectSettings := ".claude/settings.local.json"

	globalHooks := hasPasoHooks(globalSettings)
	projectHooks := hasPasoHooks(projectSettings)

	if globalHooks {
		cli.PrintSuccess(fmt.Sprintf("Global hooks installed: %s", globalSettings))
	} else if projectHooks {
		cli.PrintSuccess(fmt.Sprintf("Project hooks installed: %s", projectSettings))
	} else {
		fmt.Println("✗ No hooks installed")
		fmt.Println("  Run: paso setup claude")
		return cli.NewExitErr(cli.ExitError, "no hooks installed")
	}
	return nil
}

// RemoveClaude removes Claude Code hooks
func RemoveClaude(project bool) error {
	var settingsPath string

	if project {
		settingsPath = ".claude/settings.local.json"
		fmt.Println("Removing Claude hooks from project...")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get home directory: %v", err))
		}
		settingsPath = filepath.Join(home, ".claude/settings.json")
		fmt.Println("Removing Claude hooks globally...")
	}

	// Load settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		fmt.Println("No settings file found")
		return nil
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to parse settings.json: %v", err))
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		fmt.Println("No hooks found")
		return nil
	}

	// Remove paso tutorial hooks
	removeHookCommand(hooks, "SessionStart", "paso tutorial")
	removeHookCommand(hooks, "PreCompact", "paso tutorial")

	// Write back
	data, err = json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("marshal settings: %v", err))
	}

	if err := atomicWriteFile(settingsPath, data); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("write settings: %v", err))
	}

	cli.PrintSuccess("Claude hooks removed")
	return nil
}

// addHookCommand adds a hook command to an event if not already present
// Returns true if hook was added, false if already exists
func addHookCommand(hooks map[string]any, event, command string) bool {
	// Get or create event array
	eventHooks, ok := hooks[event].([]any)
	if !ok {
		eventHooks = []any{}
	}

	// Check if paso hook already registered
	for _, hook := range eventHooks {
		hookMap, ok := hook.(map[string]any)
		if !ok {
			continue
		}
		commands, ok := hookMap["hooks"].([]any)
		if !ok {
			continue
		}
		for _, cmd := range commands {
			cmdMap, ok := cmd.(map[string]any)
			if !ok {
				continue
			}
			if cmdMap["command"] == command {
				cli.PrintSuccess(fmt.Sprintf("Hook already registered: %s", event))
				return false
			}
		}
	}

	// Add paso hook to array
	newHook := map[string]any{
		"matcher": "",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": command,
			},
		},
	}

	eventHooks = append(eventHooks, newHook)
	hooks[event] = eventHooks
	return true
}

// removeHookCommand removes a hook command from an event
func removeHookCommand(hooks map[string]any, event, command string) {
	eventHooks, ok := hooks[event].([]any)
	if !ok {
		return
	}

	// Filter out paso tutorial hooks
	var filtered []any
	for _, hook := range eventHooks {
		hookMap, ok := hook.(map[string]any)
		if !ok {
			filtered = append(filtered, hook)
			continue
		}

		commands, ok := hookMap["hooks"].([]any)
		if !ok {
			filtered = append(filtered, hook)
			continue
		}

		keepHook := true
		for _, cmd := range commands {
			cmdMap, ok := cmd.(map[string]any)
			if !ok {
				continue
			}
			if cmdMap["command"] == command {
				keepHook = false
				cli.PrintSuccess(fmt.Sprintf("Removed %s hook", event))
				break
			}
		}

		if keepHook {
			filtered = append(filtered, hook)
		}
	}

	hooks[event] = filtered
}

// hasPasoHooks checks if a settings file has paso tutorial hooks
func hasPasoHooks(settingsPath string) bool {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return false
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return false
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		return false
	}

	// Check SessionStart and PreCompact for "paso tutorial"
	for _, event := range []string{"SessionStart", "PreCompact"} {
		eventHooks, ok := hooks[event].([]any)
		if !ok {
			continue
		}

		for _, hook := range eventHooks {
			hookMap, ok := hook.(map[string]any)
			if !ok {
				continue
			}
			commands, ok := hookMap["hooks"].([]any)
			if !ok {
				continue
			}
			for _, cmd := range commands {
				cmdMap, ok := cmd.(map[string]any)
				if !ok {
					continue
				}
				if cmdMap["command"] == "paso tutorial" {
					return true
				}
			}
		}
	}

	return false
}
