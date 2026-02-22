// Package jira provides CLI commands for Jira integration.
package jira

import (
	"github.com/spf13/cobra"
)

// JiraCmd returns the jira parent command.
func JiraCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jira",
		Short: "Jira integration commands",
	}

	cmd.AddCommand(ImportCmd())

	return cmd
}
