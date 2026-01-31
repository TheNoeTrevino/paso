package db

import "github.com/spf13/cobra"

func DbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage database connections",
	}

	cmd.AddCommand(AddCmd())
	cmd.AddCommand(ConnectCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(RemoveCmd())

	return cmd
}
