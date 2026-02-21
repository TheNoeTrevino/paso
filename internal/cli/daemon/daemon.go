// Package daemon provides CLI commands for managing the paso event daemon.
//
// The daemon enables real-time sync between multiple paso sessions
// (CLI and TUI) by broadcasting events over a Unix domain socket.
package daemon

import (
	"github.com/spf13/cobra"
)

// DaemonCmd returns the daemon parent command
func DaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the paso event daemon",
		Long: `Manage the paso event daemon for real-time sync between terminal sessions.

The daemon runs as a background service and broadcasts events (task changes,
project updates, etc.) to all connected paso instances over a Unix socket.`,
	}

	cmd.AddCommand(StartCmd())
	cmd.AddCommand(SetupCmd())
	cmd.AddCommand(StatusCmd())
	cmd.AddCommand(StopCmd())

	return cmd
}
