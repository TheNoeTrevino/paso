package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/thenoetrevino/paso/internal/cli"
)

const serviceTemplate = `[Unit]
Description=Paso Daemon - Terminal Kanban Board Real-time Sync

[Service]
Type=simple
ExecStart=%s daemon start
Restart=on-failure
RestartSec=5

# Security hardening
PrivateTmp=yes
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths=%s

# Logging
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
`

// SetupCmd returns the daemon setup subcommand
func SetupCmd() *cobra.Command {
	var removeFlag bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Install the daemon as a systemd user service",
		Long: `Install the paso daemon as a systemd user service.

This creates a systemd service file that runs 'paso daemon start' on login,
enabling automatic real-time sync between terminal sessions.

Examples:
  # Install the systemd service
  paso daemon setup

  # Remove the systemd service
  paso daemon setup --remove`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if removeFlag {
				return removeDaemonService(cmd.Context())
			}
			return installDaemonService(cmd.Context())
		},
	}

	cmd.Flags().BoolVar(&removeFlag, "remove", false, "Remove the systemd service")

	return cmd
}

func requireSystemctl() error {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return cli.NewExitErr(cli.ExitError, "systemctl not found; systemd is required")
	}
	return nil
}

func installDaemonService(ctx context.Context) error {
	if err := requireSystemctl(); err != nil {
		return err
	}

	binaryPath, err := os.Executable()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to determine paso binary path: %v", err))
	}

	// Resolve symlinks to get the real path
	binaryPath, err = filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to resolve binary path: %v", err))
	}

	pasoDir, err := getPasoDir()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get paso directory: %v", err))
	}

	serviceDir, err := getSystemdUserDir()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get systemd user directory: %v", err))
	}

	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to create systemd directory: %v", err))
	}

	serviceFile := filepath.Join(serviceDir, "paso.service")
	serviceContent := fmt.Sprintf(serviceTemplate, binaryPath, pasoDir)

	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to write service file: %v", err))
	}
	cli.PrintSuccess(fmt.Sprintf("Created service file: %s", serviceFile))

	// Reload systemd daemon
	if err := runSystemctl(ctx, "daemon-reload"); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to reload systemd: %v", err))
	}
	cli.PrintSuccess("Reloaded systemd daemon")

	// Enable the service
	if err := runSystemctl(ctx, "enable", "paso.service"); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to enable service: %v", err))
	}
	cli.PrintSuccess("Enabled paso.service (will start on login)")

	// Start the service
	if err := runSystemctl(ctx, "start", "paso.service"); err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to start service: %v", err))
	}
	cli.PrintSuccess("Started paso.service")

	fmt.Println()
	fmt.Println("Useful commands:")
	fmt.Println("  paso daemon status                    View status")
	fmt.Println("  paso daemon stop                      Stop service")
	fmt.Println("  journalctl --user -u paso -f          Follow logs")
	fmt.Println("  systemctl --user restart paso         Restart service")

	return nil
}

func removeDaemonService(ctx context.Context) error {
	if err := requireSystemctl(); err != nil {
		return err
	}

	// Stop the service (ignore errors -- may not be running)
	_ = runSystemctl(ctx, "stop", "paso.service")

	// Disable the service (ignore errors -- may not be enabled)
	_ = runSystemctl(ctx, "disable", "paso.service")

	serviceDir, err := getSystemdUserDir()
	if err != nil {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to get systemd user directory: %v", err))
	}

	serviceFile := filepath.Join(serviceDir, "paso.service")
	if err := os.Remove(serviceFile); err != nil && !os.IsNotExist(err) {
		return cli.NewExitErr(cli.ExitError, fmt.Sprintf("failed to remove service file: %v", err))
	}

	_ = runSystemctl(ctx, "daemon-reload")

	cli.PrintSuccess("Removed paso daemon systemd service")
	return nil
}

func getSystemdUserDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "systemd", "user"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".config", "systemd", "user"), nil
}

func runSystemctl(ctx context.Context, args ...string) error {
	fullArgs := append([]string{"--user"}, args...)
	cmd := exec.CommandContext(ctx, "systemctl", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
