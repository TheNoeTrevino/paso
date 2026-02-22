package standup

import (
	"github.com/spf13/cobra"
)

// StandupCmd returns the standup parent command
func StandupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "standup",
		Short: "Manage standup logs",
		Long: `Track what you've been working on for standup meetings.

Log your progress throughout the day and review it later.

Examples:
  # Log what you did
  paso standup log -m "Fixed auth bug in login flow"

  # Log using your editor
  paso standup log

  # List recent logs
  paso standup list`,
	}

	cmd.AddCommand(LogCmd())
	cmd.AddCommand(ListCmd())
	cmd.AddCommand(DeleteCmd())

	return cmd
}
