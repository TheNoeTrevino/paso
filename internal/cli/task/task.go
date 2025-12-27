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
	cmd.AddCommand(ShowCmd())
	cmd.AddCommand(UpdateCmd())
	cmd.AddCommand(DeleteCmd())
	cmd.AddCommand(LinkCmd())
	cmd.AddCommand(ReadyCmd())
	cmd.AddCommand(BlockedCmd())
	cmd.AddCommand(MoveCmd())
	cmd.AddCommand(ReadyMoveCmd())
	cmd.AddCommand(DoneCmd())

	return cmd
}
