// Package daemon provides CLI commands for managing the paso event daemon.
//
// The daemon enables real-time sync between multiple paso sessions
// (CLI and TUI) by broadcasting events over a Unix domain socket.
package daemon

import (
	"fmt"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

var supportedOS = []string{"linux"}

// checkOS verifies the current OS is in the supported list.
// The goos parameter allows testing without relying on runtime.GOOS directly.
func checkOS(goos string) error {
	if slices.Contains(supportedOS, goos) {
		return nil
	}

	return cli.NewExitErr(cli.ExitError, fmt.Sprintf(
		"\"paso daemon\" requires a supported operating system (detected: %s)\n\n"+
			"Supported: %s\n"+
			"The daemon relies on systemd and Unix domain sockets which are only available on Linux.",
		goos, strings.Join(supportedOS, ", "),
	))
}

// DaemonCmd returns the daemon parent command
func DaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the paso event daemon",
		Long: `Manage the paso event daemon for real-time sync between terminal sessions.

The daemon runs as a background service and broadcasts events (task changes,
project updates, etc.) to all connected paso instances over a Unix socket.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return checkOS(runtime.GOOS)
		},
	}

	cmd.AddCommand(StartCmd())
	cmd.AddCommand(SetupCmd())
	cmd.AddCommand(StatusCmd())
	cmd.AddCommand(StopCmd())

	return cmd
}
