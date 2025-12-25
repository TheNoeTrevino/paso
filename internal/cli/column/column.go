package column

import (
	"github.com/spf13/cobra"
)

// ColumnCmd returns the column parent command
func ColumnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "column",
		Short: "Manage columns",
	}

	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(UpdateCmd())
	cmd.AddCommand(DeleteCmd())

	return cmd
}
