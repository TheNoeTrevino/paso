package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ClaudeCmd returns the setup claude subcommand
func ClaudeCmd() *cobra.Command {
	var projectFlag bool
	var checkFlag bool
	var removeFlag bool

	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Setup Claude Code integration",
		Long: `Install hooks to automatically run 'paso prime' on SessionStart and PreCompact.

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
		Run: func(cmd *cobra.Command, args []string) {
			if checkFlag {
				CheckClaude()
				return
			}

			if removeFlag {
				RemoveClaude(projectFlag)
				return
			}

			InstallClaude(projectFlag)
		},
	}

	cmd.Flags().BoolVar(&projectFlag, "project", false, "Install for current project only")
	cmd.Flags().BoolVar(&checkFlag, "check", false, "Check installation status")
	cmd.Flags().BoolVar(&removeFlag, "remove", false, "Remove hooks")

	return cmd
}

// InstallClaude installs Claude Code hooks
func InstallClaude(project bool) {
	var settingsPath string

	if project {
		settingsPath = ".claude/settings.local.json"
		fmt.Println("Installing Claude hooks for this project...")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		settingsPath = filepath.Join(home, ".claude/settings.json")
		fmt.Println("Installing Claude hooks globally...")
	}

	// Ensure parent directory exists
	if err := EnsureDir(filepath.Dir(settingsPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load or create settings
	var settings map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		settings = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to parse settings.json: %v\n", err)
			os.Exit(1)
		}
	}

	// Get or create hooks section
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
		settings["hooks"] = hooks
	}

	// Add SessionStart hook
	if addHookCommand(hooks, "SessionStart", "paso prime") {
		fmt.Println("✓ Registered SessionStart hook")
	}

	// Add PreCompact hook
	if addHookCommand(hooks, "PreCompact", "paso prime") {
		fmt.Println("✓ Registered PreCompact hook")
	}

	// Write back to file
	data, err = json.MarshalIndent(settings, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshal settings: %v\n", err)
		os.Exit(1)
	}

	if err := atomicWriteFile(settingsPath, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: write settings: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Claude Code integration installed\n")
	fmt.Printf("  Settings: %s\n", settingsPath)
	fmt.Println("\nRestart Claude Code for changes to take effect.")
}

// CheckClaude checks if Claude integration is installed
func CheckClaude() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	globalSettings := filepath.Join(home, ".claude/settings.json")
	projectSettings := ".claude/settings.local.json"

	globalHooks := hasPasoHooks(globalSettings)
	projectHooks := hasPasoHooks(projectSettings)

	if globalHooks {
		fmt.Println("✓ Global hooks installed:", globalSettings)
	} else if projectHooks {
		fmt.Println("✓ Project hooks installed:", projectSettings)
	} else {
		fmt.Println("✗ No hooks installed")
		fmt.Println("  Run: paso setup claude")
		os.Exit(1)
	}
}

// RemoveClaude removes Claude Code hooks
func RemoveClaude(project bool) {
	var settingsPath string

	if project {
		settingsPath = ".claude/settings.local.json"
		fmt.Println("Removing Claude hooks from project...")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		settingsPath = filepath.Join(home, ".claude/settings.json")
		fmt.Println("Removing Claude hooks globally...")
	}

	// Load settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		fmt.Println("No settings file found")
		return
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse settings.json: %v\n", err)
		os.Exit(1)
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		fmt.Println("No hooks found")
		return
	}

	// Remove paso prime hooks
	removeHookCommand(hooks, "SessionStart", "paso prime")
	removeHookCommand(hooks, "PreCompact", "paso prime")

	// Write back
	data, err = json.MarshalIndent(settings, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshal settings: %v\n", err)
		os.Exit(1)
	}

	if err := atomicWriteFile(settingsPath, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: write settings: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Claude hooks removed")
}

// addHookCommand adds a hook command to an event if not already present
// Returns true if hook was added, false if already exists
func addHookCommand(hooks map[string]interface{}, event, command string) bool {
	// Get or create event array
	eventHooks, ok := hooks[event].([]interface{})
	if !ok {
		eventHooks = []interface{}{}
	}

	// Check if paso hook already registered
	for _, hook := range eventHooks {
		hookMap, ok := hook.(map[string]interface{})
		if !ok {
			continue
		}
		commands, ok := hookMap["hooks"].([]interface{})
		if !ok {
			continue
		}
		for _, cmd := range commands {
			cmdMap, ok := cmd.(map[string]interface{})
			if !ok {
				continue
			}
			if cmdMap["command"] == command {
				fmt.Printf("✓ Hook already registered: %s\n", event)
				return false
			}
		}
	}

	// Add paso hook to array
	newHook := map[string]interface{}{
		"matcher": "",
		"hooks": []interface{}{
			map[string]interface{}{
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
func removeHookCommand(hooks map[string]interface{}, event, command string) {
	eventHooks, ok := hooks[event].([]interface{})
	if !ok {
		return
	}

	// Filter out paso prime hooks
	var filtered []interface{}
	for _, hook := range eventHooks {
		hookMap, ok := hook.(map[string]interface{})
		if !ok {
			filtered = append(filtered, hook)
			continue
		}

		commands, ok := hookMap["hooks"].([]interface{})
		if !ok {
			filtered = append(filtered, hook)
			continue
		}

		keepHook := true
		for _, cmd := range commands {
			cmdMap, ok := cmd.(map[string]interface{})
			if !ok {
				continue
			}
			if cmdMap["command"] == command {
				keepHook = false
				fmt.Printf("✓ Removed %s hook\n", event)
				break
			}
		}

		if keepHook {
			filtered = append(filtered, hook)
		}
	}

	hooks[event] = filtered
}

// hasPasoHooks checks if a settings file has paso prime hooks
func hasPasoHooks(settingsPath string) bool {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return false
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return false
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		return false
	}

	// Check SessionStart and PreCompact for "paso prime"
	for _, event := range []string{"SessionStart", "PreCompact"} {
		eventHooks, ok := hooks[event].([]interface{})
		if !ok {
			continue
		}

		for _, hook := range eventHooks {
			hookMap, ok := hook.(map[string]interface{})
			if !ok {
				continue
			}
			commands, ok := hookMap["hooks"].([]interface{})
			if !ok {
				continue
			}
			for _, cmd := range commands {
				cmdMap, ok := cmd.(map[string]interface{})
				if !ok {
					continue
				}
				if cmdMap["command"] == "paso prime" {
					return true
				}
			}
		}
	}

	return false
}
