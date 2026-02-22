package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
	internaldaemon "github.com/thenoetrevino/paso/internal/daemon"
)

// StartCmd returns the daemon start subcommand
func StartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the daemon in the foreground",
		Long: `Start the paso event daemon in the foreground.

This is typically invoked by systemd (via 'paso daemon setup') rather than
run directly. The daemon listens on a Unix socket at ~/.paso/paso.sock
and broadcasts events between connected paso sessions.

To run the daemon as a background service, use 'paso daemon setup' to
install it as a systemd user service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon()
		},
	}
}

func runDaemon() error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	pasoDir, err := getPasoDir()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get paso directory: %v", err))
	}

	socketPath := filepath.Join(pasoDir, "paso.sock")

	if err := os.MkdirAll(pasoDir, 0700); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to create .paso directory: %v", err))
	}

	server, err := internaldaemon.NewServer(socketPath)
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to create daemon: %v", err))
	}

	slog.Info("paso daemon starting", "socket_path", socketPath, "pid", os.Getpid())

	if err := server.Start(ctx); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("daemon error: %v", err))
	}

	slog.Info("paso daemon shut down gracefully")
	return nil
}
