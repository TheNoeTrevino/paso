package db

import "github.com/spf13/cobra"

func DbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage database connections",
	}

	cmd.AddCommand(ConnectCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(UseCmd())
	cmd.AddCommand(RemoveCmd())

	return cmd
}
