package daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thenoetrevino/paso/internal/cli"
)

func TestServiceTemplate_ContainsRequiredDirectives(t *testing.T) {
	t.Parallel()

	directives := []string{
		"[Unit]",
		"[Service]",
		"[Install]",
		"ExecStart=%s daemon start",
		"Restart=on-failure",
		"WantedBy=default.target",
		"ReadWritePaths=%s",
		"PrivateTmp=yes",
		"NoNewPrivileges=yes",
		"ProtectSystem=strict",
		"StandardOutput=journal",
		"StandardError=journal",
	}

	for _, directive := range directives {
		assert.Contains(t, serviceTemplate, directive)
	}
}

func TestInstallDaemonService_NoSystemctl(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := installDaemonService(context.Background())
	require.Error(t, err)

	var exitErr *cli.ExitErr
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, cli.ExitError, exitErr.Code)
	assert.Contains(t, exitErr.Message, "systemctl not found")
}

func TestRemoveDaemonService_NoSystemctl(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := removeDaemonService(context.Background())
	require.Error(t, err)

	var exitErr *cli.ExitErr
	require.True(t, errors.As(err, &exitErr))
	assert.Equal(t, cli.ExitError, exitErr.Code)
	assert.Contains(t, exitErr.Message, "systemctl not found")
}
