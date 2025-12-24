package project

import (
	"github.com/spf13/cobra"
)

// ProjectCmd returns the project parent command
func ProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}

	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(DeleteCmd())

	return cmd
}
