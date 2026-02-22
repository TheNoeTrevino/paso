package assignee

import (
	"github.com/spf13/cobra"
)

// AssigneeCmd returns the assignee parent command
func AssigneeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assignee",
		Short: "Manage assignees",
	}

	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(SetCmd())
	cmd.AddCommand(WhoAmICmd())
	cmd.AddCommand(DeleteCmd())
	cmd.AddCommand(PickCmd())

	return cmd
}
