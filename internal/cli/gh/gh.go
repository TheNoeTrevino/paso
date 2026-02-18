// Package gh provides CLI commands for GitHub integration.
package gh

import (
	"github.com/spf13/cobra"
)

// GhCmd returns the gh parent command.
func GhCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh",
		Short: "GitHub integration commands",
	}

	cmd.AddCommand(ImportCmd())

	return cmd
}
