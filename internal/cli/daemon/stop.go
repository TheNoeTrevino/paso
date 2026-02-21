package daemon

import (
	"fmt"

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
	if err := requireSystemctl(); err != nil {
		return err
	}

	if err := runSystemctl(cmd.Context(), "stop", "paso.service"); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to stop daemon: %v", err))
	}

	cli.PrintSuccess("Stopped paso daemon")
	return nil
}
