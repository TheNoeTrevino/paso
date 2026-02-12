package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

func TestCommandTree(t *testing.T) {
	expectedTopLevel := []string{
		"task", "project", "column", "label", "assignee",
		"tutorial", "setup", "db", "tui", "completion",
	}

	topCmds := rootCmd.Commands()
	cmdMap := make(map[string]*cobra.Command, len(topCmds))
	for _, cmd := range topCmds {
		cmdMap[cmd.Name()] = cmd
	}

	t.Run("all top-level commands are registered", func(t *testing.T) {
		for _, name := range expectedTopLevel {
			assert.Contains(t, cmdMap, name, "expected top-level command %q to be registered", name)
		}
	})

	t.Run("every command has a non-empty Use field", func(t *testing.T) {
		var walk func(cmd *cobra.Command)
		walk = func(cmd *cobra.Command) {
			assert.NotEmpty(t, cmd.Use, "command %q should have a non-empty Use field", cmd.CommandPath())
			for _, child := range cmd.Commands() {
				walk(child)
			}
		}
		walk(rootCmd)
	})

	subcommandTests := []struct {
		parent   string
		children []string
	}{
		{
			parent: "task",
			children: []string{
				"create", "update", "delete", "move", "done",
				"ready", "to-ready", "show", "list", "comment",
				"assign", "link", "blocked", "in-progress",
			},
		},
		{
			parent:   "project",
			children: []string{"create", "delete", "list", "tree", "git-link", "git-unlink"},
		},
		{
			parent:   "column",
			children: []string{"create", "update", "delete", "list"},
		},
		{
			parent:   "label",
			children: []string{"create", "update", "delete", "list", "attach", "detach"},
		},
		{
			parent:   "assignee",
			children: []string{"create", "list", "set", "whoami", "delete"},
		},
		{
			parent:   "setup",
			children: []string{"claude", "opencode"},
		},
		{
			parent:   "db",
			children: []string{"add", "connect", "list", "remove"},
		},
	}

	for _, tt := range subcommandTests {
		t.Run(tt.parent+" subcommands", func(t *testing.T) {
			parentCmd, ok := cmdMap[tt.parent]
			require.True(t, ok, "parent command %q must exist", tt.parent)

			childMap := make(map[string]bool)
			for _, child := range parentCmd.Commands() {
				childMap[child.Name()] = true
			}

			for _, expected := range tt.children {
				assert.True(t, childMap[expected],
					"expected subcommand %q under %q, got: %v", expected, tt.parent, childMap)
			}
		})
	}

	t.Run("tutorial has no subcommands", func(t *testing.T) {
		tutorialCmd := findCommand(rootCmd, "tutorial")
		require.NotNil(t, tutorialCmd)
		assert.Empty(t, tutorialCmd.Commands(), "tutorial should be a leaf command")
	})

	t.Run("tui has no subcommands", func(t *testing.T) {
		tuiCmd := findCommand(rootCmd, "tui")
		require.NotNil(t, tuiCmd)
		assert.Empty(t, tuiCmd.Commands(), "tui should be a leaf command")
	})

	t.Run("completion has no subcommands", func(t *testing.T) {
		completionCmd := findCommand(rootCmd, "completion")
		require.NotNil(t, completionCmd)
		assert.Empty(t, completionCmd.Commands(), "completion should be a leaf command")
	})
}
