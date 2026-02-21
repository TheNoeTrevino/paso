package daemon

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// StopCmd returns the daemon stop subcommand
func StopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the daemon",
		Long:  `Stop the paso event daemon via systemctl.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return stopDaemon(cmd)
		},
	}
}

func stopDaemon(cmd *cobra.Command) error {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return cli.NewExitErr(cli.ExitError, "systemctl not found; cannot stop service without systemd")
	}

	if err := runSystemctl(cmd.Context(), "stop", "paso.service"); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to stop daemon: %v", err))
	}

	cli.PrintSuccess("Stopped paso daemon")
	return nil
}
