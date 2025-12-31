// Package use holds all cli commands related to setting contextual information
// e.g., paso use ...
package use

import (
	"github.com/spf13/cobra"
)

// UseCmd returns the use parent command
func UseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Manage contextual settings (project, ssh, etc.)",
		Long: `Set and manage contextual information for the current shell session.

The 'use' command allows you to set persistent context that applies to
subsequent commands, eliminating the need to repeatedly specify flags.

Available contexts:
  - project: Set the current project context
  - ssh: (future) Set SSH configuration context

Examples:
  eval $(paso use project 3)       # Use project 3
  eval $(paso use project --clear) # Clear project context
  paso use project --show          # Show current project`,
	}

	cmd.AddCommand(ProjectCmd())

	return cmd
}
