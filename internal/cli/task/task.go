package task

import (
	"github.com/spf13/cobra"
)

// TaskCmd returns the task parent command
func TaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}

	cmd.AddCommand(CreateCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(UpdateCmd())
	cmd.AddCommand(DeleteCmd())
	cmd.AddCommand(LinkCmd())
	cmd.AddCommand(ReadyCmd())
	cmd.AddCommand(BlockedCmd())
	cmd.AddCommand(MoveCmd())

	return cmd
}
