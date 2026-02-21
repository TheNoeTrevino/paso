package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

// StatusCmd returns the daemon status subcommand
func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if the daemon is running",
		Long: `Check the status of the paso event daemon.

Reports whether the daemon socket is reachable and, if systemd is available,
shows the systemd service status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkDaemonStatus(cmd.Context())
		},
	}
}

func checkDaemonStatus(ctx context.Context) error {
	socketPath, err := getSocketPath()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to determine socket path: %v", err))
	}

	socketReachable := isSocketReachable(ctx, socketPath)

	if socketReachable {
		cli.PrintSuccess(fmt.Sprintf("Daemon is running (socket: %s)", socketPath))
	} else {
		fmt.Printf("Daemon is not running (no socket at %s)\n", socketPath)
	}

	// If systemctl is available, also show the service status
	if _, err := exec.LookPath("systemctl"); err == nil {
		fmt.Println()
		showSystemdStatus(ctx)
	}

	if !socketReachable {
		fmt.Println()
		fmt.Println("To start the daemon:")
		fmt.Println("  paso daemon setup       Install and start as systemd service")
		fmt.Println("  paso daemon start       Run in foreground (for debugging)")
		return cli.NewExitErr(cli.ExitError, "daemon is not running")
	}

	return nil
}

func isSocketReachable(ctx context.Context, socketPath string) bool {
	if _, err := os.Stat(socketPath); err != nil {
		return false
	}

	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func showSystemdStatus(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "status", "paso.service", "--no-pager")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run() // Ignore exit code — systemctl returns non-zero if service is inactive
}
